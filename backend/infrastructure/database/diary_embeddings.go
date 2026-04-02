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

// DiaryEmbeddingSearchResult は意味的検索の結果を表す
type DiaryEmbeddingSearchResult struct {
	DiaryID    uuid.UUID
	Date       time.Time
	Content    string
	Similarity float64
}

// DiaryEmbeddingStatus は日記のRAGインデックス状態を表す
type DiaryEmbeddingStatus struct {
	Indexed         bool
	ModelVersion    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	EmbeddingValues []float32
}

// GetDiaryEmbeddingStatus は指定された日記のRAGインデックス状態を返す
func GetDiaryEmbeddingStatus(ctx context.Context, db DB, diaryID, userID uuid.UUID) (*DiaryEmbeddingStatus, error) {
	query := `
		SELECT model_version, created_at, updated_at, embedding::text
		FROM diary_embeddings
		WHERE diary_id = $1 AND user_id = $2
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

// UpsertDiaryEmbedding は日記のベクトル埋め込みをUPSERTで保存する
func UpsertDiaryEmbedding(ctx context.Context, db DB, diaryID, userID uuid.UUID, embedding []float32, modelVersion string) error {
	query := `
		INSERT INTO diary_embeddings (id, diary_id, user_id, embedding, model_version, created_at, updated_at)
		VALUES ($1, $2, $3, $4::halfvec, $5, NOW(), NOW())
		ON CONFLICT (diary_id) DO UPDATE SET
			embedding = EXCLUDED.embedding,
			model_version = EXCLUDED.model_version,
			updated_at = NOW()
	`

	id := uuid.New()
	embeddingStr := embeddingToSQL(embedding)

	_, err := db.ExecContext(ctx, query, id, diaryID, userID, embeddingStr, modelVersion)
	if err != nil {
		return fmt.Errorf("failed to upsert diary embedding: %w", err)
	}

	return nil
}

// SearchDiaryEntriesByEmbedding はベクトル類似度で日記を検索する
// threshold: コサイン類似度の下限（0.0〜1.0）
// limit: 最大取得件数
func SearchDiaryEntriesByEmbedding(ctx context.Context, db DB, userID uuid.UUID, queryEmbedding []float32, limit int, threshold float64) ([]*DiaryEmbeddingSearchResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	query := `
		SELECT
			d.id,
			d.date,
			d.content,
			1 - (e.embedding <=> $2::halfvec) AS similarity
		FROM diary_embeddings e
		JOIN diaries d ON d.id = e.diary_id
		WHERE e.user_id = $1
			AND 1 - (e.embedding <=> $2::halfvec) >= $3
		ORDER BY e.embedding <=> $2::halfvec
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
		if err := rows.Scan(&result.DiaryID, &result.Date, &result.Content, &result.Similarity); err != nil {
			return nil, fmt.Errorf("failed to scan diary embedding search result: %w", err)
		}
		results = append(results, &result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate diary embedding search results: %w", err)
	}

	return results, nil
}
