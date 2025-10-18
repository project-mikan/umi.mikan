package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/project-mikan/umi.mikan/backend/container"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
)

var (
	// Prometheus metrics
	jobExecutionCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scheduler_job_executions_total",
			Help: "Total number of job executions",
		},
		[]string{"job_name", "status"},
	)
	jobDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "scheduler_job_duration_seconds",
			Help:    "Duration of job executions",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"job_name"},
	)
	queuedMessagesCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "scheduler_queued_messages_total",
			Help: "Total number of messages queued to Redis",
		},
		[]string{"message_type"},
	)
	usersWithAutoSummaryGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "scheduler_users_with_auto_summary",
			Help: "Number of users with auto summary enabled",
		},
		[]string{"summary_type"},
	)
)

func init() {
	prometheus.MustRegister(jobExecutionCounter)
	prometheus.MustRegister(jobDuration)
	prometheus.MustRegister(queuedMessagesCounter)
	prometheus.MustRegister(usersWithAutoSummaryGauge)
}

// Scheduler types and functions
type Scheduler struct {
	db     database.DB
	redis  rueidis.Client
	ctx    context.Context
	cancel context.CancelFunc
	logger *logrus.Entry
}

type ScheduledJob interface {
	Name() string
	Interval() time.Duration
	Execute(ctx context.Context, s *Scheduler) error
}

func NewScheduler(app *container.SchedulerApp, logger *logrus.Entry) (*Scheduler, error) {
	ctx, cancel := context.WithCancel(context.Background())

	return &Scheduler{
		db:     app.DB,
		redis:  app.Redis,
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}, nil
}

func (s *Scheduler) AddJob(job ScheduledJob) {
	go func() {
		ticker := time.NewTicker(job.Interval())
		defer ticker.Stop()

		s.logger.WithFields(logrus.Fields{
			"job_name": job.Name(),
			"interval": job.Interval(),
		}).Info("Scheduled job started")

		for {
			select {
			case <-s.ctx.Done():
				s.logger.WithField("job_name", job.Name()).Info("Scheduled job stopped")
				return
			case <-ticker.C:
				s.logger.WithField("job_name", job.Name()).Debug("Executing job")

				// Metrics tracking
				start := time.Now()
				err := job.Execute(s.ctx, s)
				duration := time.Since(start)

				jobDuration.WithLabelValues(job.Name()).Observe(duration.Seconds())

				if err != nil {
					s.logger.WithError(err).WithFields(logrus.Fields{
						"job_name": job.Name(),
						"duration": duration,
					}).Error("Error executing job")
					jobExecutionCounter.WithLabelValues(job.Name(), "error").Inc()
				} else {
					s.logger.WithFields(logrus.Fields{
						"job_name": job.Name(),
						"duration": duration,
					}).Debug("Job executed successfully")
					jobExecutionCounter.WithLabelValues(job.Name(), "success").Inc()
				}
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.cancel()
}

func main() {
	// Initialize structured logger
	logger := logrus.WithFields(logrus.Fields{
		"service": "scheduler",
	})
	logger.Info("=== umi.mikan scheduler started ===")

	// Create DI container
	diContainer, err := container.NewContainer()
	if err != nil {
		logger.WithError(err).Fatal("Failed to create DI container")
	}

	// Initialize and run scheduler using DI container
	if err := diContainer.Invoke(func(app *container.SchedulerApp, cleanup *container.Cleanup) error {
		return runScheduler(app, cleanup, logger)
	}); err != nil {
		logger.WithError(err).Fatal("Failed to start scheduler")
	}
}

func runScheduler(app *container.SchedulerApp, cleanup *container.Cleanup, logger *logrus.Entry) error {
	// Redis接続確認
	ctx := context.Background()
	pingCmd := app.Redis.B().Ping().Build()
	if err := app.Redis.Do(ctx, pingCmd).Error(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	logger.Info("Connected to Redis successfully")

	// メトリクスサーバー開始
	metricsServer := &http.Server{Addr: ":2006"}
	http.Handle("/metrics", promhttp.Handler())

	go func() {
		logger.Info("Metrics server starting on :2006")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Error("Metrics server error")
		}
	}()

	// スケジューラー作成
	scheduler, err := NewScheduler(app, logger)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	// ジョブを追加
	scheduler.AddJob(NewDailySummaryJob(app.SchedulerConfig.DailySummaryInterval))
	scheduler.AddJob(NewMonthlySummaryJob(app.SchedulerConfig.MonthlySummaryInterval))
	scheduler.AddJob(NewLatestTrendJob(app.SchedulerConfig.LatestTrendInterval))

	logger.Info("Scheduler is running...")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.WithField("signal", sig).Info("Received signal, initiating graceful shutdown...")

	// Create context with timeout for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop scheduler
	scheduler.Stop()
	logger.Info("Scheduler stopped")

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

	return nil
}

// DailySummaryJob handles daily summary generation
type DailySummaryJob struct {
	interval time.Duration
}

func NewDailySummaryJob(interval time.Duration) *DailySummaryJob {
	return &DailySummaryJob{interval: interval}
}

func (j *DailySummaryJob) Name() string {
	return "DailySummaryGeneration"
}

func (j *DailySummaryJob) Interval() time.Duration {
	return j.interval
}

func (j *DailySummaryJob) Execute(ctx context.Context, s *Scheduler) error {
	s.logger.Info("Checking for missing daily summaries...")

	// 1. auto_summary_daily が true のユーザーを取得
	usersQuery := `
		SELECT user_id
		FROM user_llms
		WHERE auto_summary_daily = true
	`

	rows, err := s.db.QueryContext(ctx, usersQuery)
	if err != nil {
		return fmt.Errorf("failed to query users with auto summary enabled: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close rows")
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
		s.logger.Info("No users with auto daily summary enabled")
		return nil
	}

	s.logger.WithField("count", len(userIDs)).Info("Found users with auto daily summary enabled")
	usersWithAutoSummaryGauge.WithLabelValues("daily").Set(float64(len(userIDs)))

	// 2. 各ユーザーについて、summaryが作られていない日を確認
	for _, userID := range userIDs {
		if err := j.processUserSummaries(ctx, s, userID); err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Error processing summaries for user")
			continue
		}
	}

	return nil
}

func (j *DailySummaryJob) processUserSummaries(ctx context.Context, s *Scheduler, userID string) error {
	// diariesテーブルから該当ユーザーの日記がある日を取得し、
	// diary_summary_daysにsummaryがない日、または要約のupdated_atが日記のupdated_atより古い日を見つける（今日を除く）
	// 文字数が1000以上の日記のみ対象とする
	query := `
		SELECT d.date
		FROM diaries d
		LEFT JOIN diary_summary_days dsd ON d.user_id = dsd.user_id AND d.date = dsd.date
		WHERE d.user_id = $1
		  AND d.date < CURRENT_DATE
		  AND LENGTH(d.content) >= 1000
		  AND (dsd.id IS NULL OR dsd.updated_at < d.updated_at)
		ORDER BY d.date
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to query missing summaries for user %s: %w", userID, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close rows")
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
		s.logger.WithField("user_id", userID).Debug("No missing summaries for user")
		return nil
	}

	s.logger.WithFields(map[string]any{"user_id": userID, "count": len(missingDates)}).Info("Found missing summaries for user")

	// 3. 各日付についてRedisキューにジョブを投入
	for _, date := range missingDates {
		message := map[string]any{
			"type":    "daily_summary",
			"user_id": userID,
			"date":    date.Format("2006-01-02"),
		}

		messageBytes, err := json.Marshal(message)
		if err != nil {
			s.logger.WithError(err).WithFields(map[string]any{"user_id": userID, "date": date.Format("2006-01-02")}).Error("Failed to marshal message")
			continue
		}

		// Redisにメッセージを送信
		publishCmd := s.redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
		if err := s.redis.Do(ctx, publishCmd).Error(); err != nil {
			s.logger.WithError(err).WithFields(map[string]any{"user_id": userID, "date": date.Format("2006-01-02")}).Error("Failed to publish message")
			continue
		}

		queuedMessagesCounter.WithLabelValues("daily_summary").Inc()
		s.logger.WithFields(map[string]any{"user_id": userID, "date": date.Format("2006-01-02")}).Debug("Queued summary generation")
	}

	return nil
}

// MonthlySummaryJob handles monthly summary generation
type MonthlySummaryJob struct {
	interval time.Duration
}

func NewMonthlySummaryJob(interval time.Duration) *MonthlySummaryJob {
	return &MonthlySummaryJob{interval: interval}
}

func (j *MonthlySummaryJob) Name() string {
	return "MonthlySummaryGeneration"
}

func (j *MonthlySummaryJob) Interval() time.Duration {
	return j.interval
}

func (j *MonthlySummaryJob) Execute(ctx context.Context, s *Scheduler) error {
	s.logger.Info("Checking for missing monthly summaries...")

	// 1. auto_summary_monthly が true のユーザーを取得
	usersQuery := `
		SELECT user_id
		FROM user_llms
		WHERE auto_summary_monthly = true
	`

	rows, err := s.db.QueryContext(ctx, usersQuery)
	if err != nil {
		return fmt.Errorf("failed to query users with auto monthly summary enabled: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close rows")
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
		s.logger.Info("No users with auto monthly summary enabled")
		return nil
	}

	s.logger.WithField("count", len(userIDs)).Info("Found users with auto monthly summary enabled")
	usersWithAutoSummaryGauge.WithLabelValues("monthly").Set(float64(len(userIDs)))

	// 2. 各ユーザーについて、summaryが作られていない月を確認
	for _, userID := range userIDs {
		if err := j.processUserMonthlySummaries(ctx, s, userID); err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Error processing monthly summaries for user")
			continue
		}
	}

	return nil
}

func (j *MonthlySummaryJob) processUserMonthlySummaries(ctx context.Context, s *Scheduler, userID string) error {
	// diariesテーブルから該当ユーザーの日記がある年月を取得し、
	// diary_summary_monthsに月次要約がない月、またはその月の日記の最新updated_atより月次要約のupdated_atが古い月を見つける（今月を除く）
	// 日記数が1以上の月のみ対象とする
	query := `
		WITH monthly_diary_stats AS (
			SELECT
				EXTRACT(YEAR FROM d.date) as year,
				EXTRACT(MONTH FROM d.date) as month,
				MAX(d.updated_at) as latest_diary_updated_at,
				COUNT(*) as diary_count
			FROM diaries d
			WHERE d.user_id = $1
			GROUP BY EXTRACT(YEAR FROM d.date), EXTRACT(MONTH FROM d.date)
			HAVING COUNT(*) >= 1
		)
		SELECT mds.year, mds.month
		FROM monthly_diary_stats mds
		LEFT JOIN diary_summary_months dsm ON dsm.user_id = $1
			AND dsm.year = mds.year
			AND dsm.month = mds.month
		WHERE (mds.year < EXTRACT(YEAR FROM CURRENT_DATE)
			OR (mds.year = EXTRACT(YEAR FROM CURRENT_DATE) AND mds.month < EXTRACT(MONTH FROM CURRENT_DATE)))
		AND (dsm.updated_at IS NULL OR dsm.updated_at < mds.latest_diary_updated_at)
		ORDER BY mds.year, mds.month
	`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to query missing monthly summaries for user %s: %w", userID, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close rows")
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
		s.logger.WithField("user_id", userID).Debug("No missing monthly summaries for user")
		return nil
	}

	s.logger.WithFields(map[string]any{"user_id": userID, "count": len(missingMonths)}).Info("Found missing monthly summaries for user")

	// 3. 各年月についてRedisキューにジョブを投入
	for _, ym := range missingMonths {
		message := map[string]any{
			"type":    "monthly_summary",
			"user_id": userID,
			"year":    ym.Year,
			"month":   ym.Month,
		}

		messageBytes, err := json.Marshal(message)
		if err != nil {
			s.logger.WithError(err).WithFields(map[string]any{"user_id": userID, "year": ym.Year, "month": ym.Month}).Error("Failed to marshal message")
			continue
		}

		// Redisにメッセージを送信
		publishCmd := s.redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
		if err := s.redis.Do(ctx, publishCmd).Error(); err != nil {
			s.logger.WithError(err).WithFields(map[string]any{"user_id": userID, "year": ym.Year, "month": ym.Month}).Error("Failed to publish message")
			continue
		}

		queuedMessagesCounter.WithLabelValues("monthly_summary").Inc()
		s.logger.WithFields(map[string]any{"user_id": userID, "year": ym.Year, "month": ym.Month}).Debug("Queued monthly summary generation")
	}

	return nil
}

// LatestTrendJob handles latest trend analysis generation
type LatestTrendJob struct {
	interval time.Duration
}

func NewLatestTrendJob(interval time.Duration) *LatestTrendJob {
	return &LatestTrendJob{interval: interval}
}

func (j *LatestTrendJob) Name() string {
	return "LatestTrendGeneration"
}

func (j *LatestTrendJob) Interval() time.Duration {
	return j.interval
}

func (j *LatestTrendJob) Execute(ctx context.Context, s *Scheduler) error {
	s.logger.Info("Checking for latest trend analysis generation...")

	// 現在時刻（UTC）を取得
	now := time.Now().UTC()

	// 1. auto_latest_trend_enabled が true のユーザーを取得
	usersQuery := `
		SELECT user_id
		FROM user_llms
		WHERE auto_latest_trend_enabled = true
	`

	rows, err := s.db.QueryContext(ctx, usersQuery)
	if err != nil {
		return fmt.Errorf("failed to query users with auto latest trend enabled: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close rows")
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
		s.logger.Info("No users with auto latest trend enabled")
		return nil
	}

	s.logger.WithField("count", len(userIDs)).Info("Found users with auto latest trend enabled")
	usersWithAutoSummaryGauge.WithLabelValues("latest_trend").Set(float64(len(userIDs)))

	// 2. 直近3日間の期間を計算（今日を除く）
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	periodEnd := today.AddDate(0, 0, -1)   // 昨日（1日前）
	periodStart := today.AddDate(0, 0, -3) // 3日前

	// 3. 各ユーザーについて、対象期間に日記があるかチェックし、メッセージをキューイング
	for _, userID := range userIDs {
		if err := j.processUserLatestTrend(ctx, s, userID, periodStart, periodEnd); err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Error processing latest trend for user")
			continue
		}
	}

	return nil
}

func (j *LatestTrendJob) processUserLatestTrend(ctx context.Context, s *Scheduler, userID string, periodStart, periodEnd time.Time) error {
	// 対象期間に日記が最小必要数以上存在するかチェック
	var count int
	checkQuery := `
		SELECT COUNT(*) FROM diaries
		WHERE user_id = $1
		AND date >= $2 AND date <= $3
	`
	if err := s.db.QueryRowContext(ctx, checkQuery, userID, periodStart, periodEnd).Scan(&count); err != nil {
		return fmt.Errorf("failed to check diary entries: %w", err)
	}

	if count < constants.MinDiaryEntriesForTrend {
		s.logger.WithFields(map[string]any{
			"user_id":       userID,
			"entry_count":   count,
			"required_days": constants.MinDiaryEntriesForTrend,
		}).Debug("Not enough diary entries for latest trend analysis")
		return nil
	}

	// Redis Pub/Sub経由でトレンド分析生成を依頼
	message := map[string]any{
		"type":         "latest_trend",
		"user_id":      userID,
		"period_start": periodStart.Format(time.RFC3339),
		"period_end":   periodEnd.Format(time.RFC3339),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to marshal message")
		return err
	}

	// Redisにメッセージを送信
	publishCmd := s.redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
	if err := s.redis.Do(ctx, publishCmd).Error(); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to publish message")
		return err
	}

	queuedMessagesCounter.WithLabelValues("latest_trend").Inc()
	s.logger.WithFields(map[string]any{
		"user_id":      userID,
		"period_start": periodStart.Format("2006-01-02"),
		"period_end":   periodEnd.Format("2006-01-02"),
	}).Debug("Queued latest trend generation")

	return nil
}
