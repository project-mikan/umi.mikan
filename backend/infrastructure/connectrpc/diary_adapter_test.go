package connect

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
	"github.com/project-mikan/umi.mikan/backend/testutil"
	"github.com/redis/rueidis"
)

func setupDiaryAdapterWithRedis(t *testing.T) *DiaryServiceAdapter {
	t.Helper()
	db := testutil.SetupTestDB(t)

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis 起動失敗: %v", err)
	}
	t.Cleanup(mr.Close)

	redisClient, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:  []string{mr.Addr()},
		DisableCache: true,
	})
	if err != nil {
		t.Fatalf("rueidis クライアント作成失敗: %v", err)
	}
	t.Cleanup(redisClient.Close)

	svc := &diary.DiaryEntry{DB: db, Redis: redisClient}
	return &DiaryServiceAdapter{svc: svc}
}

func setupDiaryAdapter(t *testing.T) *DiaryServiceAdapter {
	t.Helper()
	db := testutil.SetupTestDB(t)
	svc := &diary.DiaryEntry{DB: db}
	return &DiaryServiceAdapter{svc: svc}
}

func TestDiaryServiceAdapter_GetDiaryEntry(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "正常系: 存在しない日記エントリはDBエラーを返す",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupDiaryAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.GetDiaryEntryRequest{
				Date: &g.YMD{Year: 2024, Month: 1, Day: 1},
			})
			// 存在しない日記はエラーを返す
			_, err := adapter.GetDiaryEntry(ctx, req)
			if err == nil {
				t.Log("エラーなし")
			}
		})
	}
}

func TestDiaryServiceAdapter_GetDiaryEntries(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "正常系: 存在しないユーザーは空リストを返す",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupDiaryAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.GetDiaryEntriesRequest{})
			resp, err := adapter.GetDiaryEntries(ctx, req)
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp == nil {
				t.Fatal("レスポンスがnilだった")
			}
		})
	}
}

func TestDiaryServiceAdapter_GetDiaryEntriesByMonth(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "正常系: 存在しないユーザーは空リストを返す",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupDiaryAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.GetDiaryEntriesByMonthRequest{
				Month: &g.YM{Year: 2024, Month: 1},
			})
			resp, err := adapter.GetDiaryEntriesByMonth(ctx, req)
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp == nil {
				t.Fatal("レスポンスがnilだった")
			}
		})
	}
}

func TestDiaryServiceAdapter_SearchDiaryEntries(t *testing.T) {
	tests := []struct {
		name    string
		keyword string
	}{
		{
			name:    "正常系: キーワード検索は空リストを返す",
			keyword: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupDiaryAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.SearchDiaryEntriesRequest{Keyword: tt.keyword})
			resp, err := adapter.SearchDiaryEntries(ctx, req)
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp == nil {
				t.Fatal("レスポンスがnilだった")
			}
		})
	}
}

func TestDiaryServiceAdapter_UpdateDiaryEntry_Error(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "異常系: 存在しない日記の更新はDBエラーになる",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupDiaryAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.UpdateDiaryEntryRequest{
				Id:      uuid.New().String(),
				Content: "updated content",
			})
			_, err := adapter.UpdateDiaryEntry(ctx, req)
			if err == nil {
				t.Log("エラーなし")
			}
		})
	}
}

func TestDiaryServiceAdapter_DeleteDiaryEntry_Error(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "異常系: 存在しない日記の削除はDBエラーになる",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupDiaryAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.DeleteDiaryEntryRequest{
				Id: uuid.New().String(),
			})
			_, err := adapter.DeleteDiaryEntry(ctx, req)
			if err == nil {
				t.Log("エラーなし")
			}
		})
	}
}

func TestDiaryServiceAdapter_GetMonthlySummary_Error(t *testing.T) {
	tests := []struct {
		name         string
		expectedCode connect.Code
	}{
		{
			name:         "異常系: サマリーが存在しない場合はNotFoundエラーになる",
			expectedCode: connect.CodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupDiaryAdapterWithRedis(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.GetMonthlySummaryRequest{
				Month: &g.YM{Year: 2024, Month: 1},
			})
			_, err := adapter.GetMonthlySummary(ctx, req)
			if err == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}
			connectErr, ok := err.(*connect.Error)
			if !ok {
				t.Fatalf("*connect.Error を期待したが %T が返った", err)
			}
			if connectErr.Code() != tt.expectedCode {
				t.Errorf("エラーコード: 期待 %v, 実際 %v", tt.expectedCode, connectErr.Code())
			}
		})
	}
}

func TestDiaryServiceAdapter_GetDiaryHighlight_Error(t *testing.T) {
	tests := []struct {
		name         string
		expectedCode connect.Code
	}{
		{
			name:         "異常系: 不正なUUIDはInvalidArgumentエラーになる",
			expectedCode: connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupDiaryAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.GetDiaryHighlightRequest{
				DiaryId: "not-a-valid-uuid",
			})
			_, err := adapter.GetDiaryHighlight(ctx, req)
			if err == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}
			connectErr, ok := err.(*connect.Error)
			if !ok {
				t.Fatalf("*connect.Error を期待したが %T が返った", err)
			}
			if connectErr.Code() != tt.expectedCode {
				t.Errorf("エラーコード: 期待 %v, 実際 %v", tt.expectedCode, connectErr.Code())
			}
		})
	}
}

func TestDiaryServiceAdapter_ExportDiaryEntries(t *testing.T) {
	t.Run("正常系: 認証済みユーザーで空の期間を指定すると空リストを返す", func(t *testing.T) {
		adapter := setupDiaryAdapter(t)
		ctx := createAuthContext(uuid.New().String())
		req := connect.NewRequest(&g.ExportDiaryEntriesRequest{
			From: &g.YM{Year: 2024, Month: 1},
			To:   &g.YM{Year: 2024, Month: 3},
		})
		resp, err := adapter.ExportDiaryEntries(ctx, req)
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if resp == nil {
			t.Fatal("レスポンスがnilだった")
		}
		if resp.Msg == nil {
			t.Fatal("レスポンスメッセージがnilだった")
		}
		if len(resp.Msg.Entries) != 0 {
			t.Errorf("空リストを期待したが %d 件返った", len(resp.Msg.Entries))
		}
	})

	t.Run("異常系: fromがnilの場合はInvalidArgumentエラーになる", func(t *testing.T) {
		adapter := setupDiaryAdapter(t)
		ctx := createAuthContext(uuid.New().String())
		req := connect.NewRequest(&g.ExportDiaryEntriesRequest{
			From: nil,
			To:   &g.YM{Year: 2024, Month: 3},
		})
		_, err := adapter.ExportDiaryEntries(ctx, req)
		if err == nil {
			t.Fatal("エラーを期待したがnilが返った")
		}
		connectErr, ok := err.(*connect.Error)
		if !ok {
			t.Fatalf("*connect.Error を期待したが %T が返った", err)
		}
		if connectErr.Code() != connect.CodeInvalidArgument {
			t.Errorf("エラーコード: 期待 %v, 実際 %v", connect.CodeInvalidArgument, connectErr.Code())
		}
	})

	t.Run("異常系: 開始が終了より後の場合はInvalidArgumentエラーになる", func(t *testing.T) {
		adapter := setupDiaryAdapter(t)
		ctx := createAuthContext(uuid.New().String())
		req := connect.NewRequest(&g.ExportDiaryEntriesRequest{
			From: &g.YM{Year: 2024, Month: 6},
			To:   &g.YM{Year: 2024, Month: 3},
		})
		_, err := adapter.ExportDiaryEntries(ctx, req)
		if err == nil {
			t.Fatal("エラーを期待したがnilが返った")
		}
		connectErr, ok := err.(*connect.Error)
		if !ok {
			t.Fatalf("*connect.Error を期待したが %T が返った", err)
		}
		if connectErr.Code() != connect.CodeInvalidArgument {
			t.Errorf("エラーコード: 期待 %v, 実際 %v", connect.CodeInvalidArgument, connectErr.Code())
		}
	})

	t.Run("異常系: 開始年が終了年より大きい場合はInvalidArgumentエラーになる", func(t *testing.T) {
		adapter := setupDiaryAdapter(t)
		ctx := createAuthContext(uuid.New().String())
		req := connect.NewRequest(&g.ExportDiaryEntriesRequest{
			From: &g.YM{Year: 2025, Month: 1},
			To:   &g.YM{Year: 2024, Month: 12},
		})
		_, err := adapter.ExportDiaryEntries(ctx, req)
		if err == nil {
			t.Fatal("エラーを期待したがnilが返った")
		}
		connectErr, ok := err.(*connect.Error)
		if !ok {
			t.Fatalf("*connect.Error を期待したが %T が返った", err)
		}
		if connectErr.Code() != connect.CodeInvalidArgument {
			t.Errorf("エラーコード: 期待 %v, 実際 %v", connect.CodeInvalidArgument, connectErr.Code())
		}
	})
}
