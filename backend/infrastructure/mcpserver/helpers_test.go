package mcpserver

import (
	"context"
	"testing"

	"github.com/google/uuid"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
)

// testUUID はバリデーションのみを検証するテスト用にランダムなユーザーIDを返す
func testUUID(t *testing.T) uuid.UUID {
	t.Helper()
	return uuid.New()
}

// createDiaryReq はテスト用の日記作成リクエストを組み立てる
func createDiaryReq(year, month, day uint32, content string) *g.CreateDiaryEntryRequest {
	return &g.CreateDiaryEntryRequest{
		Content: content,
		Date:    &g.YMD{Year: year, Month: month, Day: day},
	}
}

// testEmbeddingDimension は diary_embeddings.embedding (halfvec) の次元数
const testEmbeddingDimension = 3072

// mockEmbedder はテスト用のGeminiEmbedderモック（固定のベクトルを返す）
type mockEmbedder struct{}

func (m *mockEmbedder) GenerateEmbedding(_ context.Context, _ string, _ bool) ([]float32, error) {
	vec := make([]float32, testEmbeddingDimension)
	vec[0] = 1
	return vec, nil
}

func (m *mockEmbedder) Close() error { return nil }

// mockLLMFactory はテスト用のdiary.LLMFactoryモック
type mockLLMFactory struct{}

func (f *mockLLMFactory) CreateGeminiClient(_ context.Context, _ string) (diary.GeminiEmbedder, error) {
	return &mockEmbedder{}, nil
}
