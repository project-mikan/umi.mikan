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

	"github.com/project-mikan/umi.mikan/backend/container"
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

type ScheduledJob interface {
	Name() string
	Interval() time.Duration
	Execute(ctx context.Context, s *Scheduler) error
}

func NewScheduler(app *container.SchedulerApp, logger *logrus.Entry) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		db:     app.DB.(*sql.DB),
		redis:  app.Redis,
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}
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
	diContainer := container.NewContainer()

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
	scheduler := NewScheduler(app, logger)

	// ジョブを追加
	scheduler.AddJob(NewDailySummaryJob(app.SchedulerConfig.DailySummaryInterval))
	scheduler.AddJob(NewMonthlySummaryJob(app.SchedulerConfig.MonthlySummaryInterval))

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
	return "daily_summary"
}

func (j *DailySummaryJob) Interval() time.Duration {
	return j.interval
}

func (j *DailySummaryJob) Execute(ctx context.Context, s *Scheduler) error {
	// Query users with auto-summary enabled for daily summaries
	rows, err := s.db.QueryContext(ctx, `
		SELECT user_id FROM user_llms
		WHERE auto_summary_daily = true
	`)
	if err != nil {
		return fmt.Errorf("failed to query users with daily auto-summary: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close rows")
		}
	}()

	var userCount int
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			s.logger.WithError(err).Error("Failed to scan user ID")
			continue
		}

		// Create daily summary message
		message := map[string]interface{}{
			"type":    "daily_summary",
			"user_id": userID,
			"date":    yesterday,
		}

		messageBytes, err := json.Marshal(message)
		if err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to marshal daily summary message")
			continue
		}

		// Publish to Redis
		cmd := s.redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
		if err := s.redis.Do(ctx, cmd).Error(); err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to publish daily summary message")
			continue
		}

		queuedMessagesCounter.WithLabelValues("daily_summary").Inc()
		userCount++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over users: %w", err)
	}

	usersWithAutoSummaryGauge.WithLabelValues("daily").Set(float64(userCount))
	s.logger.WithField("user_count", userCount).Info("Daily summary jobs queued")
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
	return "monthly_summary"
}

func (j *MonthlySummaryJob) Interval() time.Duration {
	return j.interval
}

func (j *MonthlySummaryJob) Execute(ctx context.Context, s *Scheduler) error {
	// Query users with auto-summary enabled for monthly summaries
	rows, err := s.db.QueryContext(ctx, `
		SELECT user_id FROM user_llms
		WHERE auto_summary_monthly = true
	`)
	if err != nil {
		return fmt.Errorf("failed to query users with monthly auto-summary: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close rows")
		}
	}()

	var userCount int
	lastMonth := time.Now().AddDate(0, -1, 0)
	year := lastMonth.Year()
	month := int(lastMonth.Month())

	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			s.logger.WithError(err).Error("Failed to scan user ID")
			continue
		}

		// Create monthly summary message
		message := map[string]interface{}{
			"type":    "monthly_summary",
			"user_id": userID,
			"year":    year,
			"month":   month,
		}

		messageBytes, err := json.Marshal(message)
		if err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to marshal monthly summary message")
			continue
		}

		// Publish to Redis
		cmd := s.redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
		if err := s.redis.Do(ctx, cmd).Error(); err != nil {
			s.logger.WithError(err).WithField("user_id", userID).Error("Failed to publish monthly summary message")
			continue
		}

		queuedMessagesCounter.WithLabelValues("monthly_summary").Inc()
		userCount++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over users: %w", err)
	}

	usersWithAutoSummaryGauge.WithLabelValues("monthly").Set(float64(userCount))
	s.logger.WithField("user_count", userCount).Info("Monthly summary jobs queued")
	return nil
}
