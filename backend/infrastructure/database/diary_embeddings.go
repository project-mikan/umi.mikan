package database

import (
	"context"
	"fmt"
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
		VALUES ($1, $2, $3, $4::vector, $5, NOW(), NOW())
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
			1 - (e.embedding <=> $2::vector) AS similarity
		FROM diary_embeddings e
		JOIN diaries d ON d.id = e.diary_id
		WHERE e.user_id = $1
			AND 1 - (e.embedding <=> $2::vector) >= $3
		ORDER BY e.embedding <=> $2::vector
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
