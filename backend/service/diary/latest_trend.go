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
	UserID       string `json:"user_id"`
	Health       string `json:"health"`        // 体調: "bad", "slight", "normal", "good"
	HealthReason string `json:"health_reason"` // 体調の理由（10文字以内）
	Mood         string `json:"mood"`          // 気分: "bad", "slight", "normal", "good"
	MoodReason   string `json:"mood_reason"`   // 気分の理由（10文字以内）
	Activities   string `json:"activities"`    // 活動・行動（箇条書き・階層構造のテキスト）
	PeriodStart  string `json:"period_start"`
	PeriodEnd    string `json:"period_end"`
	GeneratedAt  string `json:"generated_at"`
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
		Health:       trendData.Health,
		HealthReason: trendData.HealthReason,
		Mood:         trendData.Mood,
		MoodReason:   trendData.MoodReason,
		Activities:   trendData.Activities,
		PeriodStart:  trendData.PeriodStart,
		PeriodEnd:    trendData.PeriodEnd,
		GeneratedAt:  trendData.GeneratedAt,
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

	// 直近1週間程度の期間を計算（今日を除く、前日を中心に参考）
	// JST時刻を基準にして「昨日」を計算（実行時刻の前日が昨日）
	jst, jstErr := time.LoadLocation("Asia/Tokyo")
	if jstErr != nil {
		jst = time.FixedZone("Asia/Tokyo", 9*60*60)
	}
	nowJST := time.Now().In(jst)
	todayJST := time.Date(nowJST.Year(), nowJST.Month(), nowJST.Day(), 0, 0, 0, 0, jst)
	// UTC時刻に変換して期間を設定
	todayUTC := todayJST.UTC()
	periodEnd := todayUTC.AddDate(0, 0, -1)   // 昨日（JST基準での前日）
	periodStart := todayUTC.AddDate(0, 0, -7) // 7日前

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
