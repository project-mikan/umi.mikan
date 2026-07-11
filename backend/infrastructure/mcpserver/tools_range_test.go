package mcpserver

import (
	"testing"

	"github.com/project-mikan/umi.mikan/backend/service/diary"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestGetDiaryEntriesByRangeHandler_Validation(t *testing.T) {
	// バリデーションエラーはDBアクセス前に発生するため、DBなしのDiaryEntryで検証できる
	diaryService := &diary.DiaryEntry{}

	tests := []struct {
		name  string
		input GetDiaryEntriesByRangeInput
	}{
		{
			name:  "異常系: fromが不正な日付形式",
			input: GetDiaryEntriesByRangeInput{From: "2024/05/01", To: "2024-05-10"},
		},
		{
			name:  "異常系: toが不正な日付形式",
			input: GetDiaryEntriesByRangeInput{From: "2024-05-01", To: "not-a-date"},
		},
		{
			name:  "異常系: toがfromより前",
			input: GetDiaryEntriesByRangeInput{From: "2024-05-10", To: "2024-05-01"},
		},
		{
			name:  "異常系: 範囲が広すぎる",
			input: GetDiaryEntriesByRangeInput{From: "2020-01-01", To: "2024-01-01"},
		},
	}

	handler := getDiaryEntriesByRangeHandler(diaryService)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := testutil.CreateAuthenticatedContext(testUUID(t))
			_, _, err := handler(ctx, nil, tt.input)
			if err == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}
		})
	}
}

func TestGetDiaryEntriesByRangeHandler(t *testing.T) {
	t.Run("異常系: 未認証の場合はエラー", func(t *testing.T) {
		diaryService := &diary.DiaryEntry{}
		handler := getDiaryEntriesByRangeHandler(diaryService)

		_, _, err := handler(testutil.CreateUnauthenticatedContext(), nil, GetDiaryEntriesByRangeInput{
			From: "2024-05-01",
			To:   "2024-05-10",
		})
		if err == nil {
			t.Fatal("未認証時にエラーを期待したがnilが返った")
		}
	})

	t.Run("正常系: 範囲内の日記を日付昇順で返す", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		userID := testutil.CreateTestUser(t, db, "mcp-range-test@example.com", "MCPRangeUser")
		diaryService := &diary.DiaryEntry{DB: db}
		ctx := testutil.CreateAuthenticatedContext(userID)

		if _, err := diaryService.CreateDiaryEntry(ctx, createDiaryReq(2024, 5, 1, "5月1日の日記")); err != nil {
			t.Fatalf("日記作成失敗: %v", err)
		}
		if _, err := diaryService.CreateDiaryEntry(ctx, createDiaryReq(2024, 5, 10, "5月10日の日記")); err != nil {
			t.Fatalf("日記作成失敗: %v", err)
		}
		if _, err := diaryService.CreateDiaryEntry(ctx, createDiaryReq(2024, 6, 1, "6月1日の日記")); err != nil {
			t.Fatalf("日記作成失敗: %v", err)
		}

		handler := getDiaryEntriesByRangeHandler(diaryService)
		_, out, err := handler(ctx, nil, GetDiaryEntriesByRangeInput{From: "2024-05-01", To: "2024-05-31"})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(out.Entries) != 2 {
			t.Fatalf("期待件数 2 に対して %d 件取得", len(out.Entries))
		}
		if out.Entries[0].Date != "2024-05-01" || out.Entries[1].Date != "2024-05-10" {
			t.Errorf("日付が正しくフォーマットされていない: %+v", out.Entries)
		}
	})

	t.Run("異常系: DBエラー時はエラーを返す", func(t *testing.T) {
		db := testutil.SetupTestDB(t)
		userID := testutil.CreateTestUser(t, db, "mcp-range-dberror@example.com", "MCPRangeDBErrorUser")
		diaryService := &diary.DiaryEntry{DB: db}
		ctx := testutil.CreateAuthenticatedContext(userID)

		// DBを閉じてクエリエラーを発生させる
		if err := db.Close(); err != nil {
			t.Fatalf("DB クローズに失敗: %v", err)
		}

		handler := getDiaryEntriesByRangeHandler(diaryService)
		_, _, err := handler(ctx, nil, GetDiaryEntriesByRangeInput{From: "2024-05-01", To: "2024-05-31"})
		if err == nil {
			t.Fatal("DBエラー時にエラーが返ることを期待したがnilが返った")
		}
	})
}
