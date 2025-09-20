package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type MonthlySummaryJob struct{}

type MonthlySummaryGenerationMessage struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	Year   int    `json:"year"`
	Month  int    `json:"month"`
}

func (j *MonthlySummaryJob) Name() string {
	return "MonthlySummaryGeneration"
}

func (j *MonthlySummaryJob) Interval() time.Duration {
	return 5 * time.Minute
}

func (j *MonthlySummaryJob) Execute(ctx context.Context, s *Scheduler) error {
	log.Print("Checking for missing monthly summaries...")

	// 1. auto_summary_monthly が true のユーザーを取得
	usersQuery := `
		SELECT user_id
		FROM user_llms
		WHERE auto_summary_monthly = true
	`

	rows, err := s.db.Query(usersQuery)
	if err != nil {
		return fmt.Errorf("failed to query users with auto monthly summary enabled: %w", err)
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
		log.Print("No users with auto monthly summary enabled")
		return nil
	}

	log.Printf("Found %d users with auto monthly summary enabled", len(userIDs))

	// 2. 各ユーザーについて、summaryが作られていない月を確認
	for _, userID := range userIDs {
		if err := j.processUserMonthlySummaries(ctx, s, userID); err != nil {
			log.Printf("Error processing monthly summaries for user %s: %v", userID, err)
			continue
		}
	}

	return nil
}

func (j *MonthlySummaryJob) processUserMonthlySummaries(ctx context.Context, s *Scheduler, userID string) error {
	// diary_summary_daysから該当ユーザーの要約がある年月を取得し、
	// diary_summary_monthsに月次要約がない月を見つける（今月を除く）
	query := `
		SELECT EXTRACT(YEAR FROM dsd.date) as year, EXTRACT(MONTH FROM dsd.date) as month
		FROM diary_summary_days dsd
		LEFT JOIN diary_summary_months dsm ON dsd.user_id = dsm.user_id
			AND EXTRACT(YEAR FROM dsd.date) = dsm.year
			AND EXTRACT(MONTH FROM dsd.date) = dsm.month
		WHERE dsd.user_id = $1
			AND dsm.id IS NULL
			AND (EXTRACT(YEAR FROM dsd.date) < EXTRACT(YEAR FROM CURRENT_DATE)
				OR (EXTRACT(YEAR FROM dsd.date) = EXTRACT(YEAR FROM CURRENT_DATE)
					AND EXTRACT(MONTH FROM dsd.date) < EXTRACT(MONTH FROM CURRENT_DATE)))
		GROUP BY EXTRACT(YEAR FROM dsd.date), EXTRACT(MONTH FROM dsd.date)
		ORDER BY year, month
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return fmt.Errorf("failed to query missing monthly summaries for user %s: %w", userID, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

	type YearMonth struct {
		Year  int
		Month int
	}

	var missingMonths []YearMonth
	for rows.Next() {
		var year, month int
		if err := rows.Scan(&year, &month); err != nil {
			return fmt.Errorf("failed to scan year/month: %w", err)
		}
		missingMonths = append(missingMonths, YearMonth{Year: year, Month: month})
	}

	if len(missingMonths) == 0 {
		log.Printf("No missing monthly summaries for user %s", userID)
		return nil
	}

	log.Printf("Found %d missing monthly summaries for user %s", len(missingMonths), userID)

	// 3. 各年月についてRedisキューにジョブを投入
	for _, ym := range missingMonths {
		message := MonthlySummaryGenerationMessage{
			Type:   "monthly_summary",
			UserID: userID,
			Year:   ym.Year,
			Month:  ym.Month,
		}

		messageBytes, err := json.Marshal(message)
		if err != nil {
			log.Printf("Failed to marshal message for user %s, year %d, month %d: %v", userID, ym.Year, ym.Month, err)
			continue
		}

		// Redisにメッセージを送信
		publishCmd := s.redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
		if err := s.redis.Do(ctx, publishCmd).Error(); err != nil {
			log.Printf("Failed to publish message for user %s, year %d, month %d: %v", userID, ym.Year, ym.Month, err)
			continue
		}

		log.Printf("Queued monthly summary generation for user %s, year %d, month %d", userID, ym.Year, ym.Month)
	}

	return nil
}
