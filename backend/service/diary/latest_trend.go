package diary

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LatestTrendGenerationMessage はトレンド分析生成のためのメッセージ
type LatestTrendGenerationMessage struct {
	Type        string `json:"type"`
	UserID      string `json:"user_id"`
	PeriodStart string `json:"period_start"` // ISO 8601 format
	PeriodEnd   string `json:"period_end"`   // ISO 8601 format
}

// LatestTrendData はRedisに保存するトレンド分析データ
type LatestTrendData struct {
	UserID      string `json:"user_id"`
	Analysis    string `json:"analysis"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
	GeneratedAt string `json:"generated_at"`
}

// GetLatestTrend は直近1週間の日記のトレンド分析を取得します
func (s *DiaryEntry) GetLatestTrend(
	ctx context.Context,
	req *g.GetLatestTrendRequest,
) (*g.GetLatestTrendResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Redisからトレンド分析を取得
	trendKey := fmt.Sprintf("latest_trend:%s", userIDStr)
	getCmd := s.Redis.B().Get().Key(trendKey).Build()
	result := s.Redis.Do(ctx, getCmd)

	if err := result.Error(); err != nil {
		return nil, status.Error(codes.NotFound, "Latest trend analysis not found")
	}

	trendDataStr, err := result.ToString()
	if err != nil {
		return nil, status.Error(codes.NotFound, "Latest trend analysis not found")
	}

	// JSONをパース
	var trendData LatestTrendData
	if err := json.Unmarshal([]byte(trendDataStr), &trendData); err != nil {
		return nil, status.Error(codes.Internal, "Failed to parse trend data")
	}

	return &g.GetLatestTrendResponse{
		Analysis:    trendData.Analysis,
		PeriodStart: trendData.PeriodStart,
		PeriodEnd:   trendData.PeriodEnd,
		GeneratedAt: trendData.GeneratedAt,
	}, nil
}

// TriggerLatestTrend はトレンド分析の生成を手動でトリガーします（デバッグ用）
func (s *DiaryEntry) TriggerLatestTrend(
	ctx context.Context,
	req *g.TriggerLatestTrendRequest,
) (*g.TriggerLatestTrendResponse, error) {
	// production環境では使用不可
	if os.Getenv("ENV") == "production" {
		return nil, status.Error(codes.PermissionDenied, "This operation is not available in production environment")
	}

	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// ユーザーのLLMキーが設定されているかチェック
	_, err = database.UserLlmByUserIDLlmProvider(ctx, s.DB, userID, 1) // Gemini
	if err != nil {
		return nil, status.Error(codes.NotFound, "Gemini API key not configured")
	}

	// 直近7日間の期間を計算（今日を除く）
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	periodEnd := today.AddDate(0, 0, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)                      // 昨日の23:59:59
	periodStart := periodEnd.AddDate(0, 0, -6).Add(-23*time.Hour - 59*time.Minute - 59*time.Second + time.Second) // 7日前の00:00:00

	// 対象期間の日記エントリが存在するかチェック
	var count int
	checkQuery := `
		SELECT COUNT(*) FROM diaries
		WHERE user_id = $1
		AND date >= $2 AND date <= $3
	`
	err = s.DB.QueryRowContext(ctx, checkQuery, userID, periodStart, periodEnd).Scan(&count)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to check diary entries")
	}
	if count == 0 {
		return &g.TriggerLatestTrendResponse{
			Success: false,
			Message: "分析期間の日記エントリが見つかりませんでした",
		}, nil
	}

	// Redis Pub/Sub経由でトレンド分析生成を依頼
	message := LatestTrendGenerationMessage{
		Type:        "latest_trend",
		UserID:      userIDStr,
		PeriodStart: periodStart.Format(time.RFC3339),
		PeriodEnd:   periodEnd.Format(time.RFC3339),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to create latest trend generation request")
	}

	// Redisにメッセージを送信
	publishCmd := s.Redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
	if err := s.Redis.Do(ctx, publishCmd).Error(); err != nil {
		return nil, status.Error(codes.Internal, "Failed to queue latest trend generation")
	}

	return &g.TriggerLatestTrendResponse{
		Success: true,
		Message: "トレンド分析の生成をキューに追加しました",
	}, nil
}
