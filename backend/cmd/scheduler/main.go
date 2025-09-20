package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/redis/rueidis"
)

type Scheduler struct {
	db     *database.DB
	redis  rueidis.Client
	ctx    context.Context
	cancel context.CancelFunc
}

type ScheduledJob interface {
	Name() string
	Interval() time.Duration
	Execute(ctx context.Context, s *Scheduler) error
}

func NewScheduler(db *database.DB, redis rueidis.Client) *Scheduler {
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
				if err := job.Execute(s.ctx, s); err != nil {
					log.Printf("Error executing job '%s': %v", job.Name(), err)
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

	// スケジューラー作成
	scheduler := NewScheduler(db, redisClient)

	// ジョブを追加
	scheduler.AddJob(&DailySummaryJob{})
	scheduler.AddJob(&MonthlySummaryJob{})

	log.Print("Scheduler is running...")

	// プログラム終了まで待機
	select {}
}