package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/request"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/redis/rueidis"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserEntry struct {
	g.UnimplementedUserServiceServer
	DB          *sql.DB
	RedisClient rueidis.Client
}

func (s *UserEntry) UpdateUserName(ctx context.Context, req *g.UpdateUserNameRequest) (*g.UpdateUserNameResponse, error) {
	// リクエストのバリデーション
	if req.GetNewName() == "" {
		return &g.UpdateUserNameResponse{
			Success: false,
			Message: "nameRequired",
		}, nil
	}

	// 名前の長さチェック（20文字以内）
	if len([]rune(req.GetNewName())) > 20 {
		return &g.UpdateUserNameResponse{
			Success: false,
			Message: "nameTooLong",
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.UpdateUserNameResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	// ユーザーの取得
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.UpdateUserNameResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	userDB, err := database.UserByID(ctx, s.DB, parsedUserID)
	if err != nil {
		return &g.UpdateUserNameResponse{
			Success: false,
			Message: "userNotFound",
		}, nil
	}

	// ユーザー名を更新
	userDB.Name = req.GetNewName()
	userDB.UpdatedAt = time.Now().Unix()

	if err := userDB.Update(ctx, s.DB); err != nil {
		return &g.UpdateUserNameResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	return &g.UpdateUserNameResponse{
		Success: true,
		Message: "usernameUpdateSuccess",
	}, nil
}

func (s *UserEntry) ChangePassword(ctx context.Context, req *g.ChangePasswordRequest) (*g.ChangePasswordResponse, error) {
	// リクエストのバリデーション
	if req.GetCurrentPassword() == "" || req.GetNewPassword() == "" {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "passwordsRequired",
		}, nil
	}

	// 新しいパスワードの強度チェック
	if len(req.GetNewPassword()) < 8 {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "passwordTooShort",
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	// 現在のパスワードを取得して検証
	passwordAuthDB, err := database.UserPasswordAutheByUserID(ctx, s.DB, parsedUserID)
	if err != nil {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "userNotFound",
		}, nil
	}

	// 現在のパスワードの検証
	if err := request.VerifyPassword(req.GetCurrentPassword(), passwordAuthDB.PasswordHashed); err != nil {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "currentPasswordIncorrect",
		}, nil
	}

	// 新しいパスワードをハッシュ化
	hashedNewPassword, err := request.EncryptPassword(req.GetNewPassword())
	if err != nil {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	// パスワードを更新
	passwordAuthDB.PasswordHashed = hashedNewPassword
	passwordAuthDB.UpdatedAt = time.Now().Unix()

	if err := passwordAuthDB.Update(ctx, s.DB); err != nil {
		return &g.ChangePasswordResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	return &g.ChangePasswordResponse{
		Success: true,
		Message: "passwordChangeSuccess",
	}, nil
}

func (s *UserEntry) UpdateLLMKey(ctx context.Context, req *g.UpdateLLMKeyRequest) (*g.UpdateLLMKeyResponse, error) {
	// リクエストのバリデーション
	if req.GetKey() == "" {
		return &g.UpdateLLMKeyResponse{
			Success: false,
			Message: "tokenRequired",
		}, nil
	}

	// トークンの長さチェック（100文字以内）
	if len(req.GetKey()) > 100 {
		return &g.UpdateLLMKeyResponse{
			Success: false,
			Message: "tokenTooLong",
		}, nil
	}

	// プロバイダーの検証
	if req.GetLlmProvider() < 0 {
		return &g.UpdateLLMKeyResponse{
			Success: false,
			Message: "invalidProvider",
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.UpdateLLMKeyResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.UpdateLLMKeyResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	// 既存のLLMトークンを確認
	userLLMDB, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, int16(req.GetLlmProvider()))
	currentTime := time.Now().Unix()

	if err != nil && err != sql.ErrNoRows {
		return &g.UpdateLLMKeyResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	if err == sql.ErrNoRows {
		// 新規作成
		newUserLLM := &database.UserLlm{
			UserID:             parsedUserID,
			LlmProvider:        int16(req.GetLlmProvider()),
			Key:                req.GetKey(),
			AutoSummaryDaily:   false, // デフォルトは無効
			AutoSummaryMonthly: false, // デフォルトは無効
			CreatedAt:          currentTime,
			UpdatedAt:          currentTime,
		}

		if err := newUserLLM.Insert(ctx, s.DB); err != nil {
			return &g.UpdateLLMKeyResponse{
				Success: false,
				Message: "updateFailed",
			}, nil
		}
	} else {
		// 更新
		userLLMDB.Key = req.GetKey()
		userLLMDB.UpdatedAt = currentTime

		if err := userLLMDB.Update(ctx, s.DB); err != nil {
			return &g.UpdateLLMKeyResponse{
				Success: false,
				Message: "updateFailed",
			}, nil
		}
	}

	return &g.UpdateLLMKeyResponse{
		Success: true,
		Message: "llmTokenUpdateSuccess",
	}, nil
}

func (s *UserEntry) GetUserInfo(ctx context.Context, req *g.GetUserInfoRequest) (*g.GetUserInfoResponse, error) {
	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// ユーザー情報を取得
	userDB, err := database.UserByID(ctx, s.DB, parsedUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	// LLMキーを取得（存在する場合）
	var llmKeys []*g.LLMKeyInfo

	// 現在はGemini（provider 1）のみサポート
	userLLM, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, 1)
	if err == nil && userLLM != nil {
		llmKeys = append(llmKeys, &g.LLMKeyInfo{
			LlmProvider:            int32(userLLM.LlmProvider),
			Key:                    userLLM.Key,
			AutoSummaryDaily:       userLLM.AutoSummaryDaily,
			AutoSummaryMonthly:     userLLM.AutoSummaryMonthly,
			AutoLatestTrendEnabled: userLLM.AutoLatestTrendEnabled,
			SemanticSearchEnabled:  userLLM.SemanticSearchEnabled,
		})
	}

	return &g.GetUserInfoResponse{
		Name:    userDB.Name,
		Email:   userDB.Email,
		LlmKeys: llmKeys,
	}, nil
}

func (s *UserEntry) DeleteLLMKey(ctx context.Context, req *g.DeleteLLMKeyRequest) (*g.DeleteLLMKeyResponse, error) {
	// プロバイダーの検証
	if req.GetLlmProvider() < 0 {
		return &g.DeleteLLMKeyResponse{
			Success: false,
			Message: "invalidProvider",
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.DeleteLLMKeyResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.DeleteLLMKeyResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	// 既存のLLMトークンを取得
	userLLMDB, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, int16(req.GetLlmProvider()))
	if err != nil {
		if err == sql.ErrNoRows {
			return &g.DeleteLLMKeyResponse{
				Success: false,
				Message: "tokenNotFound",
			}, nil
		}
		return &g.DeleteLLMKeyResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	// LLMトークンを削除
	if err := userLLMDB.Delete(ctx, s.DB); err != nil {
		return &g.DeleteLLMKeyResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	return &g.DeleteLLMKeyResponse{
		Success: true,
		Message: "llmTokenDeleteSuccess",
	}, nil
}

func (s *UserEntry) DeleteAccount(ctx context.Context, req *g.DeleteAccountRequest) (*g.DeleteAccountResponse, error) {
	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.DeleteAccountResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.DeleteAccountResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	// ユーザーの存在確認
	userDB, err := database.UserByID(ctx, s.DB, parsedUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return &g.DeleteAccountResponse{
				Success: false,
				Message: "userNotFound",
			}, nil
		}
		return &g.DeleteAccountResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	// トランザクション内で関連データを削除
	err = database.RwTransaction(ctx, s.DB, func(tx *sql.Tx) error {
		// 1. 日記データを削除
		if err := database.DeleteDiariesByUserID(ctx, tx, parsedUserID); err != nil {
			return err
		}

		// 2. LLMトークンを削除
		if err := database.DeleteUserLLMsByUserID(ctx, tx, parsedUserID); err != nil {
			return err
		}

		// 3. パスワード認証を削除
		if err := database.DeleteUserPasswordAuthesByUserID(ctx, tx, parsedUserID); err != nil {
			return err
		}

		// 4. ユーザーを削除
		if err := userDB.Delete(ctx, tx); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	})
	if err != nil {
		return &g.DeleteAccountResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	return &g.DeleteAccountResponse{
		Success: true,
		Message: "accountDeleteSuccess",
	}, nil
}

func (s *UserEntry) UpdateAutoSummarySettings(ctx context.Context, req *g.UpdateAutoSummarySettingsRequest) (*g.UpdateAutoSummarySettingsResponse, error) {
	// プロバイダーの検証
	if req.GetLlmProvider() < 0 {
		return &g.UpdateAutoSummarySettingsResponse{
			Success: false,
			Message: "invalidProvider",
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.UpdateAutoSummarySettingsResponse{
			Success: false,
			Message: "unauthorized",
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.UpdateAutoSummarySettingsResponse{
			Success: false,
			Message: "invalidUserId",
		}, nil
	}

	// 既存のLLM設定を取得
	userLLMDB, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, int16(req.GetLlmProvider()))
	if err != nil {
		if err == sql.ErrNoRows {
			return &g.UpdateAutoSummarySettingsResponse{
				Success: false,
				Message: "llmKeyNotFound",
			}, nil
		}
		return &g.UpdateAutoSummarySettingsResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	// 自動要約設定を更新
	userLLMDB.AutoSummaryDaily = req.GetAutoSummaryDaily()
	userLLMDB.AutoSummaryMonthly = req.GetAutoSummaryMonthly()
	userLLMDB.AutoLatestTrendEnabled = req.GetAutoLatestTrendEnabled()
	userLLMDB.SemanticSearchEnabled = req.GetSemanticSearchEnabled()
	userLLMDB.UpdatedAt = time.Now().Unix()

	if err := userLLMDB.Update(ctx, s.DB); err != nil {
		return &g.UpdateAutoSummarySettingsResponse{
			Success: false,
			Message: "updateFailed",
		}, nil
	}

	return &g.UpdateAutoSummarySettingsResponse{
		Success: true,
		Message: "autoSummarySettingsUpdateSuccess",
	}, nil
}

func (s *UserEntry) GetAutoSummarySettings(ctx context.Context, req *g.GetAutoSummarySettingsRequest) (*g.GetAutoSummarySettingsResponse, error) {
	// プロバイダーの検証
	if req.GetLlmProvider() < 0 {
		return &g.GetAutoSummarySettingsResponse{
			AutoSummaryDaily:       false,
			AutoSummaryMonthly:     false,
			AutoLatestTrendEnabled: false,
		}, nil
	}

	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return &g.GetAutoSummarySettingsResponse{
			AutoSummaryDaily:       false,
			AutoSummaryMonthly:     false,
			AutoLatestTrendEnabled: false,
		}, nil
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return &g.GetAutoSummarySettingsResponse{
			AutoSummaryDaily:       false,
			AutoSummaryMonthly:     false,
			AutoLatestTrendEnabled: false,
		}, nil
	}

	// LLM設定を取得
	userLLMDB, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, parsedUserID, int16(req.GetLlmProvider()))
	if err != nil {
		// 設定が存在しない場合はデフォルト値を返す
		return &g.GetAutoSummarySettingsResponse{
			AutoSummaryDaily:       false,
			AutoSummaryMonthly:     false,
			AutoLatestTrendEnabled: false,
		}, nil
	}

	return &g.GetAutoSummarySettingsResponse{
		AutoSummaryDaily:       userLLMDB.AutoSummaryDaily,
		AutoSummaryMonthly:     userLLMDB.AutoSummaryMonthly,
		AutoLatestTrendEnabled: userLLMDB.AutoLatestTrendEnabled,
		SemanticSearchEnabled:  userLLMDB.SemanticSearchEnabled,
	}, nil
}

func (s *UserEntry) GetPubSubMetrics(ctx context.Context, req *g.GetPubSubMetricsRequest) (*g.GetPubSubMetricsResponse, error) {
	// コンテキストからユーザーIDを取得
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unauthorized")
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// 1. 過去24時間の時間別メトリクスを生成
	hourlyMetrics, err := s.getHourlyMetrics(ctx, parsedUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get hourly metrics: %w", err)
	}

	// 2. 現在処理中のタスクを取得
	processingTasks, err := s.getProcessingTasks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get processing tasks: %w", err)
	}

	// 3. 統計情報を取得
	summary, err := s.getMetricsSummary(ctx, parsedUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics summary: %w", err)
	}

	return &g.GetPubSubMetricsResponse{
		HourlyMetrics:   hourlyMetrics,
		ProcessingTasks: processingTasks,
		Summary:         summary,
	}, nil
}

func (s *UserEntry) getHourlyMetrics(ctx context.Context, userID uuid.UUID) ([]*g.HourlyMetrics, error) {
	// 過去24時間のデータを1時間ごとに集約
	// 注: トレンド生成は現状Redisのみに保存されるため、hourly_metricsでは0になる
	// 将来的にトレンド生成履歴をDBに保存する場合は、ここにクエリを追加する
	rawMetrics, err := database.HourlyPubSubMetrics(ctx, s.DB, userID)
	if err != nil {
		return nil, err
	}

	metrics := make([]*g.HourlyMetrics, 0, len(rawMetrics))
	for _, m := range rawMetrics {
		metrics = append(metrics, &g.HourlyMetrics{
			Timestamp:                 m.Hour.Unix(),
			DailySummariesProcessed:   m.DailySummariesProcessed,
			MonthlySummariesProcessed: m.MonthlySummariesProcessed,
			DailySummariesFailed:      0, // TODO: 失敗ログを記録する仕組みを追加後に実装
			MonthlySummariesFailed:    0, // TODO: 失敗ログを記録する仕組みを追加後に実装
			LatestTrendsProcessed:     0, // トレンド生成履歴はDBに保存されていないため0
			LatestTrendsFailed:        0, // トレンド生成履歴はDBに保存されていないため0
			DiaryEmbeddingsProcessed:  m.EmbeddingsProcessed,
			DiaryEmbeddingsFailed:     0, // TODO: 失敗ログを記録する仕組みを追加後に実装
			SemanticSearchesProcessed: m.SemanticSearchesProcessed,
		})
	}

	return metrics, nil
}

func (s *UserEntry) getProcessingTasks(ctx context.Context, userID string) ([]*g.ProcessingTask, error) {
	// Redisから現在処理中のタスクを取得
	var tasks []*g.ProcessingTask

	// 日次サマリータスクの検索
	dailyPattern := fmt.Sprintf("task:daily_summary:%s:*", userID)
	dailyCmd := s.RedisClient.B().Keys().Pattern(dailyPattern).Build()
	dailyKeys, err := s.RedisClient.Do(ctx, dailyCmd).AsStrSlice()
	if err == nil {
		for _, key := range dailyKeys {
			// キーから日付を抽出: task:daily_summary:userID:YYYY-MM-DD
			parts := strings.Split(key, ":")
			if len(parts) >= 4 {
				date := parts[3]
				// タスクの開始時刻は現在時刻から推定（より正確には開始時刻をRedisに保存すべき）
				tasks = append(tasks, &g.ProcessingTask{
					TaskType:  "daily_summary",
					Date:      date,
					StartedAt: time.Now().Add(-time.Minute * 5).Unix(), // 推定値
				})
			}
		}
	}

	// 月次サマリータスクの検索
	monthlyPattern := fmt.Sprintf("task:monthly_summary:%s:*", userID)
	monthlyCmd := s.RedisClient.B().Keys().Pattern(monthlyPattern).Build()
	monthlyKeys, err := s.RedisClient.Do(ctx, monthlyCmd).AsStrSlice()
	if err == nil {
		for _, key := range monthlyKeys {
			// キーから年月を抽出: task:monthly_summary:userID:YYYY-MM
			parts := strings.Split(key, ":")
			if len(parts) >= 4 {
				yearMonth := parts[3]
				tasks = append(tasks, &g.ProcessingTask{
					TaskType:  "monthly_summary",
					Date:      yearMonth,
					StartedAt: time.Now().Add(-time.Minute * 5).Unix(), // 推定値
				})
			}
		}
	}

	// トレンド分析タスクの検索
	trendTaskKey := fmt.Sprintf("task:latest_trend:%s", userID)
	trendTaskCmd := s.RedisClient.B().Exists().Key(trendTaskKey).Build()
	exists, err := s.RedisClient.Do(ctx, trendTaskCmd).AsInt64()
	if err == nil && exists > 0 {
		// Redisからタスクの開始時刻を取得
		getCmd := s.RedisClient.B().Get().Key(trendTaskKey).Build()
		startTimeStr, err := s.RedisClient.Do(ctx, getCmd).ToString()
		var startedAt int64
		if err == nil {
			startedAt, _ = strconv.ParseInt(startTimeStr, 10, 64)
		} else {
			// フォールバック: 取得できない場合は5分前を推定値として使用
			startedAt = time.Now().Add(-time.Minute * 5).Unix()
		}

		tasks = append(tasks, &g.ProcessingTask{
			TaskType:  "latest_trend",
			Date:      "直近3日",
			StartedAt: startedAt,
		})
	}

	return tasks, nil
}

func (s *UserEntry) getMetricsSummary(ctx context.Context, userID uuid.UUID) (*g.MetricsSummary, error) {
	// 日次サマリー総数を取得
	totalDaily, err := database.TotalDailySummaryCount(ctx, s.DB, userID)
	if err != nil {
		return nil, err
	}

	// 月次サマリー総数を取得
	totalMonthly, err := database.TotalMonthlySummaryCount(ctx, s.DB, userID)
	if err != nil {
		return nil, err
	}

	// 未作成の日次サマリー数を取得（今日を除く）
	pendingDaily, err := database.PendingDailySummaryCount(ctx, s.DB, userID)
	if err != nil {
		return nil, err
	}

	// 未作成の月次サマリー数を取得（今月を除く）
	pendingMonthly, err := database.PendingMonthlySummaryCount(ctx, s.DB, userID)
	if err != nil {
		return nil, err
	}

	// 自動要約設定とRAG設定を取得
	autoSettings, err := database.UserLLMAutoSettingsByUserID(ctx, s.DB, userID)
	if err != nil {
		return nil, err
	}

	// embeddingチャンク数を取得
	totalEmbeddings, err := database.TotalEmbeddingCount(ctx, s.DB, userID)
	if err != nil {
		return nil, err
	}

	// embedding生成済み日記数を取得
	totalEmbeddingDiaries, err := database.TotalEmbeddingDiaryCount(ctx, s.DB, userID)
	if err != nil {
		return nil, err
	}

	// embedding未生成の日記数を取得
	pendingEmbeddings, err := database.PendingEmbeddingCount(ctx, s.DB, userID)
	if err != nil {
		return nil, err
	}

	// 最新トレンド生成日時をRedisから取得
	var latestTrendGeneratedAt string
	trendKey := fmt.Sprintf("latest_trend:%s", userID)
	getCmd := s.RedisClient.B().Get().Key(trendKey).Build()
	result := s.RedisClient.Do(ctx, getCmd)

	// エラーハンドリングを改善（goto文の代わりに早期リターンパターンを使用）
	if result.Error() == nil {
		if trendDataStr, err := result.ToString(); err == nil {
			// JSONをパース
			var trendData struct {
				GeneratedAt string `json:"generated_at"`
			}
			if err := json.Unmarshal([]byte(trendDataStr), &trendData); err == nil {
				latestTrendGeneratedAt = trendData.GeneratedAt
			}
		}
	}

	return &g.MetricsSummary{
		TotalDailySummaries:       totalDaily,
		TotalMonthlySummaries:     totalMonthly,
		PendingDailySummaries:     pendingDaily,
		PendingMonthlySummaries:   pendingMonthly,
		AutoSummaryDailyEnabled:   autoSettings.AutoSummaryDaily,
		AutoSummaryMonthlyEnabled: autoSettings.AutoSummaryMonthly,
		AutoLatestTrendEnabled:    autoSettings.AutoLatestTrend,
		LatestTrendGeneratedAt:    latestTrendGeneratedAt,
		SemanticSearchEnabled:     autoSettings.SemanticSearchEnabled,
		TotalEmbeddings:           totalEmbeddings,
		TotalEmbeddingDiaries:     totalEmbeddingDiaries,
		PendingEmbeddings:         pendingEmbeddings,
	}, nil
}
