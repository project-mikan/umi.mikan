package container

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/project-mikan/umi.mikan/backend/constants"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/llm"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/lock"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/ratelimiter"
	"github.com/project-mikan/umi.mikan/backend/service/auth"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
	"github.com/project-mikan/umi.mikan/backend/service/user"
	"github.com/redis/rueidis"
	"go.uber.org/dig"
)

// Container wraps the dig container
type Container struct {
	container *dig.Container
}

// NewContainer creates and configures a new DI container
func NewContainer() (*Container, error) {
	container := dig.New()

	c := &Container{
		container: container,
	}

	// Register all providers
	if err := c.registerProviders(); err != nil {
		return nil, fmt.Errorf("failed to register providers: %w", err)
	}

	return c, nil
}

// registerProviders registers all the providers
func (c *Container) registerProviders() error {
	// Configuration providers
	if err := c.container.Provide(NewDBConfig); err != nil {
		return fmt.Errorf("failed to provide NewDBConfig: %w", err)
	}
	if err := c.container.Provide(NewRedisConfig); err != nil {
		return fmt.Errorf("failed to provide NewRedisConfig: %w", err)
	}
	if err := c.container.Provide(NewSchedulerConfig); err != nil {
		return fmt.Errorf("failed to provide NewSchedulerConfig: %w", err)
	}
	if err := c.container.Provide(NewSubscriberConfig); err != nil {
		return fmt.Errorf("failed to provide NewSubscriberConfig: %w", err)
	}
	if err := c.container.Provide(NewRateLimitConfig); err != nil {
		return fmt.Errorf("failed to provide NewRateLimitConfig: %w", err)
	}

	// Infrastructure providers
	if err := c.container.Provide(NewDatabase); err != nil {
		return fmt.Errorf("failed to provide NewDatabase: %w", err)
	}
	if err := c.container.Provide(NewRedisClient); err != nil {
		return fmt.Errorf("failed to provide NewRedisClient: %w", err)
	}
	if err := c.container.Provide(NewLLMClientFactory); err != nil {
		return fmt.Errorf("failed to provide NewLLMClientFactory: %w", err)
	}
	if err := c.container.Provide(NewLockService); err != nil {
		return fmt.Errorf("failed to provide NewLockService: %w", err)
	}
	if err := c.container.Provide(NewRateLimiter); err != nil {
		return fmt.Errorf("failed to provide NewRateLimiter: %w", err)
	}
	if err := c.container.Provide(NewLoginAttemptLimiter); err != nil {
		return fmt.Errorf("failed to provide NewLoginAttemptLimiter: %w", err)
	}

	// Service providers
	if err := c.container.Provide(NewAuthService); err != nil {
		return fmt.Errorf("failed to provide NewAuthService: %w", err)
	}
	if err := c.container.Provide(NewDiaryService); err != nil {
		return fmt.Errorf("failed to provide NewDiaryService: %w", err)
	}
	if err := c.container.Provide(NewUserService); err != nil {
		return fmt.Errorf("failed to provide NewUserService: %w", err)
	}

	// Application providers
	if err := c.container.Provide(NewServerApp); err != nil {
		return fmt.Errorf("failed to provide NewServerApp: %w", err)
	}
	if err := c.container.Provide(NewSchedulerApp); err != nil {
		return fmt.Errorf("failed to provide NewSchedulerApp: %w", err)
	}
	if err := c.container.Provide(NewSubscriberApp); err != nil {
		return fmt.Errorf("failed to provide NewSubscriberApp: %w", err)
	}

	// Cleanup provider
	if err := c.container.Provide(NewCleanup); err != nil {
		return fmt.Errorf("failed to provide NewCleanup: %w", err)
	}

	return nil
}

// Invoke runs the provided function with dependency injection
func (c *Container) Invoke(fn any) error {
	return c.container.Invoke(fn)
}

// Configuration types
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type RedisConfig struct {
	Host string
	Port int
}

type SchedulerConfig struct {
	DailySummaryInterval   time.Duration
	MonthlySummaryInterval time.Duration
}

type SubscriberConfig struct {
	MaxConcurrentJobs int
}

type RateLimitConfig struct {
	LoginMaxAttempts int
	LoginWindow      time.Duration
}

// LLMClientFactory creates LLM clients
type LLMClientFactory interface {
	CreateGeminiClient(ctx context.Context, apiKey string) (*llm.GeminiClient, error)
}

type geminiClientFactory struct{}

func (f *geminiClientFactory) CreateGeminiClient(ctx context.Context, apiKey string) (*llm.GeminiClient, error) {
	return llm.NewGeminiClient(ctx, apiKey)
}

// LockService provides distributed locking functionality
type LockService interface {
	NewDistributedLock(key string, duration time.Duration) lock.DistributedLockInterface
}

type lockService struct {
	redis rueidis.Client
}

func (s *lockService) NewDistributedLock(key string, duration time.Duration) lock.DistributedLockInterface {
	return lock.NewDistributedLock(s.redis, key, duration)
}

// Provider functions

// NewDBConfig creates database configuration
func NewDBConfig() (*DBConfig, error) {
	config, err := constants.LoadDBConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load DB config: %w", err)
	}

	return &DBConfig{
		Host:     config.Host,
		Port:     config.Port,
		User:     config.User,
		Password: config.Password,
		DBName:   config.DBName,
	}, nil
}

// NewRedisConfig creates Redis configuration
func NewRedisConfig() (*RedisConfig, error) {
	config, err := constants.LoadRedisConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load Redis config: %w", err)
	}

	return &RedisConfig{
		Host: config.Host,
		Port: config.Port,
	}, nil
}

// NewSchedulerConfig creates scheduler configuration
func NewSchedulerConfig() (*SchedulerConfig, error) {
	config, err := constants.LoadSchedulerConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load scheduler config: %w", err)
	}

	return &SchedulerConfig{
		DailySummaryInterval:   config.DailySummaryInterval,
		MonthlySummaryInterval: config.MonthlySummaryInterval,
	}, nil
}

// NewSubscriberConfig creates subscriber configuration
func NewSubscriberConfig() (*SubscriberConfig, error) {
	config, err := constants.LoadSubscriberConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load subscriber config: %w", err)
	}

	return &SubscriberConfig{
		MaxConcurrentJobs: config.MaxConcurrentJobs,
	}, nil
}

// NewRateLimitConfig creates rate limit configuration
func NewRateLimitConfig() (*RateLimitConfig, error) {
	config, err := constants.LoadRateLimitConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load rate limit config: %w", err)
	}

	return &RateLimitConfig{
		LoginMaxAttempts: config.LoginMaxAttempts,
		LoginWindow:      config.LoginWindow,
	}, nil
}

// NewDatabase creates a database connection
func NewDatabase(config *DBConfig) (database.DB, error) {
	db := database.NewDB(config.Host, config.Port, config.User, config.Password, config.DBName)
	log.Printf("Database connection established: %s:%d/%s", config.Host, config.Port, config.DBName)
	return db, nil
}

// NewRedisClient creates a Redis client
func NewRedisClient(config *RedisConfig) (rueidis.Client, error) {
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	log.Printf("Redis connection established: %s:%d", config.Host, config.Port)
	return client, nil
}

// NewLLMClientFactory creates an LLM client factory
func NewLLMClientFactory() LLMClientFactory {
	return &geminiClientFactory{}
}

// NewLockService creates a lock service
func NewLockService(redis rueidis.Client) LockService {
	return &lockService{redis: redis}
}

// NewRateLimiter creates a rate limiter
func NewRateLimiter(redis rueidis.Client) ratelimiter.RateLimiter {
	return ratelimiter.NewRedisRateLimiter(redis)
}

// NewLoginAttemptLimiter creates a login attempt limiter
func NewLoginAttemptLimiter(rateLimiter ratelimiter.RateLimiter, config *RateLimitConfig) *ratelimiter.LoginAttemptLimiter {
	return ratelimiter.NewLoginAttemptLimiter(rateLimiter, config.LoginMaxAttempts, config.LoginWindow)
}

// NewAuthService creates an auth service
func NewAuthService(db database.DB, loginLimiter *ratelimiter.LoginAttemptLimiter) *auth.AuthEntry {
	return &auth.AuthEntry{DB: db, LoginLimiter: loginLimiter}
}

// NewDiaryService creates a diary service
func NewDiaryService(db database.DB, redis rueidis.Client) *diary.DiaryEntry {
	return &diary.DiaryEntry{DB: db, Redis: redis}
}

// NewUserService creates a user service
func NewUserService(db database.DB, redis rueidis.Client) *user.UserEntry {
	return &user.UserEntry{DB: db, RedisClient: redis}
}

// Application types

// ServerApp represents the gRPC server application
type ServerApp struct {
	DB           database.DB
	Redis        rueidis.Client
	AuthService  *auth.AuthEntry
	DiaryService *diary.DiaryEntry
	UserService  *user.UserEntry
}

// SchedulerApp represents the scheduler application
type SchedulerApp struct {
	DB              database.DB
	Redis           rueidis.Client
	SchedulerConfig *SchedulerConfig
}

// SubscriberApp represents the subscriber application
type SubscriberApp struct {
	DB               database.DB
	Redis            rueidis.Client
	LLMFactory       LLMClientFactory
	LockService      LockService
	SubscriberConfig *SubscriberConfig
}

// NewServerApp creates a server application
func NewServerApp(
	db database.DB,
	redis rueidis.Client,
	authService *auth.AuthEntry,
	diaryService *diary.DiaryEntry,
	userService *user.UserEntry,
) *ServerApp {
	return &ServerApp{
		DB:           db,
		Redis:        redis,
		AuthService:  authService,
		DiaryService: diaryService,
		UserService:  userService,
	}
}

// NewSchedulerApp creates a scheduler application
func NewSchedulerApp(
	db database.DB,
	redis rueidis.Client,
	config *SchedulerConfig,
) *SchedulerApp {
	return &SchedulerApp{
		DB:              db,
		Redis:           redis,
		SchedulerConfig: config,
	}
}

// NewSubscriberApp creates a subscriber application
func NewSubscriberApp(
	db database.DB,
	redis rueidis.Client,
	llmFactory LLMClientFactory,
	lockService LockService,
	config *SubscriberConfig,
) *SubscriberApp {
	return &SubscriberApp{
		DB:               db,
		Redis:            redis,
		LLMFactory:       llmFactory,
		LockService:      lockService,
		SubscriberConfig: config,
	}
}

// Cleanup provides a way to clean up resources
type Cleanup struct {
	db    database.DB
	redis rueidis.Client
}

// NewCleanup creates a cleanup handler
func NewCleanup(db database.DB, redis rueidis.Client) *Cleanup {
	return &Cleanup{db: db, redis: redis}
}

// Close closes all connections
func (c *Cleanup) Close() error {
	var err error

	if c.db != nil {
		// Cast database.DB to *sql.DB to access Close method
		if sqlDB, ok := c.db.(*sql.DB); ok {
			if closeErr := sqlDB.Close(); closeErr != nil {
				log.Printf("Failed to close database connection: %v", closeErr)
				err = closeErr
			} else {
				log.Printf("Database connection closed")
			}
		} else {
			log.Printf("Warning: Cannot close database connection - unexpected type %T", c.db)
		}
	}

	if c.redis != nil {
		c.redis.Close()
		log.Printf("Redis connection closed")
	}

	return err
}
