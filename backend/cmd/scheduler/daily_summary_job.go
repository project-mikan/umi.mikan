package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type DailySummaryJob struct{}

type SummaryGenerationMessage struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	Date   string `json:"date"` // YYYY-MM-DD format
}

func (j *DailySummaryJob) Name() string {
	return "DailySummaryGeneration"
}

func (j *DailySummaryJob) Interval() time.Duration {
	return 5 * time.Minute
}

func (j *DailySummaryJob) Execute(ctx context.Context, s *Scheduler) error {
	log.Print("Checking for missing daily summaries...")

	// 1. auto_summary_daily が true のユーザーを取得
	usersQuery := `
		SELECT user_id
		FROM user_llms
		WHERE auto_summary_daily = true
	`

	rows, err := s.db.Query(usersQuery)
	if err != nil {
		return fmt.Errorf("failed to query users with auto summary enabled: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	if len(userIDs) == 0 {
		log.Print("No users with auto daily summary enabled")
		return nil
	}

	log.Printf("Found %d users with auto daily summary enabled", len(userIDs))

	// 2. 各ユーザーについて、summaryが作られていない日を確認
	for _, userID := range userIDs {
		if err := j.processUserSummaries(ctx, s, userID); err != nil {
			log.Printf("Error processing summaries for user %s: %v", userID, err)
			continue
		}
	}

	return nil
}

func (j *DailySummaryJob) processUserSummaries(ctx context.Context, s *Scheduler, userID string) error {
	// diariesテーブルから該当ユーザーの日記がある日を取得し、
	// diary_summary_daysにsummaryがない日を見つける（今日を除く）
	query := `
		SELECT d.date
		FROM diaries d
		LEFT JOIN diary_summary_days dsd ON d.user_id = dsd.user_id AND d.date = dsd.date
		WHERE d.user_id = $1 AND dsd.id IS NULL AND d.date < CURRENT_DATE
		ORDER BY d.date
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return fmt.Errorf("failed to query missing summaries for user %s: %w", userID, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

	var missingDates []time.Time
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err != nil {
			return fmt.Errorf("failed to scan date: %w", err)
		}
		missingDates = append(missingDates, date)
	}

	if len(missingDates) == 0 {
		log.Printf("No missing summaries for user %s", userID)
		return nil
	}

	log.Printf("Found %d missing summaries for user %s", len(missingDates), userID)

	// 3. 各日付についてRedisキューにジョブを投入
	for _, date := range missingDates {
		message := SummaryGenerationMessage{
			Type:   "daily_summary",
			UserID: userID,
			Date:   date.Format("2006-01-02"),
		}

		messageBytes, err := json.Marshal(message)
		if err != nil {
			log.Printf("Failed to marshal message for user %s, date %s: %v", userID, date.Format("2006-01-02"), err)
			continue
		}

		// Redisにメッセージを送信
		publishCmd := s.redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
		if err := s.redis.Do(ctx, publishCmd).Error(); err != nil {
			log.Printf("Failed to publish message for user %s, date %s: %v", userID, date.Format("2006-01-02"), err)
			continue
		}

		log.Printf("Queued summary generation for user %s, date %s", userID, date.Format("2006-01-02"))
	}

	return nil
}
