package main

import (
	"context"
	"database/sql"
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
	db     *sql.DB
	redis  rueidis.Client
	ctx    context.Context
	cancel context.CancelFunc
	logger *logrus.Entry
}

// ScheduledJob インターフェース: 間隔ベースのジョブ用
type ScheduledJob interface {
	Name() string
	Interval() time.Duration
	Execute(ctx context.Context, s *Scheduler) error
}

// DailyScheduledJob インターフェース: 毎日特定時刻に実行するジョブ用
type DailyScheduledJob interface {
	Name() string
	TargetHour() int   // 実行する時（0-23, JST）
	TargetMinute() int // 実行する分（0-59, JST）
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

// AddDailyJob 毎日特定時刻に実行するジョブを追加
func (s *Scheduler) AddDailyJob(job DailyScheduledJob) {
	go func() {
		// 毎分チェックする
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		s.logger.WithFields(logrus.Fields{
			"job_name":      job.Name(),
			"target_hour":   job.TargetHour(),
			"target_minute": job.TargetMinute(),
		}).Info("Daily scheduled job started")

		// 最後に実行した日付を記録（重複実行を防ぐ）
		var lastExecutedDate string

		for {
			select {
			case <-s.ctx.Done():
				s.logger.WithField("job_name", job.Name()).Info("Daily scheduled job stopped")
				return
			case <-ticker.C:
				// 現在時刻（JST）を取得
				jst, err := time.LoadLocation("Asia/Tokyo")
				if err != nil {
					s.logger.WithError(err).Warn("Failed to load Asia/Tokyo location, using fixed offset")
					jst = time.FixedZone("Asia/Tokyo", 9*60*60)
				}
				now := time.Now().In(jst)

				// 現在の時刻が目的の時刻でない場合はスキップ
				if now.Hour() != job.TargetHour() || now.Minute() != job.TargetMinute() {
					continue
				}

				// 今日の日付
				currentDate := now.Format("2006-01-02")

				// 今日既に実行済みの場合はスキップ
				if lastExecutedDate == currentDate {
					continue
				}

				s.logger.WithFields(logrus.Fields{
					"job_name":      job.Name(),
					"current_time":  now.Format("2006-01-02 15:04:05"),
					"target_hour":   job.TargetHour(),
					"target_minute": job.TargetMinute(),
				}).Info("Executing daily scheduled job")

				// Metrics tracking
				start := time.Now()
				execErr := job.Execute(s.ctx, s)
				duration := time.Since(start)

				jobDuration.WithLabelValues(job.Name()).Observe(duration.Seconds())

				if execErr != nil {
					s.logger.WithError(execErr).WithFields(logrus.Fields{
						"job_name": job.Name(),
						"duration": duration,
					}).Error("Error executing daily job")
					jobExecutionCounter.WithLabelValues(job.Name(), "error").Inc()
				} else {
					// 実行成功時のみ日付を記録
					lastExecutedDate = currentDate
					s.logger.WithFields(logrus.Fields{
						"job_name": job.Name(),
						"duration": duration,
					}).Info("Daily job executed successfully")
					jobExecutionCounter.WithLabelValues(job.Name(), "success").Inc()
				}
			}
		}
	}()
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
	scheduler.AddJob(NewMonthlySummaryJob(app.SchedulerConfig.MonthlySummaryInterval))
	scheduler.AddDailyJob(NewLatestTrendJob(
		app.SchedulerConfig.LatestTrendTargetHour,
		app.SchedulerConfig.LatestTrendTargetMinute,
	))
	scheduler.AddDailyJob(NewDiaryEmbeddingJob(
		app.SchedulerConfig.DiaryEmbeddingTargetHour,
		app.SchedulerConfig.DiaryEmbeddingTargetMinute,
	))

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
	userIDs, err := database.UserIDsWithAutoSummaryMonthly(ctx, s.db)
	if err != nil {
		return fmt.Errorf("failed to query users with auto monthly summary enabled: %w", err)
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
	missingMonths, err := database.MonthsNeedingMonthlySummary(ctx, s.db, userID)
	if err != nil {
		return fmt.Errorf("failed to query missing monthly summaries for user %s: %w", userID, err)
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
	targetHour   int // 実行する時（0-23）
	targetMinute int // 実行する分（0-59）
}

func NewLatestTrendJob(targetHour, targetMinute int) *LatestTrendJob {
	return &LatestTrendJob{
		targetHour:   targetHour,
		targetMinute: targetMinute,
	}
}

func (j *LatestTrendJob) Name() string {
	return "LatestTrendGeneration"
}

func (j *LatestTrendJob) TargetHour() int {
	return j.targetHour
}

func (j *LatestTrendJob) TargetMinute() int {
	return j.targetMinute
}

func (j *LatestTrendJob) Execute(ctx context.Context, s *Scheduler) error {
	s.logger.Info("Starting latest trend analysis generation")

	// 1. auto_latest_trend_enabled が true のユーザーを取得
	userIDs, err := database.UserIDsWithAutoLatestTrendEnabled(ctx, s.db)
	if err != nil {
		return fmt.Errorf("failed to query users with auto latest trend enabled: %w", err)
	}

	if len(userIDs) == 0 {
		s.logger.Info("No users with auto latest trend enabled")
		return nil
	}

	s.logger.WithField("count", len(userIDs)).Info("Found users with auto latest trend enabled")
	usersWithAutoSummaryGauge.WithLabelValues("latest_trend").Set(float64(len(userIDs)))

	// 2. 直近3日間の期間を計算（今日を除く）
	periodStart, periodEnd := calculateTrendPeriod(time.Now())

	// 3. 各ユーザーについて、対象期間に日記があるかチェックし、メッセージをキューイング
	for _, userID := range userIDs {
		if err := j.processUserLatestTrend(ctx, s, userID, periodStart, periodEnd); err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Error processing latest trend for user")
			continue
		}
	}

	return nil
}

// calculateTrendPeriod は、指定された時刻を基準にトレンド分析対象期間を計算します
// 日記のdate列は日本時間ベースの日付をUTC 00:00:00として保存しているため、
// JST時刻を基準にして「昨日」「3日前」の日付を計算し、UTC 00:00:00として表現してDB検索に使用
func calculateTrendPeriod(now time.Time) (periodStart, periodEnd time.Time) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		jst = time.FixedZone("Asia/Tokyo", 9*60*60)
	}
	nowJST := now.In(jst)
	todayJST := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), 0, 0, 0, 0, jst)

	// 昨日と3日前のJST日付を計算
	yesterdayJST := todayJST.AddDate(0, 0, -1)
	threeDaysAgoJST := todayJST.AddDate(0, 0, -3)

	// JSTの日付をUTC 00:00:00として表現（diariesテーブルの保存形式に合わせる）
	periodEnd = time.Date(yesterdayJST.Year(), yesterdayJST.Month(), yesterdayJST.Day(), 0, 0, 0, 0, time.UTC)
	periodStart = time.Date(threeDaysAgoJST.Year(), threeDaysAgoJST.Month(), threeDaysAgoJST.Day(), 0, 0, 0, 0, time.UTC)

	return periodStart, periodEnd
}

func (j *LatestTrendJob) processUserLatestTrend(ctx context.Context, s *Scheduler, userID string, periodStart, periodEnd time.Time) error {
	// 古いRedisキーを明示的に削除（新しいデータ生成前にクリーンアップ）
	trendKey := fmt.Sprintf("latest_trend:%s", userID)
	delCmd := s.redis.B().Del().Key(trendKey).Build()
	if err := s.redis.Do(ctx, delCmd).Error(); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Warn("Failed to delete old trend key (continuing anyway)")
		// エラーがあっても処理は継続
	} else {
		s.logger.WithField("user_id", userID).Debug("Deleted old trend key")
	}

	// タスク開始時刻をRedisに記録
	taskKey := fmt.Sprintf("task:latest_trend:%s", userID)
	startTime := time.Now().Unix()
	setCmd := s.redis.B().Set().Key(taskKey).Value(fmt.Sprintf("%d", startTime)).Ex(time.Hour).Build()
	if err := s.redis.Do(ctx, setCmd).Error(); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Warn("Failed to record task start time")
		// エラーがあっても処理は継続
	}

	// 対象期間に日記が最小必要数以上存在するかチェック
	count, err := database.DiaryCountInDateRange(ctx, s.db, userID, periodStart, periodEnd)
	if err != nil {
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

// DiaryEmbeddingJob は前日の日記の埋め込みベクトルを翌朝生成するジョブ
// 当日中は日記を継ぎ足す可能性があるため、on-saveでの即時処理をスキップし
// 翌朝このジョブが昨日の日記をまとめて処理する（意味的検索有効ユーザーのみ）
type DiaryEmbeddingJob struct {
	targetHour   int // 実行する時（0-23, JST）
	targetMinute int // 実行する分（0-59, JST）
}

func NewDiaryEmbeddingJob(targetHour, targetMinute int) *DiaryEmbeddingJob {
	return &DiaryEmbeddingJob{
		targetHour:   targetHour,
		targetMinute: targetMinute,
	}
}

func (j *DiaryEmbeddingJob) Name() string {
	return "DiaryEmbeddingGeneration"
}

func (j *DiaryEmbeddingJob) TargetHour() int {
	return j.targetHour
}

func (j *DiaryEmbeddingJob) TargetMinute() int {
	return j.targetMinute
}

func (j *DiaryEmbeddingJob) Execute(ctx context.Context, s *Scheduler) error {
	s.logger.Info("Starting diary embedding generation for yesterday's diaries")

	// 1. semantic_search_enabled が true のユーザーを取得
	userIDs, err := database.UserIDsWithSemanticSearchEnabled(ctx, s.db)
	if err != nil {
		return fmt.Errorf("failed to query users with semantic search enabled: %w", err)
	}

	if len(userIDs) == 0 {
		s.logger.Info("No users with semantic search enabled")
		return nil
	}

	s.logger.WithField("count", len(userIDs)).Info("Found users with semantic search enabled")
	usersWithAutoSummaryGauge.WithLabelValues("diary_embedding").Set(float64(len(userIDs)))

	// 2. 昨日の日付を計算（JST基準）
	yesterdayUTC := calculateYesterdayUTC(time.Now())

	// 3. 各ユーザーについて昨日の日記のembedding生成をキューイング
	for _, userID := range userIDs {
		if err := j.processUserDiaryEmbedding(ctx, s, userID, yesterdayUTC); err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Error processing diary embedding for user")
			continue
		}
	}

	return nil
}

// calculateYesterdayUTC は指定時刻を基準に昨日（JST）の日付をUTC 00:00:00として返す
// diariesテーブルのdate列はJSTの日付をUTC 00:00:00として保存しているため、それに合わせる
func calculateYesterdayUTC(now time.Time) time.Time {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		jst = time.FixedZone("Asia/Tokyo", 9*60*60)
	}
	nowJST := now.In(jst)
	yesterdayJST := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day()-1, 0, 0, 0, 0, jst)
	return time.Date(yesterdayJST.Year(), yesterdayJST.Month(), yesterdayJST.Day(), 0, 0, 0, 0, time.UTC)
}

func (j *DiaryEmbeddingJob) processUserDiaryEmbedding(ctx context.Context, s *Scheduler, userID string, targetDate time.Time) error {
	// 対象日付の日記を取得
	// embedding未生成またはembeddingのupdated_atより日記のupdated_atが新しい場合に処理対象とする
	diaryIDs, err := database.DiaryIDsNeedingEmbedding(ctx, s.db, userID, targetDate)
	if err != nil {
		return fmt.Errorf("failed to query diary for user %s date %s: %w", userID, targetDate.Format("2006-01-02"), err)
	}

	if len(diaryIDs) == 0 {
		s.logger.WithFields(map[string]any{
			"user_id": userID,
			"date":    targetDate.Format("2006-01-02"),
		}).Debug("No diary requiring embedding for user on target date")
		return nil
	}

	for _, diaryID := range diaryIDs {
		message := map[string]any{
			"type":     "diary_embedding",
			"user_id":  userID,
			"diary_id": diaryID,
		}

		messageBytes, err := json.Marshal(message)
		if err != nil {
			s.logger.WithError(err).WithFields(map[string]any{"user_id": userID, "diary_id": diaryID}).Error("Failed to marshal message")
			continue
		}

		publishCmd := s.redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
		if err := s.redis.Do(ctx, publishCmd).Error(); err != nil {
			s.logger.WithError(err).WithFields(map[string]any{"user_id": userID, "diary_id": diaryID}).Error("Failed to publish message")
			continue
		}

		queuedMessagesCounter.WithLabelValues("diary_embedding").Inc()
		s.logger.WithFields(map[string]any{
			"user_id":  userID,
			"diary_id": diaryID,
			"date":     targetDate.Format("2006-01-02"),
		}).Debug("Queued diary embedding generation")
	}

	return nil
}
