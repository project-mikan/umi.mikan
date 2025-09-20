package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/redis/rueidis"
)

type SummaryGenerationMessage struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	Date   string `json:"date"` // YYYY-MM-DD format
}

type MonthlySummaryGenerationMessage struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	Year   int    `json:"year"`
	Month  int    `json:"month"`
}

func main() {
	log.Print("=== umi.mikan subscriber started ===")

	// DB設定の読み込み
	dbConfig, err := constants.LoadDBConfig()
	if err != nil {
		log.Fatalf("Failed to load DB config: %v", err)
	}

	// DB接続
	db := database.NewDB(dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database connection: %v", err)
		}
	}()

	// Redis設定の読み込み
	redisConfig, err := constants.LoadRedisConfig()
	if err != nil {
		log.Fatalf("Failed to load Redis config: %v", err)
	}

	// Redisクライアント作成
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)},
	})
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Redis接続確認
	pingCmd := client.B().Ping().Build()
	if err := client.Do(ctx, pingCmd).Error(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Print("Connected to Redis successfully")

	log.Print("Subscriber is listening for messages...")

	// SUBSCRIBE コマンドでチャンネル購読
	err = client.Receive(ctx, client.B().Subscribe().Channel("diary_events").Build(), func(msg rueidis.PubSubMessage) {
		log.Printf("Received message: %s from channel: %s", msg.Message, msg.Channel)

		err := processMessage(ctx, db, msg.Message)
		if err != nil {
			log.Printf("Failed to process message: %v", err)
		}
	})

	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	log.Print("Subscriber ended")
}

func processMessage(ctx context.Context, db *database.DB, payload string) error {
	// まずメッセージタイプを確認
	var baseMessage struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(payload), &baseMessage); err != nil {
		return fmt.Errorf("failed to unmarshal base message: %w", err)
	}

	switch baseMessage.Type {
	case "daily_summary":
		var message SummaryGenerationMessage
		if err := json.Unmarshal([]byte(payload), &message); err != nil {
			return fmt.Errorf("failed to unmarshal daily summary message: %w", err)
		}
		return generateDailySummary(ctx, db, message.UserID, message.Date)
	case "monthly_summary":
		var message MonthlySummaryGenerationMessage
		if err := json.Unmarshal([]byte(payload), &message); err != nil {
			return fmt.Errorf("failed to unmarshal monthly summary message: %w", err)
		}
		return generateMonthlySummary(ctx, db, message.UserID, message.Year, message.Month)
	default:
		log.Printf("Unknown message type: %s", baseMessage.Type)
		return nil
	}
}

func generateDailySummary(ctx context.Context, db *database.DB, userID, dateStr string) error {
	log.Printf("Generating daily summary for user %s, date %s", userID, dateStr)

	// 1. 指定された日の日記内容を取得
	var diaryContent string
	query := `SELECT content FROM diaries WHERE user_id = $1 AND date = $2`
	err := db.QueryRow(query, userID, dateStr).Scan(&diaryContent)
	if err != nil {
		return fmt.Errorf("failed to get diary content: %w", err)
	}

	// 2. LLMで要約生成 (TODO: 実際のLLM API呼び出しを実装)
	summary := generateSummaryWithLLM(diaryContent)

	// 3. diary_summary_daysに保存
	insertQuery := `
		INSERT INTO diary_summary_days (id, user_id, date, summary, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, date) DO UPDATE SET
		summary = EXCLUDED.summary,
		updated_at = EXCLUDED.updated_at
	`

	now := time.Now().Unix()
	summaryID := uuid.New()

	_, err = db.Exec(insertQuery, summaryID, userID, dateStr, summary, now, now)
	if err != nil {
		return fmt.Errorf("failed to save summary: %w", err)
	}

	log.Printf("Successfully generated and saved summary for user %s, date %s", userID, dateStr)
	return nil
}

func generateMonthlySummary(ctx context.Context, db *database.DB, userID string, year, month int) error {
	log.Printf("Generating monthly summary for user %s, year %d, month %d", userID, year, month)

	// 1. 指定された年月の日次要約を全て取得
	query := `
		SELECT summary
		FROM diary_summary_days
		WHERE user_id = $1 AND EXTRACT(YEAR FROM date) = $2 AND EXTRACT(MONTH FROM date) = $3
		ORDER BY date
	`

	rows, err := db.Query(query, userID, year, month)
	if err != nil {
		return fmt.Errorf("failed to get daily summaries: %w", err)
	}
	defer rows.Close()

	var dailySummaries []string
	for rows.Next() {
		var summary string
		if err := rows.Scan(&summary); err != nil {
			return fmt.Errorf("failed to scan daily summary: %w", err)
		}
		dailySummaries = append(dailySummaries, summary)
	}

	if len(dailySummaries) == 0 {
		return fmt.Errorf("no daily summaries found for user %s, year %d, month %d", userID, year, month)
	}

	// 2. LLMで月次要約生成
	combinedDailySummaries := fmt.Sprintf("Daily summaries for %d/%d:\n%s", year, month,
		fmt.Sprintf("- %s", fmt.Sprintf("%s\n", dailySummaries)))
	monthlySummary := generateMonthlySummaryWithLLM(combinedDailySummaries)

	// 3. diary_summary_monthsに保存
	insertQuery := `
		INSERT INTO diary_summary_months (id, user_id, year, month, summary, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, year, month) DO UPDATE SET
		summary = EXCLUDED.summary,
		updated_at = EXCLUDED.updated_at
	`

	now := time.Now().Unix()
	summaryID := uuid.New()

	_, err = db.Exec(insertQuery, summaryID, userID, year, month, monthlySummary, now, now)
	if err != nil {
		return fmt.Errorf("failed to save monthly summary: %w", err)
	}

	log.Printf("Successfully generated and saved monthly summary for user %s, year %d, month %d", userID, year, month)
	return nil
}

func generateSummaryWithLLM(content string) string {
	// TODO: 実際のLLM API（Gemini等）を呼び出してsummaryを生成
	// 現在はモックとして簡単な処理を返す
	log.Printf("Generating daily summary for content (length: %d)", len(content))
	return fmt.Sprintf("Daily summary of diary entry (length: %d characters) - Generated at %s",
		len(content), time.Now().Format("2006-01-02 15:04:05"))
}

func generateMonthlySummaryWithLLM(combinedSummaries string) string {
	// TODO: 実際のLLM API（Gemini等）を呼び出して月次要約を生成
	// 現在はモックとして簡単な処理を返す
	log.Printf("Generating monthly summary for combined summaries (length: %d)", len(combinedSummaries))
	return fmt.Sprintf("Monthly summary based on daily summaries (total length: %d characters) - Generated at %s",
		len(combinedSummaries), time.Now().Format("2006-01-02 15:04:05"))
}