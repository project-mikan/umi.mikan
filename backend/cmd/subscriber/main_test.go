package main

import (
	"context"
	"testing"

	"github.com/project-mikan/umi.mikan/backend/testutil"
	"github.com/sirupsen/logrus"
)

func TestSummaryGenerationMessage(t *testing.T) {
	msg := SummaryGenerationMessage{
		Type:   "daily_summary",
		UserID: "test-user-id",
		Date:   "2024-01-15",
	}

	if msg.Type != "daily_summary" {
		t.Errorf("expected type 'daily_summary', got '%s'", msg.Type)
	}

	if msg.UserID != "test-user-id" {
		t.Errorf("expected user ID 'test-user-id', got '%s'", msg.UserID)
	}

	if msg.Date != "2024-01-15" {
		t.Errorf("expected date '2024-01-15', got '%s'", msg.Date)
	}
}

func TestMonthlySummaryGenerationMessage(t *testing.T) {
	msg := MonthlySummaryGenerationMessage{
		Type:   "monthly_summary",
		UserID: "test-user-id",
		Year:   2024,
		Month:  1,
	}

	if msg.Type != "monthly_summary" {
		t.Errorf("expected type 'monthly_summary', got '%s'", msg.Type)
	}

	if msg.UserID != "test-user-id" {
		t.Errorf("expected user ID 'test-user-id', got '%s'", msg.UserID)
	}

	if msg.Year != 2024 {
		t.Errorf("expected year 2024, got %d", msg.Year)
	}

	if msg.Month != 1 {
		t.Errorf("expected month 1, got %d", msg.Month)
	}
}

func TestProcessMessage_UnknownType(t *testing.T) {
	ctx := context.Background()
	logger := logrus.NewEntry(logrus.New())

	payload := `{"type": "unknown_type", "user_id": "test"}`

	// This should not return an error for unknown message types
	err := processMessage(ctx, nil, nil, nil, nil, payload, logger)
	if err != nil {
		t.Errorf("expected no error for unknown message type, got %v", err)
	}
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	logger := logrus.NewEntry(logrus.New())

	payload := `invalid json`

	// This should return an error for invalid JSON
	err := processMessage(ctx, nil, nil, nil, nil, payload, logger)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestProcessMessage_LatestTrend_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	logger := logrus.NewEntry(logrus.New())

	// latestTrendメッセージのJSONが不正な場合はエラーを返すことを確認
	payload := `{"type": "latest_trend", invalid_json}`

	err := processMessage(ctx, nil, nil, nil, nil, payload, logger)
	if err == nil {
		t.Fatal("不正なJSONに対してエラーが期待されますが、nilが返りました")
	}
}

func TestProcessMessage_DiaryHighlight_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	logger := logrus.NewEntry(logrus.New())

	// diaryHighlightメッセージのJSONが不正な場合はエラーを返すことを確認
	payload := `{"type": "diary_highlight", invalid_json}`

	err := processMessage(ctx, nil, nil, nil, nil, payload, logger)
	if err == nil {
		t.Fatal("不正なJSONに対してエラーが期待されますが、nilが返りました")
	}
}

func TestGenerateDiaryHighlightWithLLM_NoLLMConfig(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "subscriber-highlight-test@example.com", "Subscriber Test User")
	ctx := context.Background()
	logger := logrus.NewEntry(logrus.New())

	// LLM設定なしのユーザーでgenerateDiaryHighlightWithLLMを呼び出すと、
	// user_llmsテーブルにレコードがないためエラーが返ることを確認
	_, err := generateDiaryHighlightWithLLM(ctx, db, nil, userID.String(), "テストコンテンツ", logger)
	if err == nil {
		t.Fatal("LLM設定なしの場合はエラーが期待されますが、nilが返りました")
	}
}
