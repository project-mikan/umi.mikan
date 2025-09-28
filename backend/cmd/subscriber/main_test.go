package main

import (
	"context"
	"testing"

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
