package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
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
	// チャンクのベクトル埋め込み
	Embedding []float32
}

// DiaryEmbeddingSearchResult は意味的検索の結果を表す
type DiaryEmbeddingSearchResult struct {
	DiaryID uuid.UUID
	Date    time.Time
	// 日記全文（キーワード検索フォールバック用）
	Content string
	// マッチしたチャンクの内容（スニペット表示用）
	ChunkContent string
	Similarity   float64
}

// DiaryEmbeddingStatus は日記のRAGインデックス状態を表す
type DiaryEmbeddingStatus struct {
	Indexed      bool
	ModelVersion string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	// 最初のチャンクのベクトル値（デバッグ用）
	EmbeddingValues []float32
}

// GetDiaryEmbeddingStatus は指定された日記のRAGインデックス状態を返す
// 複数チャンクが存在する場合はchunk_index=0のデータを返す
func GetDiaryEmbeddingStatus(ctx context.Context, db DB, diaryID, userID uuid.UUID) (*DiaryEmbeddingStatus, error) {
	query := `
		SELECT model_version, created_at, updated_at, embedding::text
		FROM diary_embeddings
		WHERE diary_id = $1 AND user_id = $2
		ORDER BY chunk_index ASC
		LIMIT 1
	`

	var modelVersion, embeddingStr string
	var createdAt, updatedAt time.Time
	err := db.QueryRowContext(ctx, query, diaryID, userID).Scan(&modelVersion, &createdAt, &updatedAt, &embeddingStr)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &DiaryEmbeddingStatus{Indexed: false}, nil
		}
		return nil, fmt.Errorf("failed to get diary embedding status: %w", err)
	}

	// pgvectorの文字列表現 "[v1,v2,...]" をfloat32スライスにパース
	embeddingValues := parseEmbeddingString(embeddingStr)

	return &DiaryEmbeddingStatus{
		Indexed:         true,
		ModelVersion:    modelVersion,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		EmbeddingValues: embeddingValues,
	}, nil
}

// parseEmbeddingString はpgvectorの文字列表現をfloat32スライスに変換する
// pgvectorは "[v1,v2,...,vn]" 形式で返す（科学表記も含む）
func parseEmbeddingString(s string) []float32 {
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	values := make([]float32, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		f, err := strconv.ParseFloat(p, 32)
		if err == nil {
			values = append(values, float32(f))
		}
	}
	return values
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
func UpsertDiaryChunkEmbeddings(ctx context.Context, db DB, diaryID, userID uuid.UUID, chunks []DiaryChunk, modelVersion string) error {
	sqlDB, ok := db.(*sql.DB)
	if !ok {
		return fmt.Errorf("db must be *sql.DB for transactional chunk upsert")
	}

	tx, err := sqlDB.BeginTx(ctx, nil)
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

	// 新チャンクを挿入
	insertQuery := `
		INSERT INTO diary_embeddings
			(id, diary_id, user_id, chunk_index, chunk_content, embedding, model_version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::halfvec, $7, NOW(), NOW())
	`
	for _, chunk := range chunks {
		embeddingStr := embeddingToSQL(chunk.Embedding)
		if _, err := tx.ExecContext(ctx, insertQuery,
			uuid.New(), diaryID, userID, chunk.Index, chunk.Content, embeddingStr, modelVersion,
		); err != nil {
			return fmt.Errorf("failed to insert diary chunk (index=%d): %w", chunk.Index, err)
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
	query := `
		SELECT diary_id, date, content, chunk_content, similarity
		FROM (
			SELECT
				d.id AS diary_id,
				d.date,
				d.content,
				e.chunk_content,
				1 - (e.embedding <=> $2::halfvec) AS similarity,
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
		if err := rows.Scan(&result.DiaryID, &result.Date, &result.Content, &result.ChunkContent, &result.Similarity); err != nil {
			return nil, fmt.Errorf("failed to scan diary embedding search result: %w", err)
		}
		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate diary embedding search results: %w", err)
	}

	return results, nil
}
