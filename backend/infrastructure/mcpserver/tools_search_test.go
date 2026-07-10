package mcpserver

import (
	"testing"

	"github.com/project-mikan/umi.mikan/backend/service/diary"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestSearchDiaryEntriesFulltextHandler_Validation(t *testing.T) {
	diaryService := &diary.DiaryEntry{}
	handler := searchDiaryEntriesFulltextHandler(diaryService)

	t.Run("異常系: keywordが空の場合はエラー", func(t *testing.T) {
		ctx := testutil.CreateAuthenticatedContext(testUUID(t))
		_, _, err := handler(ctx, nil, SearchDiaryEntriesFulltextInput{Keyword: ""})
		if err == nil {
			t.Fatal("エラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: 未認証の場合はエラー", func(t *testing.T) {
		_, _, err := handler(testutil.CreateUnauthenticatedContext(), nil, SearchDiaryEntriesFulltextInput{Keyword: "旅行"})
		if err == nil {
			t.Fatal("エラーを期待したがnilが返った")
		}
	})
}

func TestSearchDiaryEntriesFulltextHandler_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "mcp-fulltext-test@example.com", "MCPFulltextUser")
	diaryService := &diary.DiaryEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	if _, err := diaryService.CreateDiaryEntry(ctx, createDiaryReq(2024, 5, 1, "今日は旅行に行った")); err != nil {
		t.Fatalf("日記作成失敗: %v", err)
	}
	if _, err := diaryService.CreateDiaryEntry(ctx, createDiaryReq(2024, 5, 2, "今日は仕事をした")); err != nil {
		t.Fatalf("日記作成失敗: %v", err)
	}

	handler := searchDiaryEntriesFulltextHandler(diaryService)
	_, out, err := handler(ctx, nil, SearchDiaryEntriesFulltextInput{Keyword: "旅行"})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if len(out.Entries) != 1 {
		t.Fatalf("期待件数 1 に対して %d 件取得", len(out.Entries))
	}
	if out.Entries[0].Content != "今日は旅行に行った" {
		t.Errorf("マッチした内容が正しくない: %+v", out.Entries[0])
	}
}

func TestSearchDiaryEntriesFuzzyHandler_Validation(t *testing.T) {
	diaryService := &diary.DiaryEntry{}
	handler := searchDiaryEntriesFuzzyHandler(diaryService)

	t.Run("異常系: queryが空の場合はエラー", func(t *testing.T) {
		ctx := testutil.CreateAuthenticatedContext(testUUID(t))
		_, _, err := handler(ctx, nil, SearchDiaryEntriesFuzzyInput{Query: ""})
		if err == nil {
			t.Fatal("エラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: 未認証の場合はエラー", func(t *testing.T) {
		_, _, err := handler(testutil.CreateUnauthenticatedContext(), nil, SearchDiaryEntriesFuzzyInput{Query: "最近の出来事"})
		if err == nil {
			t.Fatal("エラーを期待したがnilが返った")
		}
	})
}

func TestSearchDiaryEntriesFuzzyHandler_NoLLMKey(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "mcp-fuzzy-test@example.com", "MCPFuzzyUser")
	diaryService := &diary.DiaryEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	handler := searchDiaryEntriesFuzzyHandler(diaryService)
	_, _, err := handler(ctx, nil, SearchDiaryEntriesFuzzyInput{Query: "最近の出来事"})
	if err == nil {
		t.Fatal("LLMキー未設定時にエラーを期待したがnilが返った")
	}
}

func TestSearchDiaryEntriesFuzzyHandler_Success(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "mcp-fuzzy-success@example.com", "MCPFuzzySuccessUser")
	testutil.CreateTestUserLLMWithSettings(t, db, userID, "test-api-key", false, false, true)
	diaryService := &diary.DiaryEntry{DB: db, LLMFactory: &mockLLMFactory{}}
	ctx := testutil.CreateAuthenticatedContext(userID)

	// embeddingは存在しないため、ハイブリッド検索のキーワード補完でヒットさせる
	if _, err := diaryService.CreateDiaryEntry(ctx, createDiaryReq(2024, 5, 1, "京都へ旅行に行った")); err != nil {
		t.Fatalf("日記作成失敗: %v", err)
	}

	handler := searchDiaryEntriesFuzzyHandler(diaryService)
	_, out, err := handler(ctx, nil, SearchDiaryEntriesFuzzyInput{Query: "旅行", Limit: 10})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if len(out.Results) != 1 {
		t.Fatalf("期待件数 1 に対して %d 件取得: %+v", len(out.Results), out.Results)
	}
	result := out.Results[0]
	if result.Date != "2024-05-01" {
		t.Errorf("Date: 期待 %q, 実際 %q", "2024-05-01", result.Date)
	}
	if result.Snippet != "京都へ旅行に行った" {
		t.Errorf("Snippet: 期待 %q, 実際 %q", "京都へ旅行に行った", result.Snippet)
	}
}

func TestSearchDiaryEntriesFulltextHandler_DBError(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "mcp-fulltext-dberror@example.com", "MCPFtDBErrUser")
	diaryService := &diary.DiaryEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	// DBを閉じてクエリエラーを発生させる
	if err := db.Close(); err != nil {
		t.Fatalf("DB クローズに失敗: %v", err)
	}

	handler := searchDiaryEntriesFulltextHandler(diaryService)
	_, _, err := handler(ctx, nil, SearchDiaryEntriesFulltextInput{Keyword: "旅行"})
	if err == nil {
		t.Fatal("DBエラー時にエラーが返ることを期待したがnilが返った")
	}
}
