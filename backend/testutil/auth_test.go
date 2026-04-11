package testutil

import (
	"context"
	"testing"
)

func TestCreateTestUserLLM(t *testing.T) {
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "create-user-llm@example.com", "User")

	CreateTestUserLLM(t, db, userID, "test-api-key")

	var count int
	if err := db.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM user_llms WHERE user_id = $1`, userID,
	).Scan(&count); err != nil {
		t.Fatalf("user_llmsのカウントクエリ失敗: %v", err)
	}
	if count != 1 {
		t.Errorf("期待 1, 実際 %d", count)
	}
}

func TestCreateTestUserLLMWithSettings(t *testing.T) {
	db := SetupTestDB(t)

	tests := []struct {
		name        string
		autoMonthly bool
		autoTrend   bool
		semantic    bool
	}{
		{"月次サマリー有効", true, false, false},
		{"トレンド有効", false, true, false},
		{"セマンティック検索有効", false, false, true},
		{"全て無効", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID := CreateTestUser(t, db, "user-llm-settings@example.com", "User")
			CreateTestUserLLMWithSettings(t, db, userID, "test-key", tt.autoMonthly, tt.autoTrend, tt.semantic)

			var autoMonthly, autoTrend, semantic bool
			if err := db.QueryRowContext(context.Background(),
				`SELECT auto_summary_monthly, auto_latest_trend_enabled, semantic_search_enabled FROM user_llms WHERE user_id = $1`,
				userID,
			).Scan(&autoMonthly, &autoTrend, &semantic); err != nil {
				t.Fatalf("user_llmsの取得失敗: %v", err)
			}
			if autoMonthly != tt.autoMonthly {
				t.Errorf("auto_summary_monthly: 期待 %v, 実際 %v", tt.autoMonthly, autoMonthly)
			}
			if autoTrend != tt.autoTrend {
				t.Errorf("auto_latest_trend_enabled: 期待 %v, 実際 %v", tt.autoTrend, autoTrend)
			}
			if semantic != tt.semantic {
				t.Errorf("semantic_search_enabled: 期待 %v, 実際 %v", tt.semantic, semantic)
			}
		})
	}
}
