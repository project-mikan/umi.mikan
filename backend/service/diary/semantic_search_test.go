package diary

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

// testEmbeddingDimension は diary_embeddings.embedding (halfvec) の次元数
const testEmbeddingDimension = 3072

// makeTestUnitVector は先頭要素のみ1の単位ベクトルを返す（自身とのコサイン類似度が1になる）
func makeTestUnitVector() []float32 {
	vec := make([]float32, testEmbeddingDimension)
	vec[0] = 1
	return vec
}

// vectorToSQLString はfloat32スライスをpgvectorのリテラル表記に変換する
func vectorToSQLString(vec []float32) string {
	var sb strings.Builder
	sb.WriteString("[")
	for i, v := range vec {
		if i > 0 {
			sb.WriteString(",")
		}
		if v == 0 {
			sb.WriteString("0")
		} else {
			sb.WriteString("1")
		}
	}
	sb.WriteString("]")
	return sb.String()
}

// insertTestDiaryWithEmbedding は日記とそのembeddingを直接挿入する
func insertTestDiaryWithEmbedding(t *testing.T, db *sql.DB, userID uuid.UUID, content, date, chunkSummary string, vec []float32) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	diaryID := uuid.New()
	nowMillis := int64(1700000000000)
	_, err := db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		diaryID, userID, content, date, nowMillis, nowMillis,
	)
	if err != nil {
		t.Fatalf("日記の挿入に失敗: %v", err)
	}

	_, err = db.ExecContext(ctx,
		`INSERT INTO diary_embeddings (id, diary_id, user_id, chunk_index, chunk_content, chunk_summary, embedding, model_version, chunk_model_version, created_at, updated_at)
		 VALUES ($1, $2, $3, 0, $4, $5, $6::halfvec, 'gemini-embedding-001', 'gemini-2.0-flash', NOW(), NOW())`,
		uuid.New(), diaryID, userID, content, chunkSummary, vectorToSQLString(vec),
	)
	if err != nil {
		t.Fatalf("diary_embeddingsの挿入に失敗: %v", err)
	}
	return diaryID
}

func TestDiaryEntry_SearchDiaryEntriesSemanticByUserID_Success(t *testing.T) {
	db := setupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "semantic-success@example.com", "SemSuccessUser")
	testutil.CreateTestUserLLMWithSettings(t, db, userID, "test-api-key", false, false, true)

	// ベクトル検索でヒットする日記（キーワードは含まない）
	unitVec := makeTestUnitVector()
	vectorDiaryID := insertTestDiaryWithEmbedding(t, db, userID, "海辺を散歩してのんびり過ごした", "2024-03-01", "海辺の散歩", unitVec)

	// キーワード検索のみでヒットする日記（embeddingなし、クエリ文字列を含む）
	ctx := createAuthenticatedContext(userID)
	svc := &DiaryEntry{
		DB:         db,
		LLMFactory: &mockLLMFactory{embedder: &mockGeminiEmbedder{returnVec: unitVec}},
	}
	keywordResp, err := svc.CreateDiaryEntry(ctx, &g.CreateDiaryEntryRequest{
		Content: "旅行の計画を立てた",
		Date:    &g.YMD{Year: 2024, Month: 3, Day: 5},
	})
	if err != nil {
		t.Fatalf("キーワード用日記の作成に失敗: %v", err)
	}

	// limit=100を渡して上限50へのクランプも通す
	outcome, err := svc.SearchDiaryEntriesSemanticByUserID(ctx, userID, "旅行", 100)
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	if len(outcome.Results) != 2 {
		t.Fatalf("ハイブリッド検索の期待件数 2 に対して %d 件取得: %+v", len(outcome.Results), outcome.Results)
	}

	// 1件目はベクトル検索ヒット（類似度1で降順先頭）
	first := outcome.Results[0]
	if first.DiaryID != vectorDiaryID {
		t.Errorf("ベクトル検索ヒットのDiaryID: 期待 %v, 実際 %v", vectorDiaryID, first.DiaryID)
	}
	if first.Similarity < 0.99 {
		t.Errorf("同一ベクトルの類似度が1に近いことを期待したが %v", first.Similarity)
	}
	if first.ChunkSummary != "海辺の散歩" {
		t.Errorf("ChunkSummary: 期待 %q, 実際 %q", "海辺の散歩", first.ChunkSummary)
	}
	if first.Snippet == "" {
		t.Error("ベクトル検索ヒットのSnippetが空")
	}

	// 2件目はキーワード検索による補完ヒット（閾値スコアが付与される）
	second := outcome.Results[1]
	if second.DiaryID.String() != keywordResp.Entry.Id {
		t.Errorf("キーワード補完ヒットのDiaryID: 期待 %v, 実際 %v", keywordResp.Entry.Id, second.DiaryID)
	}
	if second.Snippet != "旅行の計画を立てた" {
		t.Errorf("キーワード補完ヒットのSnippetは日記全文から生成される: 実際 %q", second.Snippet)
	}

	// 使用モデルはベクトル検索の先頭結果から取得される
	if outcome.EmbeddingModel != "gemini-embedding-001" {
		t.Errorf("EmbeddingModel: 期待 %q, 実際 %q", "gemini-embedding-001", outcome.EmbeddingModel)
	}
	if outcome.ChunkModel != "gemini-2.0-flash" {
		t.Errorf("ChunkModel: 期待 %q, 実際 %q", "gemini-2.0-flash", outcome.ChunkModel)
	}
}

func TestDiaryEntry_SearchDiaryEntriesSemantic_Success(t *testing.T) {
	db := setupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "semantic-grpc-success@example.com", "Semantic GRPC User")
	testutil.CreateTestUserLLMWithSettings(t, db, userID, "test-api-key", false, false, true)

	unitVec := makeTestUnitVector()
	diaryID := insertTestDiaryWithEmbedding(t, db, userID, "公園でピクニックをした", "2024-04-10", "ピクニック", unitVec)

	ctx := createAuthenticatedContext(userID)
	svc := &DiaryEntry{
		DB:         db,
		LLMFactory: &mockLLMFactory{embedder: &mockGeminiEmbedder{returnVec: unitVec}},
	}

	resp, err := svc.SearchDiaryEntriesSemantic(ctx, &g.SearchDiaryEntriesSemanticRequest{
		Query: "外で楽しく過ごした日",
		Limit: 5,
	})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	if len(resp.Results) != 1 {
		t.Fatalf("期待件数 1 に対して %d 件取得", len(resp.Results))
	}
	result := resp.Results[0]
	if result.DiaryId != diaryID.String() {
		t.Errorf("DiaryId: 期待 %v, 実際 %v", diaryID, result.DiaryId)
	}
	if result.Date == nil || result.Date.Year != 2024 || result.Date.Month != 4 || result.Date.Day != 10 {
		t.Errorf("Date: 期待 2024-04-10, 実際 %+v", result.Date)
	}
	if result.ChunkSummary != "ピクニック" {
		t.Errorf("ChunkSummary: 期待 %q, 実際 %q", "ピクニック", result.ChunkSummary)
	}
	if resp.EmbeddingModel != "gemini-embedding-001" {
		t.Errorf("EmbeddingModel: 期待 %q, 実際 %q", "gemini-embedding-001", resp.EmbeddingModel)
	}
}

func TestDiaryEntry_SearchDiaryEntriesSemanticByUserID_ErrorBranches(t *testing.T) {
	db := setupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "semantic-error@example.com", "Semantic Error User")
	testutil.CreateTestUserLLMWithSettings(t, db, userID, "test-api-key", false, false, true)
	ctx := createAuthenticatedContext(userID)

	t.Run("異常系: LLMFactory未設定の場合はエラー", func(t *testing.T) {
		svc := &DiaryEntry{DB: db}
		_, err := svc.SearchDiaryEntriesSemanticByUserID(ctx, userID, "クエリ", 10)
		if err == nil {
			t.Fatal("LLMFactory未設定でエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: Geminiクライアント作成失敗の場合はエラー", func(t *testing.T) {
		svc := &DiaryEntry{DB: db, LLMFactory: &mockLLMFactory{err: errors.New("client creation failed")}}
		_, err := svc.SearchDiaryEntriesSemanticByUserID(ctx, userID, "クエリ", 10)
		if err == nil {
			t.Fatal("クライアント作成失敗でエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: embedding生成失敗の場合はエラー", func(t *testing.T) {
		svc := &DiaryEntry{DB: db, LLMFactory: &mockLLMFactory{embedder: &mockGeminiEmbedder{returnErr: errors.New("embedding failed")}}}
		_, err := svc.SearchDiaryEntriesSemanticByUserID(ctx, userID, "クエリ", 10)
		if err == nil {
			t.Fatal("embedding生成失敗でエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: セマンティック検索が無効なユーザーはエラー", func(t *testing.T) {
		disabledUserID := testutil.CreateTestUser(t, db, "semantic-disabled@example.com", "SemDisabledUser")
		testutil.CreateTestUserLLMWithSettings(t, db, disabledUserID, "test-api-key", false, false, false)
		svc := &DiaryEntry{DB: db, LLMFactory: &mockLLMFactory{embedder: &mockGeminiEmbedder{}}}
		_, err := svc.SearchDiaryEntriesSemanticByUserID(ctx, disabledUserID, "クエリ", 10)
		if err == nil {
			t.Fatal("セマンティック検索無効でエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: クエリベクトルの次元が不正な場合はベクトル検索エラー", func(t *testing.T) {
		// embeddingを1件登録し、halfvec(3072)と次元の合わない3次元ベクトルを返すモックで検索する
		insertTestDiaryWithEmbedding(t, db, userID, "次元不一致テスト用の日記", "2024-07-01", "テスト", makeTestUnitVector())
		svc := &DiaryEntry{DB: db, LLMFactory: &mockLLMFactory{embedder: &mockGeminiEmbedder{}}}
		_, err := svc.SearchDiaryEntriesSemanticByUserID(ctx, userID, "クエリ", 10)
		if err == nil {
			t.Fatal("次元不一致でエラーを期待したがnilが返った")
		}
	})
}

func TestDiaryEntry_SearchDiaryEntriesSemantic_AuthErrors(t *testing.T) {
	db := setupTestDB(t)
	svc := &DiaryEntry{DB: db}

	t.Run("異常系: 未認証コンテキストはエラー", func(t *testing.T) {
		_, err := svc.SearchDiaryEntriesSemantic(context.Background(), &g.SearchDiaryEntriesSemanticRequest{Query: "クエリ"})
		if err == nil {
			t.Fatal("未認証でエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: ユーザーIDがUUID形式でない場合はエラー", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.UserIDKey, "not-a-uuid")
		_, err := svc.SearchDiaryEntriesSemantic(ctx, &g.SearchDiaryEntriesSemanticRequest{Query: "クエリ"})
		if err == nil {
			t.Fatal("不正なユーザーIDでエラーを期待したがnilが返った")
		}
	})
}

func TestDiaryEntry_SearchDiaryEntries_AuthErrors(t *testing.T) {
	db := setupTestDB(t)
	svc := &DiaryEntry{DB: db}

	t.Run("異常系: 未認証コンテキストはエラー", func(t *testing.T) {
		_, err := svc.SearchDiaryEntries(context.Background(), &g.SearchDiaryEntriesRequest{Keyword: "旅行"})
		if err == nil {
			t.Fatal("未認証でエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: ユーザーIDがUUID形式でない場合はエラー", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middleware.UserIDKey, "not-a-uuid")
		_, err := svc.SearchDiaryEntries(ctx, &g.SearchDiaryEntriesRequest{Keyword: "旅行"})
		if err == nil {
			t.Fatal("不正なユーザーIDでエラーを期待したがnilが返った")
		}
	})
}

func TestDiaryEntry_SearchDiaryEntriesByUserID_DBError(t *testing.T) {
	db := setupTestDB(t)
	userID := createTestUser(t, db)
	svc := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// DBを閉じてクエリエラーを発生させる
	if err := db.Close(); err != nil {
		t.Fatalf("DB クローズに失敗: %v", err)
	}

	_, err := svc.SearchDiaryEntriesByUserID(context.Background(), userID, "旅行")
	if err == nil {
		t.Fatal("DBエラー時にエラーが返ることを期待したがnilが返った")
	}

	// gRPCラッパー経由でもエラーが伝播することを確認する
	_, err = svc.SearchDiaryEntries(ctx, &g.SearchDiaryEntriesRequest{Keyword: "旅行"})
	if err == nil {
		t.Fatal("gRPCラッパー経由でもDBエラーが返ることを期待したがnilが返った")
	}
}
