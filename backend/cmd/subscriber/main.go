package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/llm"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/lock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/rueidis"
)

var (
	// Prometheus metrics
	messagesProcessedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "subscriber_messages_processed_total",
			Help: "Total number of messages processed",
		},
		[]string{"message_type", "status"},
	)
	processingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "subscriber_processing_duration_seconds",
			Help:    "Duration of message processing",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"message_type"},
	)
	summariesGeneratedCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "subscriber_summaries_generated_total",
			Help: "Total number of summaries generated",
		},
		[]string{"summary_type"},
	)
	lockOperationsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "subscriber_lock_operations_total",
			Help: "Total number of lock operations",
		},
		[]string{"operation", "status", "lock_type"},
	)
)

func init() {
	prometheus.MustRegister(messagesProcessedCounter)
	prometheus.MustRegister(processingDuration)
	prometheus.MustRegister(summariesGeneratedCounter)
	prometheus.MustRegister(lockOperationsCounter)
}

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

	// Redis接続確認
	ctx := context.Background()
	pingCmd := client.B().Ping().Build()
	if err := client.Do(ctx, pingCmd).Error(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Print("Connected to Redis successfully")

	// メトリクスサーバー開始
	metricsServer := &http.Server{Addr: ":2005"}
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Print("Metrics server starting on :2005")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	log.Print("Subscriber is listening for messages...")

	// Create context for subscription that can be cancelled
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Track processing messages for graceful shutdown
	var wg sync.WaitGroup
	processing := make(chan struct{}, 100) // Buffer to limit concurrent processing

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start subscriber in goroutine
	subErrChan := make(chan error, 1)
	go func() {
		err := client.Receive(subCtx, client.B().Subscribe().Channel("diary_events").Build(), func(msg rueidis.PubSubMessage) {
			log.Printf("Received message: %s from channel: %s", msg.Message, msg.Channel)

			// Track this message processing
			wg.Add(1)
			processing <- struct{}{} // Acquire processing slot

			go func() {
				defer func() {
					<-processing // Release processing slot
					wg.Done()
				}()

				start := time.Now()
				err := processMessage(subCtx, db, client, msg.Message)
				duration := time.Since(start)

				// メトリクス更新は processMessage 内で行う
				_ = duration // 使用しない場合の警告回避

				if err != nil {
					log.Printf("Failed to process message: %v", err)
				}
			}()
		})

		if err != nil {
			subErrChan <- err
		}
	}()

	// Wait for shutdown signal or subscription error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)

		// Cancel subscription context to stop receiving new messages
		cancel()
		log.Print("Stopped accepting new messages")

		// Create context with timeout for graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Wait for all processing messages to complete or timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			log.Print("All messages processed successfully")
		case <-shutdownCtx.Done():
			log.Print("Graceful shutdown timeout, some messages may not have been processed")
		}

		// Stop metrics server
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("Metrics server shutdown error: %v", err)
		} else {
			log.Print("Metrics server stopped")
		}

	case err := <-subErrChan:
		log.Printf("Subscription error: %v", err)
		cancel()
	}

	log.Print("Subscriber ended")
}

func processMessage(ctx context.Context, db *sql.DB, redisClient rueidis.Client, payload string) error {
	start := time.Now()

	// まずメッセージタイプを確認
	var baseMessage struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(payload), &baseMessage); err != nil {
		messagesProcessedCounter.WithLabelValues("unknown", "error").Inc()
		return fmt.Errorf("failed to unmarshal base message: %w", err)
	}

	var err error
	switch baseMessage.Type {
	case "daily_summary":
		processingDuration.WithLabelValues("daily_summary").Observe(time.Since(start).Seconds())
		var message SummaryGenerationMessage
		if unmarshalErr := json.Unmarshal([]byte(payload), &message); unmarshalErr != nil {
			messagesProcessedCounter.WithLabelValues("daily_summary", "error").Inc()
			return fmt.Errorf("failed to unmarshal daily summary message: %w", unmarshalErr)
		}
		err = generateDailySummary(ctx, db, redisClient, message.UserID, message.Date)
		if err != nil {
			messagesProcessedCounter.WithLabelValues("daily_summary", "error").Inc()
		} else {
			messagesProcessedCounter.WithLabelValues("daily_summary", "success").Inc()
		}
		return err
	case "monthly_summary":
		processingDuration.WithLabelValues("monthly_summary").Observe(time.Since(start).Seconds())
		var message MonthlySummaryGenerationMessage
		if unmarshalErr := json.Unmarshal([]byte(payload), &message); unmarshalErr != nil {
			messagesProcessedCounter.WithLabelValues("monthly_summary", "error").Inc()
			return fmt.Errorf("failed to unmarshal monthly summary message: %w", unmarshalErr)
		}
		err = generateMonthlySummary(ctx, db, redisClient, message.UserID, message.Year, message.Month)
		if err != nil {
			messagesProcessedCounter.WithLabelValues("monthly_summary", "error").Inc()
		} else {
			messagesProcessedCounter.WithLabelValues("monthly_summary", "success").Inc()
		}
		return err
	default:
		log.Printf("Unknown message type: %s", baseMessage.Type)
		messagesProcessedCounter.WithLabelValues("unknown", "ignored").Inc()
		return nil
	}
}

func generateDailySummary(ctx context.Context, db *sql.DB, redisClient rueidis.Client, userID, dateStr string) error {
	log.Printf("Generating daily summary for user %s, date %s", userID, dateStr)

	// 1. 分散ロックを取得
	lockKey := lock.DailySummaryLockKey(userID, dateStr)
	distributedLock := lock.NewDistributedLock(redisClient, lockKey, 5*time.Minute)

	locked, err := distributedLock.TryLock(ctx)
	if err != nil {
		lockOperationsCounter.WithLabelValues("acquire", "error", "daily").Inc()
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		// Lock already held by another process, skip processing
		lockOperationsCounter.WithLabelValues("acquire", "failed", "daily").Inc()
		log.Printf("Daily summary for user %s, date %s is already being processed by another instance", userID, dateStr)
		return nil
	}

	lockOperationsCounter.WithLabelValues("acquire", "success", "daily").Inc()
	log.Printf("Acquired lock for daily summary generation: user %s, date %s", userID, dateStr)

	// タスクステータスを「処理中」に更新
	taskKey := fmt.Sprintf("task:daily_summary:%s:%s", userID, dateStr)
	setCmd := redisClient.B().Set().Key(taskKey).Value("processing").Ex(600 * time.Second).Build()
	redisClient.Do(ctx, setCmd)

	// Ensure lock is released when function exits
	defer func() {
		// タスクステータスを削除
		delCmd := redisClient.B().Del().Key(taskKey).Build()
		redisClient.Do(ctx, delCmd)

		if unlockErr := distributedLock.Unlock(ctx); unlockErr != nil {
			lockOperationsCounter.WithLabelValues("release", "error", "daily").Inc()
			log.Printf("Failed to release lock for user %s, date %s: %v", userID, dateStr, unlockErr)
		} else {
			lockOperationsCounter.WithLabelValues("release", "success", "daily").Inc()
			log.Printf("Released lock for daily summary generation: user %s, date %s", userID, dateStr)
		}
	}()

	// 2. 指定された日の日記内容を取得
	var diaryContent string
	query := `SELECT content FROM diaries WHERE user_id = $1 AND date = $2`
	err = db.QueryRow(query, userID, dateStr).Scan(&diaryContent)
	if err != nil {
		return fmt.Errorf("failed to get diary content: %w", err)
	}

	// 2. LLMで要約生成
	summary := generateSummaryWithLLM(ctx, db, userID, diaryContent)

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

	summariesGeneratedCounter.WithLabelValues("daily").Inc()
	log.Printf("Successfully generated and saved summary for user %s, date %s", userID, dateStr)
	return nil
}

func generateMonthlySummary(ctx context.Context, db *sql.DB, redisClient rueidis.Client, userID string, year, month int) error {
	log.Printf("Generating monthly summary for user %s, year %d, month %d", userID, year, month)

	// 1. 分散ロックを取得
	lockKey := lock.MonthlySummaryLockKey(userID, year, month)
	distributedLock := lock.NewDistributedLock(redisClient, lockKey, 5*time.Minute)

	locked, err := distributedLock.TryLock(ctx)
	if err != nil {
		lockOperationsCounter.WithLabelValues("acquire", "error", "monthly").Inc()
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		// Lock already held by another process, skip processing
		lockOperationsCounter.WithLabelValues("acquire", "failed", "monthly").Inc()
		log.Printf("Monthly summary for user %s, year %d, month %d is already being processed by another instance", userID, year, month)
		return nil
	}

	lockOperationsCounter.WithLabelValues("acquire", "success", "monthly").Inc()
	log.Printf("Acquired lock for monthly summary generation: user %s, year %d, month %d", userID, year, month)

	// タスクステータスを「処理中」に更新
	taskKey := fmt.Sprintf("task:monthly_summary:%s:%d-%d", userID, year, month)
	setCmd := redisClient.B().Set().Key(taskKey).Value("processing").Ex(600 * time.Second).Build()
	redisClient.Do(ctx, setCmd)

	// Ensure lock is released when function exits
	defer func() {
		// タスクステータスを削除
		delCmd := redisClient.B().Del().Key(taskKey).Build()
		redisClient.Do(ctx, delCmd)

		if unlockErr := distributedLock.Unlock(ctx); unlockErr != nil {
			lockOperationsCounter.WithLabelValues("release", "error", "monthly").Inc()
			log.Printf("Failed to release lock for user %s, year %d, month %d: %v", userID, year, month, unlockErr)
		} else {
			lockOperationsCounter.WithLabelValues("release", "success", "monthly").Inc()
			log.Printf("Released lock for monthly summary generation: user %s, year %d, month %d", userID, year, month)
		}
	}()

	// 2. 指定された年月の日次要約を全て取得
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
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Failed to close rows: %v", err)
		}
	}()

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
	monthlySummary := generateMonthlySummaryWithLLM(ctx, db, userID, combinedDailySummaries)

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

	summariesGeneratedCounter.WithLabelValues("monthly").Inc()
	log.Printf("Successfully generated and saved monthly summary for user %s, year %d, month %d", userID, year, month)
	return nil
}

func generateSummaryWithLLM(ctx context.Context, db *sql.DB, userID, content string) string {
	// ユーザーのGemini API keyをuser_llmsテーブルから取得
	var apiKey string
	query := `SELECT key FROM user_llms WHERE user_id = $1 AND llm_provider = 1`
	err := db.QueryRow(query, userID).Scan(&apiKey)
	if err != nil {
		log.Printf("Failed to get user's Gemini API key for user %s: %v", userID, err)
		return fmt.Sprintf("Daily summary of diary entry (length: %d characters) - Generated at %s",
			len(content), time.Now().Format("2006-01-02 15:04:05"))
	}

	// Gemini クライアント作成
	geminiClient, err := llm.NewGeminiClient(ctx, apiKey)
	if err != nil {
		log.Printf("Failed to create Gemini client: %v", err)
		return fmt.Sprintf("Daily summary of diary entry (length: %d characters) - Generated at %s",
			len(content), time.Now().Format("2006-01-02 15:04:05"))
	}
	defer func() {
		if closeErr := geminiClient.Close(); closeErr != nil {
			log.Printf("Failed to close Gemini client: %v", closeErr)
		}
	}()

	// 日次要約生成
	summary, err := geminiClient.GenerateDailySummary(ctx, content)
	if err != nil {
		log.Printf("Failed to generate daily summary: %v", err)
		return fmt.Sprintf("Daily summary of diary entry (length: %d characters) - Generated at %s",
			len(content), time.Now().Format("2006-01-02 15:04:05"))
	}

	log.Printf("Successfully generated daily summary using Gemini API")
	return summary
}

func generateMonthlySummaryWithLLM(ctx context.Context, db *sql.DB, userID, combinedSummaries string) string {
	// ユーザーのGemini API keyをuser_llmsテーブルから取得
	var apiKey string
	query := `SELECT key FROM user_llms WHERE user_id = $1 AND llm_provider = 1`
	err := db.QueryRow(query, userID).Scan(&apiKey)
	if err != nil {
		log.Printf("Failed to get user's Gemini API key for user %s: %v", userID, err)
		return fmt.Sprintf("Monthly summary based on daily summaries (total length: %d characters) - Generated at %s",
			len(combinedSummaries), time.Now().Format("2006-01-02 15:04:05"))
	}

	// Gemini クライアント作成
	geminiClient, err := llm.NewGeminiClient(ctx, apiKey)
	if err != nil {
		log.Printf("Failed to create Gemini client: %v", err)
		return fmt.Sprintf("Monthly summary based on daily summaries (total length: %d characters) - Generated at %s",
			len(combinedSummaries), time.Now().Format("2006-01-02 15:04:05"))
	}
	defer func() {
		if closeErr := geminiClient.Close(); closeErr != nil {
			log.Printf("Failed to close Gemini client: %v", closeErr)
		}
	}()

	// 月次要約生成
	summary, err := geminiClient.GenerateSummary(ctx, combinedSummaries)
	if err != nil {
		log.Printf("Failed to generate monthly summary: %v", err)
		return fmt.Sprintf("Monthly summary based on daily summaries (total length: %d characters) - Generated at %s",
			len(combinedSummaries), time.Now().Format("2006-01-02 15:04:05"))
	}

	log.Printf("Successfully generated monthly summary using Gemini API")
	return summary
}
