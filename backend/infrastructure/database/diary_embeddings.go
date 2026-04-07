package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DiaryChunk はembedding生成対象の日記チャンクを表す
type DiaryChunk struct {
	// チャンクのインデックス（0始まり）
	Index int
	// チャンクのテキスト内容（スニペット表示用）
	Content string
	// チャンクの概要（検索結果に表示する短い説明）
	Summary string
	// チャンクのベクトル埋め込み
	Embedding []float32
	// チャンク分割に使用したLLMモデル
	SplitModelVersion string
}

// DiaryEmbeddingSearchResult は意味的検索の結果を表す
type DiaryEmbeddingSearchResult struct {
	DiaryID uuid.UUID
	Date    time.Time
	// 日記全文（キーワード検索フォールバック用）
	Content string
	// マッチしたチャンクの内容（スニペット表示用）
	ChunkContent string
	// マッチしたチャンクの概要（検索結果に表示する短い説明）
	ChunkSummary string
	// 日記内のチャンク総数
	ChunkCount int
	Similarity float64
	// embedding生成に使用したモデル
	EmbeddingModel string
	// チャンク分割に使用したモデル
	ChunkModel string
}

// DiaryEmbeddingStatus は日記のRAGインデックス状態を表す
type DiaryEmbeddingStatus struct {
	Indexed      bool
	ModelVersion string
	// チャンク分割に使用したモデル
	ChunkModelVersion string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	// チャンク総数
	ChunkCount int
	// 各チャンクの概要（chunk_index順）
	ChunkSummaries []string
}

// GetDiaryEmbeddingStatus は指定された日記のRAGインデックス状態を返す
// 全チャンクのsummaryを返す（ベクトル値は返さない）
func GetDiaryEmbeddingStatus(ctx context.Context, db DB, diaryID, userID uuid.UUID) (*DiaryEmbeddingStatus, error) {
	query := `
		SELECT model_version, chunk_model_version, created_at, updated_at, chunk_summary
		FROM diary_embeddings
		WHERE diary_id = $1 AND user_id = $2
		ORDER BY chunk_index ASC
	`

	rows, err := db.QueryContext(ctx, query, diaryID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get diary embedding status: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var (
		modelVersion, chunkModelVersion string
		createdAt, updatedAt            time.Time
		chunkSummaries                  []string
	)
	first := true

	for rows.Next() {
		var chunkSummary string
		var mv, cmv string
		var cat, uat time.Time
		if err := rows.Scan(&mv, &cmv, &cat, &uat, &chunkSummary); err != nil {
			return nil, fmt.Errorf("failed to scan diary embedding status: %w", err)
		}
		if first {
			modelVersion = mv
			chunkModelVersion = cmv
			createdAt = cat
			updatedAt = uat
			first = false
		}
		chunkSummaries = append(chunkSummaries, chunkSummary)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate diary embedding status: %w", err)
	}

	if first {
		// チャンクが1件もない
		return &DiaryEmbeddingStatus{Indexed: false}, nil
	}

	return &DiaryEmbeddingStatus{
		Indexed:           true,
		ModelVersion:      modelVersion,
		ChunkModelVersion: chunkModelVersion,
		CreatedAt:         createdAt,
		UpdatedAt:         updatedAt,
		ChunkCount:        len(chunkSummaries),
		ChunkSummaries:    chunkSummaries,
	}, nil
}

// embeddingToSQL はfloat32スライスをpgvector形式の文字列に変換する
func embeddingToSQL(v []float32) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// UpsertDiaryChunkEmbeddings は日記の全チャンクをトランザクション内でupsertする
// 既存チャンクを削除してから新チャンクを挿入することで常に最新状態に保つ
func UpsertDiaryChunkEmbeddings(ctx context.Context, db *sql.DB, diaryID, userID uuid.UUID, chunks []DiaryChunk, modelVersion string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// 既存チャンクを全削除
	if _, err := tx.ExecContext(ctx, `DELETE FROM diary_embeddings WHERE diary_id = $1`, diaryID); err != nil {
		return fmt.Errorf("failed to delete existing diary chunks: %w", err)
	}

	// 新チャンクをバッチINSERTで一括挿入（N+1回のクエリを1回に削減）
	if len(chunks) > 0 {
		const colCount = 9
		valuePlaceholders := make([]string, 0, len(chunks))
		valueArgs := make([]any, 0, len(chunks)*colCount)
		for i, chunk := range chunks {
			base := i * colCount
			valuePlaceholders = append(valuePlaceholders, fmt.Sprintf(
				"($%d, $%d, $%d, $%d, $%d, $%d, $%d::halfvec, $%d, $%d, NOW(), NOW())",
				base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9,
			))
			valueArgs = append(valueArgs,
				uuid.New(), diaryID, userID, chunk.Index, chunk.Content, chunk.Summary,
				embeddingToSQL(chunk.Embedding), modelVersion, chunk.SplitModelVersion,
			)
		}
		batchInsert := "INSERT INTO diary_embeddings (id, diary_id, user_id, chunk_index, chunk_content, chunk_summary, embedding, model_version, chunk_model_version, created_at, updated_at) VALUES " +
			strings.Join(valuePlaceholders, ", ")
		if _, err := tx.ExecContext(ctx, batchInsert, valueArgs...); err != nil {
			return fmt.Errorf("failed to batch insert diary chunks: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit chunk upsert transaction: %w", err)
	}

	return nil
}

// SearchDiaryEntriesByEmbedding はベクトル類似度で日記を検索する
// 各日記の中で最も類似度の高いチャンクを1件返し、chunk_contentをスニペットとして使用する
// threshold: コサイン類似度の下限（0.0〜1.0）
// limit: 最大取得件数（日記単位）
func SearchDiaryEntriesByEmbedding(ctx context.Context, db DB, userID uuid.UUID, queryEmbedding []float32, limit int, threshold float64) ([]*DiaryEmbeddingSearchResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	// 各日記のチャンクの中で最も類似度の高いものを1件だけ選ぶ（日記単位に集約）
	// chunk_countはその日記の全チャンク数（閾値フィルタ前の総数）を示す
	query := `
		SELECT diary_id, date, content, chunk_content, chunk_summary, chunk_count, similarity, model_version, chunk_model_version
		FROM (
			SELECT
				d.id AS diary_id,
				d.date,
				d.content,
				e.chunk_content,
				e.chunk_summary,
				(SELECT COUNT(*) FROM diary_embeddings WHERE diary_id = d.id) AS chunk_count,
				1 - (e.embedding <=> $2::halfvec) AS similarity,
				e.model_version,
				e.chunk_model_version,
				ROW_NUMBER() OVER (PARTITION BY d.id ORDER BY e.embedding <=> $2::halfvec ASC) AS rn
			FROM diary_embeddings e
			JOIN diaries d ON d.id = e.diary_id
			WHERE e.user_id = $1
				AND 1 - (e.embedding <=> $2::halfvec) >= $3
		) ranked
		WHERE rn = 1
		ORDER BY similarity DESC
		LIMIT $4
	`

	embeddingStr := embeddingToSQL(queryEmbedding)
	rows, err := db.QueryContext(ctx, query, userID, embeddingStr, threshold, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search diary entries by embedding: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var results []*DiaryEmbeddingSearchResult
	for rows.Next() {
		var result DiaryEmbeddingSearchResult
		if err := rows.Scan(&result.DiaryID, &result.Date, &result.Content, &result.ChunkContent, &result.ChunkSummary, &result.ChunkCount, &result.Similarity, &result.EmbeddingModel, &result.ChunkModel); err != nil {
			return nil, fmt.Errorf("failed to scan diary embedding search result: %w", err)
		}
		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate diary embedding search results: %w", err)
	}

	return results, nil
}

// SetHNSWEfSearch はHNSW検索の精度パラメータをセッションローカルに設定する
// pgvectorのANN検索精度を向上させるために使用する（デフォルト40→値を大きくすると再現率が向上）
func SetHNSWEfSearch(ctx context.Context, db DB, efSearch int) error {
	sqlstr := fmt.Sprintf("SET LOCAL hnsw.ef_search = %d", efSearch)
	if _, err := db.ExecContext(ctx, sqlstr); err != nil {
		return fmt.Errorf("failed to set hnsw.ef_search to %d: %w", efSearch, err)
	}
	return nil
}

// InsertSemanticSearchLog は意味的検索ログを記録する（メトリクス集計用）
func InsertSemanticSearchLog(ctx context.Context, db DB, userID uuid.UUID) error {
	log := &SemanticSearchLog{
		ID:        uuid.New(),
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	return log.Insert(ctx, db)
}

// DiaryIDsWithoutEmbeddings はユーザーのembedding未生成日記IDを日付降順で返す
// 手動再生成用のため全日記を対象とする（スケジューラーの DiaryIDsNeedingEmbedding とは異なる）
func DiaryIDsWithoutEmbeddings(ctx context.Context, db DB, userID uuid.UUID) ([]string, error) {
	const sqlstr = `
		SELECT d.id
		FROM diaries d
		WHERE d.user_id = $1
		  AND NOT EXISTS (
		    SELECT 1 FROM diary_embeddings e WHERE e.diary_id = d.id
		  )
		ORDER BY d.date DESC
	`
	return queryStringSlice(ctx, db, sqlstr, userID)
}
