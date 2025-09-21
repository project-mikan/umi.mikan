package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/redis/rueidis"
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

type Scheduler struct {
	db     *sql.DB
	redis  rueidis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

type ScheduledJob interface {
	Name() string
	Interval() time.Duration
	Execute(ctx context.Context, s *Scheduler) error
}

func NewScheduler(db *sql.DB, redis rueidis.Client) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		db:     db,
		redis:  redis,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Scheduler) AddJob(job ScheduledJob) {
	go func() {
		ticker := time.NewTicker(job.Interval())
		defer ticker.Stop()

		log.Printf("Scheduled job '%s' started with interval %v", job.Name(), job.Interval())

		for {
			select {
			case <-s.ctx.Done():
				log.Printf("Scheduled job '%s' stopped", job.Name())
				return
			case <-ticker.C:
				log.Printf("Executing job: %s", job.Name())

				// Metrics tracking
				start := time.Now()
				err := job.Execute(s.ctx, s)
				duration := time.Since(start)

				jobDuration.WithLabelValues(job.Name()).Observe(duration.Seconds())

				if err != nil {
					log.Printf("Error executing job '%s': %v", job.Name(), err)
					jobExecutionCounter.WithLabelValues(job.Name(), "error").Inc()
				} else {
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
	log.Print("=== umi.mikan scheduler started ===")

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
	redisClient, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port)},
	})
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer redisClient.Close()

	// Redis接続確認
	ctx := context.Background()
	pingCmd := redisClient.B().Ping().Build()
	if err := redisClient.Do(ctx, pingCmd).Error(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Print("Connected to Redis successfully")

	// メトリクスサーバー開始
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Print("Metrics server starting on :8081")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// スケジューラー作成
	scheduler := NewScheduler(db, redisClient)

	// ジョブを追加
	scheduler.AddJob(&DailySummaryJob{})
	scheduler.AddJob(&MonthlySummaryJob{})

	log.Print("Scheduler is running...")

	// プログラム終了まで待機
	select {}
}

// DailySummaryJob implementation
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
	usersWithAutoSummaryGauge.WithLabelValues("daily").Set(float64(len(userIDs)))

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
	// diary_summary_daysにsummaryがない日、または要約のupdated_atが日記のupdated_atより古い日を見つける（今日を除く）
	query := `
		SELECT d.date
		FROM diaries d
		LEFT JOIN diary_summary_days dsd ON d.user_id = dsd.user_id AND d.date = dsd.date
		WHERE d.user_id = $1
		  AND d.date < CURRENT_DATE
		  AND (dsd.id IS NULL OR dsd.updated_at < d.updated_at)
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

		queuedMessagesCounter.WithLabelValues("daily_summary").Inc()
		log.Printf("Queued summary generation for user %s, date %s", userID, date.Format("2006-01-02"))
	}

	return nil
}

// MonthlySummaryJob implementation
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
	usersWithAutoSummaryGauge.WithLabelValues("monthly").Set(float64(len(userIDs)))

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
	// diary_summary_monthsに月次要約がない月、またはその月の日記の最新updated_atより月次要約のupdated_atが古い月を見つける（今月を除く）
	query := `
		WITH monthly_diary_stats AS (
			SELECT
				EXTRACT(YEAR FROM d.date) as year,
				EXTRACT(MONTH FROM d.date) as month,
				MAX(d.updated_at) as latest_diary_updated_at
			FROM diaries d
			WHERE d.user_id = $1
			GROUP BY EXTRACT(YEAR FROM d.date), EXTRACT(MONTH FROM d.date)
		),
		monthly_summary_exists AS (
			SELECT
				mds.year,
				mds.month,
				mds.latest_diary_updated_at,
				dsm.updated_at as summary_updated_at
			FROM monthly_diary_stats mds
			LEFT JOIN diary_summary_months dsm ON dsm.user_id = $1
				AND dsm.year = mds.year
				AND dsm.month = mds.month
			WHERE EXISTS (
				SELECT 1 FROM diary_summary_days dsd
				WHERE dsd.user_id = $1
				AND EXTRACT(YEAR FROM dsd.date) = mds.year
				AND EXTRACT(MONTH FROM dsd.date) = mds.month
			)
		)
		SELECT year, month
		FROM monthly_summary_exists
		WHERE (year < EXTRACT(YEAR FROM CURRENT_DATE)
			OR (year = EXTRACT(YEAR FROM CURRENT_DATE) AND month < EXTRACT(MONTH FROM CURRENT_DATE)))
		AND (summary_updated_at IS NULL OR summary_updated_at < latest_diary_updated_at)
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

		queuedMessagesCounter.WithLabelValues("monthly_summary").Inc()
		log.Printf("Queued monthly summary generation for user %s, year %d, month %d", userID, ym.Year, ym.Month)
	}

	return nil
}
