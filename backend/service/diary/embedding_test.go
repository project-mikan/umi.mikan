package diary

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/testutil"
	"github.com/redis/rueidis"
)

// mockGeminiEmbedder はテスト用のGeminiEmbedderモック
type mockGeminiEmbedder struct {
	capturedText string
	returnErr    error
}

func (m *mockGeminiEmbedder) GenerateEmbedding(_ context.Context, text string, _ bool) ([]float32, error) {
	m.capturedText = text
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return []float32{0.1, 0.2, 0.3}, nil
}

func (m *mockGeminiEmbedder) Close() error { return nil }

// mockLLMFactory はテスト用のLLMFactoryモック
type mockLLMFactory struct {
	embedder *mockGeminiEmbedder
	err      error
}

func (f *mockLLMFactory) CreateGeminiClient(_ context.Context, _ string) (GeminiEmbedder, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.embedder, nil
}

func TestIsTodayJST(t *testing.T) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		jst = time.FixedZone("Asia/Tokyo", 9*60*60)
	}

	// 今日のJST日付をUTC 00:00:00で表現したもの（=当日の日記のdate値）
	nowJST := time.Now().In(jst)
	todayUTC := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), 0, 0, 0, 0, time.UTC)

	// 昨日のJST日付をUTC 00:00:00で表現したもの
	yesterdayUTC := todayUTC.AddDate(0, 0, -1)

	// 明日のJST日付をUTC 00:00:00で表現したもの
	tomorrowUTC := todayUTC.AddDate(0, 0, 1)

	t.Run("当日の日記はtrueを返す", func(t *testing.T) {
		if !isTodayJST(todayUTC) {
			t.Error("当日の日記でisTodayJSTがfalseを返した")
		}
	})

	t.Run("昨日の日記はfalseを返す", func(t *testing.T) {
		if isTodayJST(yesterdayUTC) {
			t.Error("昨日の日記でisTodayJSTがtrueを返した")
		}
	})

	t.Run("明日の日記はfalseを返す", func(t *testing.T) {
		if isTodayJST(tomorrowUTC) {
			t.Error("明日の日記でisTodayJSTがtrueを返した")
		}
	})

	t.Run("過去の固定日付はfalseを返す", func(t *testing.T) {
		pastDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		if isTodayJST(pastDate) {
			t.Error("過去の日付でisTodayJSTがtrueを返した")
		}
	})
}

func TestIsYesterdayJST(t *testing.T) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)

	// 固定の「今」を使い、実行日時に依存しないテストにする
	fixedNow := time.Date(2024, 6, 15, 10, 0, 0, 0, jst)
	todayUTC := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	yesterdayUTC := todayUTC.AddDate(0, 0, -1)
	tomorrowUTC := todayUTC.AddDate(0, 0, 1)
	twoDaysAgoUTC := todayUTC.AddDate(0, 0, -2)

	t.Run("昨日の日記はtrueを返す", func(t *testing.T) {
		if !isYesterdayJST(yesterdayUTC, fixedNow) {
			t.Error("昨日の日記でisYesterdayJSTがfalseを返した")
		}
	})

	t.Run("当日の日記はfalseを返す", func(t *testing.T) {
		if isYesterdayJST(todayUTC, fixedNow) {
			t.Error("当日の日記でisYesterdayJSTがtrueを返した")
		}
	})

	t.Run("明日の日記はfalseを返す", func(t *testing.T) {
		if isYesterdayJST(tomorrowUTC, fixedNow) {
			t.Error("明日の日記でisYesterdayJSTがtrueを返した")
		}
	})

	t.Run("2日前の日記はfalseを返す", func(t *testing.T) {
		if isYesterdayJST(twoDaysAgoUTC, fixedNow) {
			t.Error("2日前の日記でisYesterdayJSTがtrueを返した")
		}
	})

	t.Run("過去の固定日付はfalseを返す", func(t *testing.T) {
		pastDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		if isYesterdayJST(pastDate, fixedNow) {
			t.Error("過去の日付でisYesterdayJSTがtrueを返した")
		}
	})

	t.Run("月初の場合に前月末日を正しく昨日と判定する", func(t *testing.T) {
		// 月境界のテスト: 2024-07-01 10:00 JST のとき昨日は 2024-06-30
		monthStartNow := time.Date(2024, 7, 1, 10, 0, 0, 0, jst)
		lastDayOfJune := time.Date(2024, 6, 30, 0, 0, 0, 0, time.UTC)
		if !isYesterdayJST(lastDayOfJune, monthStartNow) {
			t.Error("月初のとき前月末日をisYesterdayJSTがfalseを返した")
		}
	})
}

func TestIsPastDiaryEmbeddingTime(t *testing.T) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)

	tests := []struct {
		name     string
		nowJST   time.Time
		expected bool
	}{
		{
			name:     "4:29 JSTは4:30前のためfalseを返す",
			nowJST:   time.Date(2024, 1, 15, 4, 29, 0, 0, jst),
			expected: false,
		},
		{
			name:     "4:30 JSTはちょうど実行時刻のためtrueを返す",
			nowJST:   time.Date(2024, 1, 15, 4, 30, 0, 0, jst),
			expected: true,
		},
		{
			name:     "4:31 JSTは4:30以降のためtrueを返す",
			nowJST:   time.Date(2024, 1, 15, 4, 31, 0, 0, jst),
			expected: true,
		},
		{
			name:     "5:00 JSTは4:30以降のためtrueを返す",
			nowJST:   time.Date(2024, 1, 15, 5, 0, 0, 0, jst),
			expected: true,
		},
		{
			name:     "0:01 JSTは4:30前のためfalseを返す",
			nowJST:   time.Date(2024, 1, 15, 0, 1, 0, 0, jst),
			expected: false,
		},
		{
			name:     "23:59 JSTは4:30以降のためtrueを返す",
			nowJST:   time.Date(2024, 1, 15, 23, 59, 0, 0, jst),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPastDiaryEmbeddingTime(tt.nowJST)
			if got != tt.expected {
				t.Errorf("isPastDiaryEmbeddingTime(%v) = %v, want %v", tt.nowJST, got, tt.expected)
			}
		})
	}
}

func TestPublishDiaryEmbeddingMessage_SkipConditions(t *testing.T) {
	db := setupTestDB(t)

	// miniredisを直接使用してパブリッシュを監視できるようにする
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis起動失敗: %v", err)
	}
	t.Cleanup(mr.Close)
	redisClient, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{mr.Addr()},
		DisableCache: true,
	})
	if err != nil {
		t.Fatalf("rueidisクライアント作成失敗: %v", err)
	}
	t.Cleanup(redisClient.Close)

	svc := &DiaryEntry{DB: db, Redis: redisClient}
	ctx := context.Background()

	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	nowJST := time.Now().In(jst)
	todayUTC := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), 0, 0, 0, 0, time.UTC)
	twoDaysAgoUTC := todayUTC.AddDate(0, 0, -2)

	// diary_events チャンネルをサブスクライブしてメッセージ到着を監視する
	received := make(chan struct{}, 1)
	subCtx, cancelSub := context.WithCancel(context.Background())
	t.Cleanup(cancelSub)
	go func() {
		//nolint:errcheck // サブスクライブはキャンセルで終了するためエラーは無視する
		redisClient.Receive(subCtx, redisClient.B().Subscribe().Channel("diary_events").Build(), func(_ rueidis.PubSubMessage) {
			select {
			case received <- struct{}{}:
			default:
			}
		})
	}()
	// サブスクリプションが確立されるまで待機
	time.Sleep(20 * time.Millisecond)

	t.Run("2日以上前の日記はメッセージを発行しない", func(t *testing.T) {
		svc.publishDiaryEmbeddingMessage(ctx, "test-user-id", "test-diary-id", twoDaysAgoUTC)
		// 100ms待機してメッセージが届かないことを確認
		select {
		case <-received:
			t.Error("2日以上前の日記でdiary_eventsにメッセージが発行された")
		case <-time.After(100 * time.Millisecond):
			// 期待通りメッセージなし
		}
	})

	t.Run("Redisがnilの場合は何もしない", func(t *testing.T) {
		svcNoRedis := &DiaryEntry{DB: db, Redis: nil}
		svcNoRedis.publishDiaryEmbeddingMessage(ctx, "test-user-id", "test-diary-id", todayUTC)
		// パニックなく完了することを確認
	})
}

func TestGenerateSnippet(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		maxLen   int
		expected string
	}{
		{
			name:     "正常系：maxLen以内のコンテンツはそのまま返す",
			content:  "短いコンテンツ",
			maxLen:   200,
			expected: "短いコンテンツ",
		},
		{
			name:     "正常系：maxLenを超えるコンテンツは切り詰める",
			content:  "あいうえおかきくけこさしすせそたちつてとなにぬねのはひふへほまみむめもやゆよらりるれろわをん",
			maxLen:   10,
			expected: "あいうえおかきくけこ...",
		},
		{
			name:     "正常系：空文字列はそのまま返す",
			content:  "",
			maxLen:   200,
			expected: "",
		},
		{
			name:     "正常系：ちょうどmaxLen文字のコンテンツはそのまま返す",
			content:  "あいうえお",
			maxLen:   5,
			expected: "あいうえお",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateSnippet(tt.content, tt.maxLen)
			if got != tt.expected {
				t.Errorf("generateSnippet(%q, %d) = %q, want %q", tt.content, tt.maxLen, got, tt.expected)
			}
		})
	}
}

func TestGetTaskTimeout(t *testing.T) {
	t.Run("正常系：TASK_TIMEOUT_SECONDSが未設定の場合はデフォルト値600を返す", func(t *testing.T) {
		_ = os.Unsetenv("TASK_TIMEOUT_SECONDS")
		got := getTaskTimeout()
		if got != 600 {
			t.Errorf("getTaskTimeout() = %d, want 600", got)
		}
	})

	t.Run("正常系：TASK_TIMEOUT_SECONDSが設定されている場合はその値を返す", func(t *testing.T) {
		if err := os.Setenv("TASK_TIMEOUT_SECONDS", "300"); err != nil {
			t.Fatalf("Setenv失敗: %v", err)
		}
		defer func() { _ = os.Unsetenv("TASK_TIMEOUT_SECONDS") }()
		got := getTaskTimeout()
		if got != 300 {
			t.Errorf("getTaskTimeout() = %d, want 300", got)
		}
	})

	t.Run("異常系：無効な値の場合はデフォルト値600を返す", func(t *testing.T) {
		if err := os.Setenv("TASK_TIMEOUT_SECONDS", "invalid"); err != nil {
			t.Fatalf("Setenv失敗: %v", err)
		}
		defer func() { _ = os.Unsetenv("TASK_TIMEOUT_SECONDS") }()
		got := getTaskTimeout()
		if got != 600 {
			t.Errorf("getTaskTimeout() = %d, want 600", got)
		}
	})

	t.Run("異常系：0以下の値の場合はデフォルト値600を返す", func(t *testing.T) {
		if err := os.Setenv("TASK_TIMEOUT_SECONDS", "-1"); err != nil {
			t.Fatalf("Setenv失敗: %v", err)
		}
		defer func() { _ = os.Unsetenv("TASK_TIMEOUT_SECONDS") }()
		got := getTaskTimeout()
		if got != 600 {
			t.Errorf("getTaskTimeout() = %d, want 600", got)
		}
	})
}

func TestDiaryEntry_GetDiaryEmbeddingStatus(t *testing.T) {
	db := setupTestDB(t)
	userID := createTestUser(t, db)
	svc := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// テスト用日記を作成
	createResp, err := svc.CreateDiaryEntry(ctx, &g.CreateDiaryEntryRequest{
		Content: "埋め込みステータステスト用日記",
		Date:    &g.YMD{Year: 2024, Month: 10, Day: 1},
	})
	if err != nil {
		t.Fatalf("日記エントリの作成に失敗: %v", err)
	}
	diaryID := createResp.Entry.Id

	t.Run("正常系：embeddingが存在しない日記のステータスを取得", func(t *testing.T) {
		resp, err := svc.GetDiaryEmbeddingStatus(ctx, &g.GetDiaryEmbeddingStatusRequest{
			DiaryId: diaryID,
		})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if resp.Indexed {
			t.Error("embeddingが存在しないのにIndexed=trueになっている")
		}
	})

	t.Run("異常系：無効な日記ID", func(t *testing.T) {
		_, err := svc.GetDiaryEmbeddingStatus(ctx, &g.GetDiaryEmbeddingStatusRequest{
			DiaryId: "invalid-uuid",
		})
		if err == nil {
			t.Error("無効なIDでエラーが返らなかった")
		}
	})

	t.Run("異常系：存在しない日記ID", func(t *testing.T) {
		_, err := svc.GetDiaryEmbeddingStatus(ctx, &g.GetDiaryEmbeddingStatusRequest{
			DiaryId: uuid.New().String(),
		})
		if err == nil {
			t.Error("存在しないIDでエラーが返らなかった")
		}
	})

	t.Run("異常系：他のユーザーの日記に対してアクセス拒否", func(t *testing.T) {
		otherUserID := testutil.CreateTestUser(t, db, "other-embedding-user@example.com", "Other User")
		otherCtx := createAuthenticatedContext(otherUserID)

		_, err := svc.GetDiaryEmbeddingStatus(otherCtx, &g.GetDiaryEmbeddingStatusRequest{
			DiaryId: diaryID,
		})
		if err == nil {
			t.Error("他のユーザーの日記へのアクセスでエラーが返らなかった")
		}
	})
}

func TestDiaryEntry_GetDiaryEmbeddingStatus_WithEmbedding(t *testing.T) {
	db := setupTestDB(t)
	userID := createTestUser(t, db)
	svc := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// テスト用日記を作成
	diaryID := uuid.New()
	now := time.Now().UnixMilli()
	_, err := db.ExecContext(ctx,
		`INSERT INTO diaries (id, user_id, content, date, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		diaryID, userID, "embeddingテスト日記", "2024-10-02", now, now,
	)
	if err != nil {
		t.Fatalf("日記の挿入に失敗: %v", err)
	}

	// embeddingを直接挿入
	dummyEmbedding := make([]float32, 3072)
	var embStr strings.Builder
	embStr.WriteString("[")
	for i, v := range dummyEmbedding {
		if i > 0 {
			embStr.WriteString(",")
		}
		embStr.WriteString("0")
		_ = v
	}
	embStr.WriteString("]")

	_, err = db.ExecContext(ctx,
		`INSERT INTO diary_embeddings (id, diary_id, user_id, chunk_index, chunk_content, chunk_summary, embedding, model_version, chunk_model_version, created_at, updated_at)
		 VALUES ($1, $2, $3, 0, 'test chunk', 'test summary', $4::halfvec, 'gemini-embedding-001', '', NOW(), NOW())`,
		uuid.New(), diaryID, userID, embStr.String(),
	)
	if err != nil {
		t.Fatalf("diary_embeddingsの挿入に失敗: %v", err)
	}

	resp, err := svc.GetDiaryEmbeddingStatus(ctx, &g.GetDiaryEmbeddingStatusRequest{
		DiaryId: diaryID.String(),
	})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if !resp.Indexed {
		t.Error("embeddingが存在するのにIndexed=falseになっている")
	}
	if resp.ModelVersion != "gemini-embedding-001" {
		t.Errorf("ModelVersion: got %q, want %q", resp.ModelVersion, "gemini-embedding-001")
	}
	if resp.ChunkCount != 1 {
		t.Errorf("ChunkCount: got %d, want 1", resp.ChunkCount)
	}
}

func TestDiaryEntry_RegenerateAllEmbeddings_NoLLMKey(t *testing.T) {
	db := setupTestDB(t)
	userID := createTestUser(t, db)
	svc := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// LLMキーが存在しない場合はエラーを返す
	_, err := svc.RegenerateAllEmbeddings(ctx, &g.RegenerateAllEmbeddingsRequest{})
	if err == nil {
		t.Error("LLMキーが存在しないのにエラーが返らなかった")
	}
}

func TestDiaryEntry_SearchDiaryEntriesSemantic_NoLLMKey(t *testing.T) {
	db := setupTestDB(t)
	userID := createTestUser(t, db)
	svc := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	_, err := svc.SearchDiaryEntriesSemantic(ctx, &g.SearchDiaryEntriesSemanticRequest{
		Query: "テスト検索クエリ",
	})
	if err == nil {
		t.Error("LLMキーが存在しないのにエラーが返らなかった")
	}
}

func TestDiaryEntry_GenerateMonthlySummary_NoDiaries(t *testing.T) {
	db := setupTestDB(t)
	userID := createTestUser(t, db)
	svc := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	// LLMキーを作成（GenerateMonthlySummaryはLLMキーチェックを通過する必要がある）
	testutil.CreateTestUserLLM(t, db, userID, "test-api-key")

	// 日記が存在しない過去月に対してサマリー生成を要求する
	_, err := svc.GenerateMonthlySummary(ctx, &g.GenerateMonthlySummaryRequest{
		Month: &g.YM{Year: 2020, Month: 1},
	})
	if err == nil {
		t.Error("日記が存在しない月でエラーが返らなかった")
	}
}

func TestDiaryEntry_SearchDiaryEntriesSemantic_EmptyQuery(t *testing.T) {
	db := setupTestDB(t)
	userID := createTestUser(t, db)
	svc := &DiaryEntry{DB: db}
	ctx := createAuthenticatedContext(userID)

	_, err := svc.SearchDiaryEntriesSemantic(ctx, &g.SearchDiaryEntriesSemanticRequest{
		Query: "",
	})
	if err == nil {
		t.Error("空クエリでエラーが返らなかった")
	}
}

func setupTestRedisForDiary(t *testing.T) rueidis.Client {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis起動失敗: %v", err)
	}
	t.Cleanup(mr.Close)
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{mr.Addr()},
		DisableCache: true,
	})
	if err != nil {
		t.Fatalf("rueidisクライアント作成失敗: %v", err)
	}
	t.Cleanup(client.Close)
	return client
}

func TestDiaryEntry_RegenerateAllEmbeddings_SemanticEnabled(t *testing.T) {
	db := setupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "regen-embedding@example.com", "Regen Test User")
	ctx := createAuthenticatedContext(userID)

	// semantic_search_enabled=trueでuser_llmsを挿入
	testutil.CreateTestUserLLMWithSettings(t, db, userID, "test-api-key", false, false, true)

	redisClient := setupTestRedisForDiary(t)
	svc := &DiaryEntry{DB: db, Redis: redisClient}

	t.Run("正常系：日記が存在しない場合はQueuedCount=0で成功する", func(t *testing.T) {
		resp, err := svc.RegenerateAllEmbeddings(ctx, &g.RegenerateAllEmbeddingsRequest{})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if !resp.Success {
			t.Error("Successがfalseになっている")
		}
		if resp.QueuedCount != 0 {
			t.Errorf("QueuedCount: got %d, want 0", resp.QueuedCount)
		}
	})
}

func TestDiaryEntry_SearchDiaryEntriesSemantic_EnrichedQuery(t *testing.T) {
	db := setupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "semantic-enriched@example.com", "Semantic Test User")
	ctx := createAuthenticatedContext(userID)

	// semantic_search_enabled=trueでuser_llmsを挿入
	testutil.CreateTestUserLLMWithSettings(t, db, userID, "test-api-key", false, false, true)

	embedder := &mockGeminiEmbedder{}
	svc := &DiaryEntry{
		DB:         db,
		LLMFactory: &mockLLMFactory{embedder: embedder},
	}

	query := "最近の日記を教えて"
	_, _ = svc.SearchDiaryEntriesSemantic(ctx, &g.SearchDiaryEntriesSemanticRequest{
		Query: query,
	})

	// クエリに現在日付が付与されていることを確認
	jst := time.FixedZone("JST", 9*60*60)
	n := time.Now().In(jst)
	expectedPrefix := fmt.Sprintf("今日は%d年%d月%d日。", n.Year(), int(n.Month()), n.Day())
	if !strings.Contains(embedder.capturedText, expectedPrefix) {
		t.Errorf("enrichedQueryに日付が付与されていない: got %q, want prefix %q", embedder.capturedText, expectedPrefix)
	}
	if !strings.Contains(embedder.capturedText, query) {
		t.Errorf("enrichedQueryに元のクエリが含まれていない: got %q", embedder.capturedText)
	}
}
