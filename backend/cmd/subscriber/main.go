package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/container"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
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
	connectionReconnectsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "subscriber_connection_reconnects_total",
			Help: "Total number of Redis connection reconnects",
		},
		[]string{"status"},
	)
	connectionStatusGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "subscriber_connection_status",
			Help: "Current Redis connection status (1=connected, 0=disconnected)",
		},
		[]string{"connection_type"},
	)
)

func init() {
	prometheus.MustRegister(messagesProcessedCounter)
	prometheus.MustRegister(processingDuration)
	prometheus.MustRegister(summariesGeneratedCounter)
	prometheus.MustRegister(lockOperationsCounter)
	prometheus.MustRegister(connectionReconnectsCounter)
	prometheus.MustRegister(connectionStatusGauge)
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

type LatestTrendGenerationMessage struct {
	Type        string `json:"type"`
	UserID      string `json:"user_id"`
	PeriodStart string `json:"period_start"` // ISO 8601 format
	PeriodEnd   string `json:"period_end"`   // ISO 8601 format
}

func main() {
	// Initialize structured logger
	logger := logrus.WithFields(logrus.Fields{
		"service": "subscriber",
	})
	logger.Info("=== umi.mikan subscriber started ===")

	// Create DI container
	diContainer, err := container.NewContainer()
	if err != nil {
		logger.WithError(err).Fatal("Failed to create DI container")
	}

	// Initialize and run subscriber using DI container
	if err := diContainer.Invoke(func(app *container.SubscriberApp, cleanup *container.Cleanup) error {
		return runSubscriber(app, cleanup, logger)
	}); err != nil {
		logger.WithError(err).Fatal("Failed to start subscriber")
	}
}

func runSubscriber(app *container.SubscriberApp, cleanup *container.Cleanup, logger *logrus.Entry) error {
	// Redis接続確認
	ctx := context.Background()
	pingCmd := app.Redis.B().Ping().Build()
	if err := app.Redis.Do(ctx, pingCmd).Error(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	logger.Info("Connected to Redis successfully")

	// メトリクスサーバー開始
	metricsServer := &http.Server{Addr: ":2005"}
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		logger.Info("Metrics server starting on :2005")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Error("Metrics server error")
		}
	}()

	logger.WithField("max_concurrent_jobs", app.SubscriberConfig.MaxConcurrentJobs).Info("Subscriber is listening for messages...")

	// Create context for subscription that can be cancelled
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Initialize connection status
	connectionStatusGauge.WithLabelValues("pubsub").Set(0)

	// Track processing messages for graceful shutdown
	var wg sync.WaitGroup
	processing := make(chan struct{}, app.SubscriberConfig.MaxConcurrentJobs) // Buffer to limit concurrent processing

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start connection health monitor
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Ping Redis to check connection health
				pingCmd := app.Redis.B().Ping().Build()
				if err := app.Redis.Do(ctx, pingCmd).Error(); err != nil {
					logger.WithError(err).Warn("Redis ping failed, connection may be unhealthy")
					connectionStatusGauge.WithLabelValues("ping").Set(0)
				} else {
					connectionStatusGauge.WithLabelValues("ping").Set(1)
				}
			case <-subCtx.Done():
				logger.Info("Connection health monitor stopping")
				return
			}
		}
	}()

	// Start subscriber with automatic reconnection
	subErrChan := make(chan error, 1)
	go func() {
		// Infinite loop for automatic reconnection
		for {
			select {
			case <-subCtx.Done():
				logger.Info("Subscription context cancelled, stopping subscriber")
				return
			default:
				// Attempt to subscribe
				logger.Info("Starting Redis Pub/Sub subscription...")
				connectionReconnectsCounter.WithLabelValues("attempt").Inc()
				err := app.Redis.Receive(subCtx, app.Redis.B().Subscribe().Channel("diary_events").Build(), func(msg rueidis.PubSubMessage) {
					// Mark connection as active when receiving messages
					connectionStatusGauge.WithLabelValues("pubsub").Set(1)
					logger.WithFields(logrus.Fields{
						"channel": msg.Channel,
						"message": msg.Message,
					}).Debug("Received message")

					// Track this message processing
					wg.Add(1)
					processing <- struct{}{} // Acquire processing slot

					go func() {
						defer func() {
							<-processing // Release processing slot
							wg.Done()
						}()

						start := time.Now()
						err := processMessage(subCtx, app.DB, app.Redis, app.LLMFactory, app.LockService, msg.Message, logger)
						duration := time.Since(start)

						// メトリクス更新は processMessage 内で行う
						_ = duration // 使用しない場合の警告回避

						if err != nil {
							logger.WithError(err).Error("Failed to process message")
						}
					}()
				})

				if err != nil {
					// Mark connection as lost
					connectionStatusGauge.WithLabelValues("pubsub").Set(0)

					if subCtx.Err() != nil {
						// Context was cancelled, stop retrying
						logger.Info("Subscription context cancelled during connection")
						return
					}

					logger.WithError(err).Error("Redis Pub/Sub connection lost, attempting to reconnect...")
					connectionReconnectsCounter.WithLabelValues("failed").Inc()

					// Wait before attempting to reconnect
					select {
					case <-time.After(5 * time.Second):
						// Continue to retry
						continue
					case <-subCtx.Done():
						logger.Info("Subscription context cancelled during reconnection wait")
						return
					}
				} else {
					// Successful connection established
					connectionReconnectsCounter.WithLabelValues("success").Inc()
					logger.Info("Redis Pub/Sub subscription established successfully")
				}
			}
		}
	}()

	// Wait for shutdown signal or subscription error
	select {
	case sig := <-sigChan:
		logger.WithField("signal", sig).Info("Received signal, initiating graceful shutdown...")

		// Cancel subscription context to stop receiving new messages
		cancel()
		logger.Info("Stopped accepting new messages")

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
			logger.Info("All messages processed successfully")
		case <-shutdownCtx.Done():
			logger.Warn("Graceful shutdown timeout, some messages may not have been processed")
		}

		// Stop metrics server
		if err := metricsServer.Shutdown(shutdownCtx); err != nil {
			logger.WithError(err).Error("Metrics server shutdown error")
		} else {
			logger.Info("Metrics server stopped")
		}

		// Cleanup resources
		if err := cleanup.Close(); err != nil {
			logger.WithError(err).Error("Error during cleanup")
			return err
		}

	case err := <-subErrChan:
		logger.WithError(err).Error("Subscription error")
		cancel()
		// Cleanup resources on error
		if cleanupErr := cleanup.Close(); cleanupErr != nil {
			logger.WithError(cleanupErr).Error("Error during cleanup")
		}
		return err
	}

	logger.Info("Subscriber ended")
	return nil
}

func processMessage(ctx context.Context, db database.DB, redisClient rueidis.Client, llmFactory container.LLMClientFactory, lockService container.LockService, payload string, logger *logrus.Entry) error {
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
		err = generateDailySummary(ctx, db, redisClient, llmFactory, lockService, message.UserID, message.Date, logger)
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
		err = generateMonthlySummary(ctx, db, redisClient, llmFactory, lockService, message.UserID, message.Year, message.Month, logger)
		if err != nil {
			messagesProcessedCounter.WithLabelValues("monthly_summary", "error").Inc()
		} else {
			messagesProcessedCounter.WithLabelValues("monthly_summary", "success").Inc()
		}
		return err
	case "latest_trend":
		processingDuration.WithLabelValues("latest_trend").Observe(time.Since(start).Seconds())
		var message LatestTrendGenerationMessage
		if unmarshalErr := json.Unmarshal([]byte(payload), &message); unmarshalErr != nil {
			messagesProcessedCounter.WithLabelValues("latest_trend", "error").Inc()
			return fmt.Errorf("failed to unmarshal latest trend message: %w", unmarshalErr)
		}
		err = generateLatestTrend(ctx, db, redisClient, llmFactory, lockService, message.UserID, message.PeriodStart, message.PeriodEnd, logger)
		if err != nil {
			messagesProcessedCounter.WithLabelValues("latest_trend", "error").Inc()
		} else {
			messagesProcessedCounter.WithLabelValues("latest_trend", "success").Inc()
		}
		return err
	default:
		logger.WithField("message_type", baseMessage.Type).Warn("Unknown message type")
		messagesProcessedCounter.WithLabelValues("unknown", "ignored").Inc()
		return nil
	}
}

func generateDailySummary(ctx context.Context, db database.DB, redisClient rueidis.Client, llmFactory container.LLMClientFactory, lockService container.LockService, userID, dateStr string, logger *logrus.Entry) error {
	logger.WithFields(logrus.Fields{
		"user_id": userID,
		"date":    dateStr,
	}).Info("Generating daily summary")

	// 1. 分散ロックを取得
	lockKey := fmt.Sprintf("summary_lock:daily:%s:%s", userID, dateStr)
	distributedLock := lockService.NewDistributedLock(lockKey, 5*time.Minute)

	locked, err := distributedLock.TryLock(ctx)
	if err != nil {
		lockOperationsCounter.WithLabelValues("acquire", "error", "daily").Inc()
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		// Lock already held by another process, skip processing
		lockOperationsCounter.WithLabelValues("acquire", "failed", "daily").Inc()
		logger.WithFields(logrus.Fields{
			"user_id": userID,
			"date":    dateStr,
		}).Info("Daily summary is already being processed by another instance, skipping")
		return nil
	}

	lockOperationsCounter.WithLabelValues("acquire", "success", "daily").Inc()
	logger.WithFields(logrus.Fields{
		"user_id": userID,
		"date":    dateStr,
	}).Debug("Acquired lock for daily summary generation")

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
			logger.WithError(unlockErr).WithFields(logrus.Fields{
				"user_id": userID,
				"date":    dateStr,
			}).Error("Failed to release lock")
		} else {
			lockOperationsCounter.WithLabelValues("release", "success", "daily").Inc()
			logger.WithFields(logrus.Fields{
				"user_id": userID,
				"date":    dateStr,
			}).Debug("Released lock for daily summary generation")
		}
	}()

	// 2. 指定された日の日記内容を取得
	var diaryContent string
	query := `SELECT content FROM diaries WHERE user_id = $1 AND date = $2`
	err = db.QueryRowContext(ctx, query, userID, dateStr).Scan(&diaryContent)
	if err != nil {
		return fmt.Errorf("failed to get diary content: %w", err)
	}

	// 2. LLMで要約生成
	summary := generateSummaryWithLLM(ctx, db, llmFactory, userID, diaryContent, logger)

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

	_, err = db.ExecContext(ctx, insertQuery, summaryID, userID, dateStr, summary, now, now)
	if err != nil {
		return fmt.Errorf("failed to save summary: %w", err)
	}

	summariesGeneratedCounter.WithLabelValues("daily").Inc()
	logger.WithFields(logrus.Fields{
		"user_id": userID,
		"date":    dateStr,
	}).Info("Successfully generated and saved daily summary")
	return nil
}

func generateMonthlySummary(ctx context.Context, db database.DB, redisClient rueidis.Client, llmFactory container.LLMClientFactory, lockService container.LockService, userID string, year, month int, logger *logrus.Entry) error {
	logger.WithFields(logrus.Fields{
		"user_id": userID,
		"year":    year,
		"month":   month,
	}).Info("Generating monthly summary")

	// 1. 分散ロックを取得
	lockKey := fmt.Sprintf("summary_lock:monthly:%s:%d:%d", userID, year, month)
	distributedLock := lockService.NewDistributedLock(lockKey, 5*time.Minute)

	locked, err := distributedLock.TryLock(ctx)
	if err != nil {
		lockOperationsCounter.WithLabelValues("acquire", "error", "monthly").Inc()
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		// Lock already held by another process, skip processing
		lockOperationsCounter.WithLabelValues("acquire", "failed", "monthly").Inc()
		logger.WithFields(logrus.Fields{"user_id": userID, "year": year, "month": month}).Info("Monthly summary is already being processed by another instance, skipping")
		return nil
	}

	lockOperationsCounter.WithLabelValues("acquire", "success", "monthly").Inc()
	logger.WithFields(logrus.Fields{"user_id": userID, "year": year, "month": month}).Debug("Acquired lock for monthly summary generation")

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
			logger.WithError(unlockErr).WithFields(logrus.Fields{"user_id": userID, "year": year, "month": month}).Error("Failed to release lock")
		} else {
			lockOperationsCounter.WithLabelValues("release", "success", "monthly").Inc()
			logger.WithFields(logrus.Fields{"user_id": userID, "year": year, "month": month}).Debug("Released lock for monthly summary generation")
		}
	}()

	// 2. 指定された年月の日記エントリーを全て取得
	query := `
		SELECT date, content
		FROM diaries
		WHERE user_id = $1 AND EXTRACT(YEAR FROM date) = $2 AND EXTRACT(MONTH FROM date) = $3
		ORDER BY date
	`

	rows, err := db.QueryContext(ctx, query, userID, year, month)
	if err != nil {
		return fmt.Errorf("failed to get diary entries: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.WithError(err).Error("Failed to close rows")
		}
	}()

	var diaryEntries []string
	for rows.Next() {
		var date, content string
		if err := rows.Scan(&date, &content); err != nil {
			return fmt.Errorf("failed to scan diary entry: %w", err)
		}
		diaryEntries = append(diaryEntries, fmt.Sprintf("[%s]\n%s", date, content))
	}

	if len(diaryEntries) == 0 {
		return fmt.Errorf("no diary entries found for user %s, year %d, month %d", userID, year, month)
	}

	// 2. LLMで月次要約生成
	combinedDiaryEntries := fmt.Sprintf("Diary entries for %d/%d:\n\n%s", year, month,
		strings.Join(diaryEntries, "\n\n"))
	monthlySummary := generateMonthlySummaryWithLLM(ctx, db, llmFactory, userID, combinedDiaryEntries, logger)

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

	_, err = db.ExecContext(ctx, insertQuery, summaryID, userID, year, month, monthlySummary, now, now)
	if err != nil {
		return fmt.Errorf("failed to save monthly summary: %w", err)
	}

	summariesGeneratedCounter.WithLabelValues("monthly").Inc()
	logger.WithFields(logrus.Fields{"user_id": userID, "year": year, "month": month}).Info("Successfully generated and saved monthly summary")
	return nil
}

func generateSummaryWithLLM(ctx context.Context, db database.DB, llmFactory container.LLMClientFactory, userID, content string, logger *logrus.Entry) string {
	// ユーザーのGemini API keyをuser_llmsテーブルから取得
	var apiKey string
	query := `SELECT key FROM user_llms WHERE user_id = $1 AND llm_provider = 1`
	err := db.QueryRowContext(ctx, query, userID).Scan(&apiKey)
	if err != nil {
		logger.WithError(err).WithField("user_id", userID).Error("Failed to get user's Gemini API key")
		return fmt.Sprintf("Daily summary of diary entry (length: %d characters) - Generated at %s",
			len(content), time.Now().Format("2006-01-02 15:04:05"))
	}

	// Gemini クライアント作成
	geminiClient, err := llmFactory.CreateGeminiClient(ctx, apiKey)
	if err != nil {
		logger.WithError(err).Error("Failed to create Gemini client")
		return fmt.Sprintf("Daily summary of diary entry (length: %d characters) - Generated at %s",
			len(content), time.Now().Format("2006-01-02 15:04:05"))
	}
	defer func() {
		if closeErr := geminiClient.Close(); closeErr != nil {
			logger.WithError(closeErr).Error("Failed to close Gemini client")
		}
	}()

	// 日次要約生成
	summary, err := geminiClient.GenerateDailySummary(ctx, content)
	if err != nil {
		logger.WithError(err).Error("Failed to generate daily summary")
		return fmt.Sprintf("Daily summary of diary entry (length: %d characters) - Generated at %s",
			len(content), time.Now().Format("2006-01-02 15:04:05"))
	}

	logger.WithError(err).Error("Successfully generated daily summary using Gemini API")
	return summary
}

func generateMonthlySummaryWithLLM(ctx context.Context, db database.DB, llmFactory container.LLMClientFactory, userID, combinedEntries string, logger *logrus.Entry) string {
	// ユーザーのGemini API keyをuser_llmsテーブルから取得
	var apiKey string
	query := `SELECT key FROM user_llms WHERE user_id = $1 AND llm_provider = 1`
	err := db.QueryRowContext(ctx, query, userID).Scan(&apiKey)
	if err != nil {
		logger.WithError(err).WithField("user_id", userID).Error("Failed to get user's Gemini API key")
		return fmt.Sprintf("Monthly summary based on diary entries (total length: %d characters) - Generated at %s",
			len(combinedEntries), time.Now().Format("2006-01-02 15:04:05"))
	}

	// Gemini クライアント作成
	geminiClient, err := llmFactory.CreateGeminiClient(ctx, apiKey)
	if err != nil {
		logger.WithError(err).Error("Failed to create Gemini client")
		return fmt.Sprintf("Monthly summary based on diary entries (total length: %d characters) - Generated at %s",
			len(combinedEntries), time.Now().Format("2006-01-02 15:04:05"))
	}
	defer func() {
		if closeErr := geminiClient.Close(); closeErr != nil {
			logger.WithError(closeErr).Error("Failed to close Gemini client")
		}
	}()

	// 月次要約生成
	summary, err := geminiClient.GenerateSummary(ctx, combinedEntries)
	if err != nil {
		logger.WithError(err).Error("Failed to generate monthly summary")
		return fmt.Sprintf("Monthly summary based on diary entries (total length: %d characters) - Generated at %s",
			len(combinedEntries), time.Now().Format("2006-01-02 15:04:05"))
	}

	logger.WithError(err).Error("Successfully generated monthly summary using Gemini API")
	return summary
}

func generateLatestTrend(ctx context.Context, db database.DB, redisClient rueidis.Client, llmFactory container.LLMClientFactory, lockService container.LockService, userID, periodStartStr, periodEndStr string, logger *logrus.Entry) error {
	logger.WithFields(logrus.Fields{
		"user_id":      userID,
		"period_start": periodStartStr,
		"period_end":   periodEndStr,
	}).Info("Generating latest trend analysis")

	// 1. 分散ロックを取得
	lockKey := fmt.Sprintf("trend_lock:latest:%s", userID)
	distributedLock := lockService.NewDistributedLock(lockKey, 5*time.Minute)

	locked, err := distributedLock.TryLock(ctx)
	if err != nil {
		lockOperationsCounter.WithLabelValues("acquire", "error", "latest_trend").Inc()
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		// Lock already held by another process, skip processing
		lockOperationsCounter.WithLabelValues("acquire", "failed", "latest_trend").Inc()
		logger.WithFields(logrus.Fields{
			"user_id": userID,
		}).Info("Latest trend is already being processed by another instance, skipping")
		return nil
	}

	lockOperationsCounter.WithLabelValues("acquire", "success", "latest_trend").Inc()
	logger.WithFields(logrus.Fields{
		"user_id": userID,
	}).Debug("Acquired lock for latest trend generation")

	// Ensure lock is released when function exits
	defer func() {
		if unlockErr := distributedLock.Unlock(ctx); unlockErr != nil {
			lockOperationsCounter.WithLabelValues("release", "error", "latest_trend").Inc()
			logger.WithError(unlockErr).WithFields(logrus.Fields{
				"user_id": userID,
			}).Error("Failed to release lock")
		} else {
			lockOperationsCounter.WithLabelValues("release", "success", "latest_trend").Inc()
			logger.WithFields(logrus.Fields{
				"user_id": userID,
			}).Debug("Released lock for latest trend generation")
		}
	}()

	// 2. 期間をパース
	periodStart, err := time.Parse(time.RFC3339, periodStartStr)
	if err != nil {
		return fmt.Errorf("failed to parse period_start: %w", err)
	}
	periodEnd, err := time.Parse(time.RFC3339, periodEndStr)
	if err != nil {
		return fmt.Errorf("failed to parse period_end: %w", err)
	}

	// 3. 指定期間の日記エントリーを取得
	query := `
		SELECT date, content
		FROM diaries
		WHERE user_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date
	`

	rows, err := db.QueryContext(ctx, query, userID, periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("failed to get diary entries: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.WithError(err).Error("Failed to close rows")
		}
	}()

	var diaryEntries []string
	for rows.Next() {
		var date, content string
		if err := rows.Scan(&date, &content); err != nil {
			return fmt.Errorf("failed to scan diary entry: %w", err)
		}
		diaryEntries = append(diaryEntries, fmt.Sprintf("[%s]\n%s", date, content))
	}

	if len(diaryEntries) < 2 {
		logger.WithFields(logrus.Fields{
			"user_id":      userID,
			"entry_count":  len(diaryEntries),
		}).Info("Not enough diary entries for latest trend analysis (need at least 2 days)")
		return fmt.Errorf("not enough diary entries (found %d, need at least 2)", len(diaryEntries))
	}

	// 4. LLMでトレンド分析生成
	combinedDiaryEntries := fmt.Sprintf("Diary entries from %s to %s:\n\n%s", periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"),
		strings.Join(diaryEntries, "\n\n"))
	trendAnalysis := generateLatestTrendWithLLM(ctx, db, llmFactory, userID, combinedDiaryEntries, logger)

	// 5. Redisに保存（TTL: 2日）
	trendData := map[string]interface{}{
		"user_id":      userID,
		"analysis":     trendAnalysis,
		"period_start": periodStartStr,
		"period_end":   periodEndStr,
		"generated_at": time.Now().Format(time.RFC3339),
	}

	trendDataJSON, err := json.Marshal(trendData)
	if err != nil {
		return fmt.Errorf("failed to marshal trend data: %w", err)
	}

	trendKey := fmt.Sprintf("latest_trend:%s", userID)
	setCmd := redisClient.B().Set().Key(trendKey).Value(string(trendDataJSON)).Ex(172800 * time.Second).Build() // 2日間
	if err := redisClient.Do(ctx, setCmd).Error(); err != nil {
		return fmt.Errorf("failed to save trend data to Redis: %w", err)
	}

	summariesGeneratedCounter.WithLabelValues("latest_trend").Inc()
	logger.WithFields(logrus.Fields{
		"user_id": userID,
	}).Info("Successfully generated and saved latest trend analysis")
	return nil
}

func generateLatestTrendWithLLM(ctx context.Context, db database.DB, llmFactory container.LLMClientFactory, userID, combinedEntries string, logger *logrus.Entry) string {
	// ユーザーのGemini API keyをuser_llmsテーブルから取得
	var apiKey string
	query := `SELECT key FROM user_llms WHERE user_id = $1 AND llm_provider = 1`
	err := db.QueryRowContext(ctx, query, userID).Scan(&apiKey)
	if err != nil {
		logger.WithError(err).WithField("user_id", userID).Error("Failed to get user's Gemini API key")
		return fmt.Sprintf("Latest trend analysis based on diary entries (total length: %d characters) - Generated at %s",
			len(combinedEntries), time.Now().Format("2006-01-02 15:04:05"))
	}

	// Gemini クライアント作成
	geminiClient, err := llmFactory.CreateGeminiClient(ctx, apiKey)
	if err != nil {
		logger.WithError(err).Error("Failed to create Gemini client")
		return fmt.Sprintf("Latest trend analysis based on diary entries (total length: %d characters) - Generated at %s",
			len(combinedEntries), time.Now().Format("2006-01-02 15:04:05"))
	}
	defer func() {
		if closeErr := geminiClient.Close(); closeErr != nil {
			logger.WithError(closeErr).Error("Failed to close Gemini client")
		}
	}()

	// トレンド分析生成
	analysis, err := geminiClient.GenerateLatestTrend(ctx, combinedEntries)
	if err != nil {
		logger.WithError(err).Error("Failed to generate latest trend analysis")
		return fmt.Sprintf("Latest trend analysis based on diary entries (total length: %d characters) - Generated at %s",
			len(combinedEntries), time.Now().Format("2006-01-02 15:04:05"))
	}

	logger.Info("Successfully generated latest trend analysis using Gemini API")
	return analysis
}
