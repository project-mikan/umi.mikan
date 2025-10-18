package constants

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

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
	DailySummaryInterval      time.Duration
	MonthlySummaryInterval    time.Duration
	LatestTrendTargetHour     int
	LatestTrendTargetMinute   int
}

type SubscriberConfig struct {
	MaxConcurrentJobs int
}

type RateLimitConfig struct {
	LoginMaxAttempts    int
	LoginWindow         time.Duration
	RegisterMaxAttempts int
	RegisterWindow      time.Duration
}

func LoadEnv(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", fmt.Errorf("env %s not found", name)
	}
	return value, nil
}

func LoadPort() (int, error) {
	portString, err := LoadEnv("PORT")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(portString)
}

func LoadJWTSecret() (string, error) {
	value, err := LoadEnv("JWT_SECRET")
	if err != nil {
		return "", err
	}
	return value, nil
}

func LoadDBConfig() (*DBConfig, error) {
	host, err := LoadEnv("DB_HOST")
	if err != nil {
		return nil, err
	}
	portString, err := LoadEnv("DB_PORT")
	if err != nil {
		return nil, err
	}
	// int型に変換
	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, err
	}

	user, err := LoadEnv("DB_USER")
	if err != nil {
		return nil, err
	}
	password, err := LoadEnv("DB_PASS")
	if err != nil {
		return nil, err
	}
	dbname, err := LoadEnv("DB_NAME")
	if err != nil {
		return nil, err
	}
	return &DBConfig{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
	}, nil
}

func LoadRedisConfig() (*RedisConfig, error) {
	host, err := LoadEnv("REDIS_HOST")
	if err != nil {
		return nil, err
	}
	portString, err := LoadEnv("REDIS_PORT")
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return nil, err
	}
	return &RedisConfig{
		Host: host,
		Port: port,
	}, nil
}

func LoadSchedulerConfig() (*SchedulerConfig, error) {
	dailyIntervalStr := os.Getenv("SCHEDULER_DAILY_INTERVAL")
	if dailyIntervalStr == "" {
		dailyIntervalStr = "5m" // Default to 5 minutes
	}

	monthlyIntervalStr := os.Getenv("SCHEDULER_MONTHLY_INTERVAL")
	if monthlyIntervalStr == "" {
		monthlyIntervalStr = "5m" // Default to 5 minutes
	}

	latestTrendHourStr := os.Getenv("SCHEDULER_LATEST_TREND_HOUR")
	if latestTrendHourStr == "" {
		latestTrendHourStr = "4" // Default to 4 AM
	}

	latestTrendMinuteStr := os.Getenv("SCHEDULER_LATEST_TREND_MINUTE")
	if latestTrendMinuteStr == "" {
		latestTrendMinuteStr = "0" // Default to 0 minutes
	}

	dailyInterval, err := time.ParseDuration(dailyIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SCHEDULER_DAILY_INTERVAL format: %w", err)
	}

	monthlyInterval, err := time.ParseDuration(monthlyIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SCHEDULER_MONTHLY_INTERVAL format: %w", err)
	}

	latestTrendHour, err := strconv.Atoi(latestTrendHourStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SCHEDULER_LATEST_TREND_HOUR format: %w", err)
	}
	if latestTrendHour < 0 || latestTrendHour > 23 {
		return nil, fmt.Errorf("SCHEDULER_LATEST_TREND_HOUR must be between 0 and 23, got %d", latestTrendHour)
	}

	latestTrendMinute, err := strconv.Atoi(latestTrendMinuteStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SCHEDULER_LATEST_TREND_MINUTE format: %w", err)
	}
	if latestTrendMinute < 0 || latestTrendMinute > 59 {
		return nil, fmt.Errorf("SCHEDULER_LATEST_TREND_MINUTE must be between 0 and 59, got %d", latestTrendMinute)
	}

	return &SchedulerConfig{
		DailySummaryInterval:    dailyInterval,
		MonthlySummaryInterval:  monthlyInterval,
		LatestTrendTargetHour:   latestTrendHour,
		LatestTrendTargetMinute: latestTrendMinute,
	}, nil
}

func LoadSubscriberConfig() (*SubscriberConfig, error) {
	maxConcurrentJobsStr := os.Getenv("SUBSCRIBER_MAX_CONCURRENT_JOBS")
	if maxConcurrentJobsStr == "" {
		maxConcurrentJobsStr = "10" // Default to 10 concurrent jobs
	}

	maxConcurrentJobs, err := strconv.Atoi(maxConcurrentJobsStr)
	if err != nil {
		return nil, fmt.Errorf("invalid SUBSCRIBER_MAX_CONCURRENT_JOBS format: %w", err)
	}

	if maxConcurrentJobs <= 0 {
		return nil, fmt.Errorf("SUBSCRIBER_MAX_CONCURRENT_JOBS must be a positive integer")
	}

	return &SubscriberConfig{
		MaxConcurrentJobs: maxConcurrentJobs,
	}, nil
}

func LoadRateLimitConfig() (*RateLimitConfig, error) {
	// Login rate limit config
	loginMaxAttemptsStr := os.Getenv("LOGIN_MAX_ATTEMPTS")
	if loginMaxAttemptsStr == "" {
		loginMaxAttemptsStr = "5" // デフォルト: 5回まで
	}

	loginWindowStr := os.Getenv("LOGIN_WINDOW")
	if loginWindowStr == "" {
		loginWindowStr = "15m" // デフォルト: 15分
	}

	loginMaxAttempts, err := strconv.Atoi(loginMaxAttemptsStr)
	if err != nil {
		return nil, fmt.Errorf("invalid LOGIN_MAX_ATTEMPTS format: %w", err)
	}

	if loginMaxAttempts <= 0 {
		return nil, fmt.Errorf("LOGIN_MAX_ATTEMPTS must be a positive integer")
	}

	loginWindow, err := time.ParseDuration(loginWindowStr)
	if err != nil {
		return nil, fmt.Errorf("invalid LOGIN_WINDOW format: %w", err)
	}

	if loginWindow <= 0 {
		return nil, fmt.Errorf("LOGIN_WINDOW must be a positive duration")
	}

	// Register rate limit config
	registerMaxAttemptsStr := os.Getenv("REGISTER_MAX_ATTEMPTS")
	if registerMaxAttemptsStr == "" {
		registerMaxAttemptsStr = "3" // デフォルト: 3回まで
	}

	registerWindowStr := os.Getenv("REGISTER_WINDOW")
	if registerWindowStr == "" {
		registerWindowStr = "1h" // デフォルト: 1時間
	}

	registerMaxAttempts, err := strconv.Atoi(registerMaxAttemptsStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REGISTER_MAX_ATTEMPTS format: %w", err)
	}

	if registerMaxAttempts <= 0 {
		return nil, fmt.Errorf("REGISTER_MAX_ATTEMPTS must be a positive integer")
	}

	registerWindow, err := time.ParseDuration(registerWindowStr)
	if err != nil {
		return nil, fmt.Errorf("invalid REGISTER_WINDOW format: %w", err)
	}

	if registerWindow <= 0 {
		return nil, fmt.Errorf("REGISTER_WINDOW must be a positive duration")
	}

	return &RateLimitConfig{
		LoginMaxAttempts:    loginMaxAttempts,
		LoginWindow:         loginWindow,
		RegisterMaxAttempts: registerMaxAttempts,
		RegisterWindow:      registerWindow,
	}, nil
}

func LoadGRPCReflectionEnabled() bool {
	env := os.Getenv("BACKEND_ENV")

	// Production環境では一律でリフレクションを無効にする
	if env == "prod" || env == "production" {
		return false
	}

	// Development環境ではデフォルトで有効
	return true
}

// LoadRegisterKey REGISTER_KEY環境変数を読み込む（設定されていない場合は空文字を返す）
func LoadRegisterKey() string {
	return os.Getenv("REGISTER_KEY")
}
