package diary

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/redis/rueidis"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LLMFactory はLLMクライアントを作成するファクトリインターフェース
type LLMFactory interface {
	CreateGeminiClient(ctx context.Context, apiKey string) (GeminiEmbedder, error)
}

// GeminiEmbedder はGemini埋め込みAPIクライアントのインターフェース
type GeminiEmbedder interface {
	GenerateEmbedding(ctx context.Context, text string, isDocument bool) ([]float32, error)
	Close() error
}

type DiaryEntry struct {
	g.UnimplementedDiaryServiceServer
	DB         database.DB
	Redis      rueidis.Client
	LLMFactory LLMFactory
}

type SummaryGenerationMessage struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	Date   string `json:"date"` // YYYY-MM-DD format
}

type MonthlySummaryGenerationMessage struct {
	Type   string `json:"type"`
	UserID string `json:"user_id"`
	Year   int    `json:"year"`
	Month  int    `json:"month"`
}

type DiaryHighlightGenerationMessage struct {
	Type    string `json:"type"`
	UserID  string `json:"user_id"`
	DiaryID string `json:"diary_id"`
}

type DiaryEmbeddingMessage struct {
	Type    string `json:"type"`
	UserID  string `json:"user_id"`
	DiaryID string `json:"diary_id"`
}

// getTaskTimeout 環境変数からタスクタイムアウトを取得(デフォルト600秒)
func getTaskTimeout() int {
	timeoutStr := os.Getenv("TASK_TIMEOUT_SECONDS")
	if timeoutStr == "" {
		return 600 // デフォルト10分
	}
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil || timeout <= 0 {
		return 600 // パースエラー時もデフォルト
	}
	return timeout
}

// タスクの状態を管理するヘルパー関数
func (s *DiaryEntry) setTaskStatus(ctx context.Context, taskKey, status string, expireSeconds int) error {
	setCmd := s.Redis.B().Set().Key(taskKey).Value(status).Ex(time.Duration(expireSeconds) * time.Second).Build()
	return s.Redis.Do(ctx, setCmd).Error()
}

func (s *DiaryEntry) getTaskStatus(ctx context.Context, taskKey string) (string, error) {
	getCmd := s.Redis.B().Get().Key(taskKey).Build()
	result := s.Redis.Do(ctx, getCmd)
	if result.Error() != nil {
		return "", result.Error()
	}
	return result.ToString()
}

func (s *DiaryEntry) deleteTaskStatus(ctx context.Context, taskKey string) error {
	delCmd := s.Redis.B().Del().Key(taskKey).Build()
	return s.Redis.Do(ctx, delCmd).Error()
}

func (s *DiaryEntry) CreateDiaryEntry(
	ctx context.Context,
	message *g.CreateDiaryEntryRequest,
) (*g.CreateDiaryEntryResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	id := uuid.New()
	currentTime := time.Now().Unix()
	date := time.Date(int(message.Date.Year), time.Month(message.Date.Month), int(message.Date.Day), 0, 0, 0, 0, time.UTC)

	diary := &database.Diary{
		ID:        id,
		UserID:    userID,
		Content:   message.Content,
		Date:      date,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	// トランザクション内でdiaryを保存
	err = database.RwTransaction(ctx, s.DB.(*sql.DB), func(tx *sql.Tx) error {
		if err := diary.Insert(ctx, tx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 非同期で埋め込みベクトルを生成（Redis Pub/Sub経由）
	s.publishDiaryEmbeddingMessage(ctx, userID.String(), diary.ID.String())

	return &g.CreateDiaryEntryResponse{
		Entry: &g.DiaryEntry{
			Id:        diary.ID.String(),
			Date:      message.Date,
			Content:   diary.Content,
			CreatedAt: diary.CreatedAt,
			UpdatedAt: diary.UpdatedAt,
		},
	}, nil
}

func (s *DiaryEntry) GetDiaryEntry(
	ctx context.Context,
	message *g.GetDiaryEntryRequest,
) (*g.GetDiaryEntryResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	date := time.Date(int(message.Date.Year), time.Month(message.Date.Month), int(message.Date.Day), 0, 0, 0, 0, time.UTC)

	diary, err := database.DiaryByUserIDDate(ctx, s.DB, userID, date)
	if err != nil {
		return nil, err
	}

	return &g.GetDiaryEntryResponse{
		Entry: &g.DiaryEntry{
			Id:        diary.ID.String(),
			Date:      message.Date,
			Content:   diary.Content,
			CreatedAt: diary.CreatedAt,
			UpdatedAt: diary.UpdatedAt,
		},
	}, nil
}

func (s *DiaryEntry) GetDiaryEntries(
	ctx context.Context,
	message *g.GetDiaryEntriesRequest,
) (*g.GetDiaryEntriesResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// 日記を収集
	type diaryWithDate struct {
		diary   *database.Diary
		dateMsg *g.YMD
	}
	diariesWithDates := make([]diaryWithDate, 0)

	for _, dateMsg := range message.Dates {
		date := time.Date(int(dateMsg.Year), time.Month(dateMsg.Month), int(dateMsg.Day), 0, 0, 0, 0, time.UTC)
		diary, err := database.DiaryByUserIDDate(ctx, s.DB, userID, date)
		if err != nil {
			continue // Skip entries that don't exist
		}
		diariesWithDates = append(diariesWithDates, diaryWithDate{
			diary:   diary,
			dateMsg: dateMsg,
		})
	}

	entries := make([]*g.DiaryEntry, 0, len(diariesWithDates))
	for _, dwd := range diariesWithDates {
		entries = append(entries, &g.DiaryEntry{
			Id:        dwd.diary.ID.String(),
			Date:      dwd.dateMsg,
			Content:   dwd.diary.Content,
			CreatedAt: dwd.diary.CreatedAt,
			UpdatedAt: dwd.diary.UpdatedAt,
		})
	}

	return &g.GetDiaryEntriesResponse{
		Entries: entries,
	}, nil
}

func (s *DiaryEntry) GetDiaryEntriesByMonth(
	ctx context.Context,
	message *g.GetDiaryEntriesByMonthRequest,
) (*g.GetDiaryEntriesByMonthResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// Query each day of the month to find diary entries
	diaries := make([]*database.Diary, 0)

	// Get the number of days in the month
	daysInMonth := time.Date(int(message.Month.Year), time.Month(message.Month.Month)+1, 0, 0, 0, 0, 0, time.UTC).Day()

	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(int(message.Month.Year), time.Month(message.Month.Month), day, 0, 0, 0, 0, time.UTC)
		diary, err := database.DiaryByUserIDDate(ctx, s.DB, userID, date)
		if err != nil {
			continue // Skip entries that don't exist
		}
		diaries = append(diaries, diary)
	}

	entries := make([]*g.DiaryEntry, 0, len(diaries))
	for _, diary := range diaries {
		entries = append(entries, &g.DiaryEntry{
			Id:        diary.ID.String(),
			Date:      &g.YMD{Year: uint32(diary.Date.Year()), Month: uint32(diary.Date.Month()), Day: uint32(diary.Date.Day())},
			Content:   diary.Content,
			CreatedAt: diary.CreatedAt,
			UpdatedAt: diary.UpdatedAt,
		})
	}

	return &g.GetDiaryEntriesByMonthResponse{
		Entries: entries,
	}, nil
}

func (s *DiaryEntry) UpdateDiaryEntry(
	ctx context.Context,
	message *g.UpdateDiaryEntryRequest,
) (*g.UpdateDiaryEntryResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	diaryID, err := uuid.Parse(message.Id)
	if err != nil {
		return nil, err
	}

	diary, err := database.DiaryByID(ctx, s.DB, diaryID)
	if err != nil {
		return nil, err
	}

	// Check if the diary belongs to the authenticated user
	if diary.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "not authorized to update this diary entry")
	}

	// トランザクション内で日記を更新
	err = database.RwTransaction(ctx, s.DB.(*sql.DB), func(tx *sql.Tx) error {
		diary.Content = message.Content
		if message.Date != nil {
			diary.Date = time.Date(int(message.Date.Year), time.Month(message.Date.Month), int(message.Date.Day), 0, 0, 0, 0, time.UTC)
		}
		currentTime := time.Now().Unix()
		diary.UpdatedAt = currentTime

		if err := diary.Update(ctx, tx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 非同期で埋め込みベクトルを再生成（Redis Pub/Sub経由）
	s.publishDiaryEmbeddingMessage(ctx, userID.String(), diary.ID.String())

	return &g.UpdateDiaryEntryResponse{
		Entry: &g.DiaryEntry{
			Id:        diary.ID.String(),
			Date:      &g.YMD{Year: uint32(diary.Date.Year()), Month: uint32(diary.Date.Month()), Day: uint32(diary.Date.Day())},
			Content:   diary.Content,
			CreatedAt: diary.CreatedAt,
			UpdatedAt: diary.UpdatedAt,
		},
	}, nil
}

func (s *DiaryEntry) DeleteDiaryEntry(
	ctx context.Context,
	message *g.DeleteDiaryEntryRequest,
) (*g.DeleteDiaryEntryResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	diaryID, err := uuid.Parse(message.Id)
	if err != nil {
		return nil, err
	}

	diary, err := database.DiaryByID(ctx, s.DB, diaryID)
	if err != nil {
		return nil, err
	}

	// Check if the diary belongs to the authenticated user
	if diary.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "not authorized to delete this diary entry")
	}

	// トランザクション内で日記を削除
	err = database.RwTransaction(ctx, s.DB.(*sql.DB), func(tx *sql.Tx) error {
		return diary.Delete(ctx, tx)
	})
	if err != nil {
		return nil, err
	}

	return &g.DeleteDiaryEntryResponse{
		Success: true,
	}, nil
}

func (s *DiaryEntry) SearchDiaryEntries(
	ctx context.Context,
	message *g.SearchDiaryEntriesRequest,
) (*g.SearchDiaryEntriesResponse, error) {
	// 認証されたユーザーIDを取得（リクエストのuserIDは無視）
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// エンティティ名・エイリアスに基づいて関連キーワードを展開
	expandedKeywords, err := database.RelatedKeywordsByUserIDAndKeyword(ctx, s.DB, userID.String(), message.Keyword)
	if err != nil {
		return nil, err
	}

	var ds []*database.Diary
	if len(expandedKeywords) == 0 {
		// 展開キーワードがない場合は通常検索
		ds, err = database.DiariesByUserIDAndContent(ctx, s.DB, userID.String(), message.Keyword)
	} else {
		// 展開キーワードがある場合は全キーワードでOR検索
		allKeywords := append([]string{message.Keyword}, expandedKeywords...)
		ds, err = database.DiariesByUserIDAndKeywords(ctx, s.DB, userID.String(), allKeywords)
	}
	if err != nil {
		return nil, err
	}

	entries := make([]*g.DiaryEntry, 0, len(ds))
	for _, d := range ds {
		entries = append(entries, &g.DiaryEntry{
			Id:        d.ID.String(),
			Content:   d.Content,
			Date:      &g.YMD{Year: uint32(d.Date.Year()), Month: uint32(d.Date.Month()), Day: uint32(d.Date.Day())},
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
		})
	}
	return &g.SearchDiaryEntriesResponse{
		SearchedKeyword:  message.Keyword,
		Entries:          entries,
		ExpandedKeywords: expandedKeywords,
	}, nil
}

func (s *DiaryEntry) GenerateMonthlySummary(
	ctx context.Context,
	message *g.GenerateMonthlySummaryRequest,
) (*g.GenerateMonthlySummaryResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// ユーザーのLLMキーが設定されているかチェック
	_, err = database.UserLlmByUserIDLlmProvider(ctx, s.DB, userID, 1)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Gemini API key not found for user")
	}

	// 指定された月が今月より前であることを確認
	now := time.Now()
	requestedMonth := time.Date(int(message.Month.Year), time.Month(message.Month.Month), 1, 0, 0, 0, 0, time.UTC)
	currentMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	if !requestedMonth.Before(currentMonth) {
		return nil, status.Errorf(codes.FailedPrecondition, "Monthly summary generation is only allowed for past months")
	}

	// その月に日記が存在するかチェック
	var count int
	checkQuery := `
		SELECT COUNT(*) FROM diaries
		WHERE user_id = $1
		AND EXTRACT(YEAR FROM date) = $2
		AND EXTRACT(MONTH FROM date) = $3
	`
	err = s.DB.(*sql.DB).QueryRow(checkQuery, userID, message.Month.Year, message.Month.Month).Scan(&count)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to check diary entries")
	}
	if count == 0 {
		return nil, status.Errorf(codes.NotFound, "no diary entries found for the specified month")
	}

	// タスクキーを生成
	taskKey := fmt.Sprintf("task:monthly_summary:%s:%d-%d", userID.String(), message.Month.Year, message.Month.Month)

	// 既にタスクが実行中かチェック
	taskStatus, err := s.getTaskStatus(ctx, taskKey)
	if err == nil && (taskStatus == "queued" || taskStatus == "processing") {
		// 既存の要約があるかチェック（レスポンス用）
		existingSummary, err := database.DiarySummaryMonthByUserIDYearMonth(ctx, s.DB, userID, int(message.Month.Year), int(message.Month.Month))
		if err != nil {
			return &g.GenerateMonthlySummaryResponse{
				Summary: &g.MonthlySummary{
					Id:        "",
					Month:     message.Month,
					Summary:   fmt.Sprintf("Monthly summary generation is %s. Please check back later.", taskStatus),
					CreatedAt: 0,
					UpdatedAt: 0,
				},
			}, nil
		} else {
			return &g.GenerateMonthlySummaryResponse{
				Summary: &g.MonthlySummary{
					Id:        existingSummary.ID.String(),
					Month:     message.Month,
					Summary:   fmt.Sprintf("%s (%s)", existingSummary.Summary, taskStatus),
					CreatedAt: existingSummary.CreatedAt,
					UpdatedAt: existingSummary.UpdatedAt,
				},
			}, nil
		}
	}

	// タスクを「キューに追加済み」としてマーク
	timeout := getTaskTimeout()
	if err := s.setTaskStatus(ctx, taskKey, "queued", timeout); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to set task status")
	}

	// Redis Pub/Sub経由で月次要約生成を依頼
	monthlyMessage := MonthlySummaryGenerationMessage{
		Type:   "monthly_summary",
		UserID: userID.String(),
		Year:   int(message.Month.Year),
		Month:  int(message.Month.Month),
	}

	messageBytes, err := json.Marshal(monthlyMessage)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create monthly summary generation request")
	}

	// Redisにメッセージを送信
	publishCmd := s.Redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
	if err := s.Redis.Do(ctx, publishCmd).Error(); err != nil {
		// タスクステータスをクリア
		_ = s.deleteTaskStatus(ctx, taskKey)
		return nil, status.Errorf(codes.Internal, "Failed to queue monthly summary generation")
	}

	// 既存の要約があるかチェック（レスポンス用）
	existingSummary, err := database.DiarySummaryMonthByUserIDYearMonth(ctx, s.DB, userID, int(message.Month.Year), int(message.Month.Month))
	if err != nil {
		// 既存の要約がない場合は、現在処理中であることを示すレスポンスを返す
		return &g.GenerateMonthlySummaryResponse{
			Summary: &g.MonthlySummary{
				Id:        "",
				Month:     message.Month,
				Summary:   "Monthly summary generation is queued. Please check back later.",
				CreatedAt: 0,
				UpdatedAt: 0,
			},
		}, nil
	} else {
		// 既存の要約がある場合は、現在のものを返す（間もなく更新される予定）
		return &g.GenerateMonthlySummaryResponse{
			Summary: &g.MonthlySummary{
				Id:        existingSummary.ID.String(),
				Month:     message.Month,
				Summary:   existingSummary.Summary + " (Updating...)",
				CreatedAt: existingSummary.CreatedAt,
				UpdatedAt: existingSummary.UpdatedAt,
			},
		}, nil
	}
}

func (s *DiaryEntry) GetMonthlySummary(
	ctx context.Context,
	message *g.GetMonthlySummaryRequest,
) (*g.GetMonthlySummaryResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// タスクの状態をチェック（月次要約タスクの状態確認）
	taskKey := fmt.Sprintf("task:monthly_summary:%s:%d-%d", userID.String(), message.Month.Year, message.Month.Month)
	taskStatus, taskErr := s.getTaskStatus(ctx, taskKey)

	// 既存の要約を取得
	summary, summaryErr := database.DiarySummaryMonthByUserIDYearMonth(ctx, s.DB, userID, int(message.Month.Year), int(message.Month.Month))

	// タスクが実行中の場合は状態メッセージを返す
	if taskErr == nil && (taskStatus == "queued" || taskStatus == "processing") {
		if summaryErr != nil {
			// 要約がまだ存在しない場合、状態メッセージのみ
			return &g.GetMonthlySummaryResponse{
				Summary: &g.MonthlySummary{
					Id:        "",
					Month:     message.Month,
					Summary:   fmt.Sprintf("Monthly summary generation is %s. Please check back later.", taskStatus),
					CreatedAt: 0,
					UpdatedAt: 0,
				},
			}, nil
		} else {
			// 既存の要約があるが更新中の場合、要約に状態を付加
			return &g.GetMonthlySummaryResponse{
				Summary: &g.MonthlySummary{
					Id:        summary.ID.String(),
					Month:     message.Month,
					Summary:   fmt.Sprintf("%s (Updating)", summary.Summary),
					CreatedAt: summary.CreatedAt,
					UpdatedAt: summary.UpdatedAt,
				},
			}, nil
		}
	}

	// タスクが実行中でない場合、通常の要約取得処理
	if summaryErr != nil {
		return nil, status.Errorf(codes.NotFound, "summary not found for the specified month")
	}

	return &g.GetMonthlySummaryResponse{
		Summary: &g.MonthlySummary{
			Id:        summary.ID.String(),
			Month:     message.Month,
			Summary:   summary.Summary,
			CreatedAt: summary.CreatedAt,
			UpdatedAt: summary.UpdatedAt,
		},
	}, nil
}

func (s *DiaryEntry) GenerateDailySummary(
	ctx context.Context,
	req *g.GenerateDailySummaryRequest,
) (*g.GenerateDailySummaryResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// 日記IDから日記を取得
	diaryID, err := uuid.Parse(req.DiaryId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid diary ID")
	}

	diary, err := database.DiaryByID(ctx, s.DB, diaryID)
	if err != nil {
		return nil, err
	}

	// 日記の所有者確認
	if diary.UserID != userID {
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	// 日記が当日のものでないことを確認
	today := time.Now().UTC().Truncate(24 * time.Hour)
	diaryDate := diary.Date.UTC().Truncate(24 * time.Hour)
	if !diaryDate.Before(today) {
		return nil, status.Error(codes.FailedPrecondition, "Summary generation is only allowed for past diary entries")
	}

	// 文字数チェック
	if len([]rune(diary.Content)) < 1000 {
		return nil, status.Error(codes.FailedPrecondition, "Content too short for summary generation (minimum 1000 characters)")
	}

	// ユーザーのLLMキーが設定されているかチェック
	_, err = database.UserLlmByUserIDLlmProvider(ctx, s.DB, userID, 1) // Gemini
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "Gemini API key not configured")
	}

	// タスクキーを生成
	taskKey := fmt.Sprintf("task:daily_summary:%s:%s", userID.String(), diary.Date.Format("2006-01-02"))

	// 既にタスクが実行中かチェック
	taskStatus, err := s.getTaskStatus(ctx, taskKey)
	if err == nil && (taskStatus == "queued" || taskStatus == "processing") {
		// 既存の要約があるかチェック（レスポンス用）
		existingSummary, err := database.DiarySummaryDayByUserIDDate(ctx, s.DB, userID, diary.Date)
		if err != nil {
			return &g.GenerateDailySummaryResponse{
				Summary: &g.DailySummary{
					Id:      "",
					DiaryId: diaryID.String(),
					Date: &g.YMD{
						Year:  uint32(diary.Date.Year()),
						Month: uint32(diary.Date.Month()),
						Day:   uint32(diary.Date.Day()),
					},
					Summary:   fmt.Sprintf("Summary generation is %s. Please check back later.", taskStatus),
					CreatedAt: 0,
					UpdatedAt: 0,
				},
			}, nil
		} else {
			return &g.GenerateDailySummaryResponse{
				Summary: &g.DailySummary{
					Id:      existingSummary.ID.String(),
					DiaryId: diaryID.String(),
					Date: &g.YMD{
						Year:  uint32(diary.Date.Year()),
						Month: uint32(diary.Date.Month()),
						Day:   uint32(diary.Date.Day()),
					},
					Summary:   fmt.Sprintf("%s (%s)", existingSummary.Summary, taskStatus),
					CreatedAt: existingSummary.CreatedAt,
					UpdatedAt: existingSummary.UpdatedAt,
				},
			}, nil
		}
	}

	// タスクを「キューに追加済み」としてマーク
	timeout := getTaskTimeout()
	if err := s.setTaskStatus(ctx, taskKey, "queued", timeout); err != nil {
		return nil, status.Error(codes.Internal, "Failed to set task status")
	}

	// Redis Pub/Sub経由で要約生成を依頼
	message := SummaryGenerationMessage{
		Type:   "daily_summary",
		UserID: userID.String(),
		Date:   diary.Date.Format("2006-01-02"),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to create summary generation request")
	}

	// Redisにメッセージを送信
	publishCmd := s.Redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
	if err := s.Redis.Do(ctx, publishCmd).Error(); err != nil {
		// タスクステータスをクリア
		_ = s.deleteTaskStatus(ctx, taskKey)
		return nil, status.Error(codes.Internal, "Failed to queue summary generation")
	}

	// 既存の要約があるかチェック（レスポンス用）
	existingSummary, err := database.DiarySummaryDayByUserIDDate(ctx, s.DB, userID, diary.Date)
	if err != nil {
		// 既存の要約がない場合は、現在処理中であることを示すレスポンスを返す
		return &g.GenerateDailySummaryResponse{
			Summary: &g.DailySummary{
				Id:      "",
				DiaryId: diaryID.String(),
				Date: &g.YMD{
					Year:  uint32(diary.Date.Year()),
					Month: uint32(diary.Date.Month()),
					Day:   uint32(diary.Date.Day()),
				},
				Summary:   "Summary generation is queued. Please check back later.",
				CreatedAt: 0,
				UpdatedAt: 0,
			},
		}, nil
	} else {
		// 既存の要約がある場合は、現在のものを返す（間もなく更新される予定）
		return &g.GenerateDailySummaryResponse{
			Summary: &g.DailySummary{
				Id:      existingSummary.ID.String(),
				DiaryId: diaryID.String(),
				Date: &g.YMD{
					Year:  uint32(diary.Date.Year()),
					Month: uint32(diary.Date.Month()),
					Day:   uint32(diary.Date.Day()),
				},
				Summary:   existingSummary.Summary + " (Updating...)",
				CreatedAt: existingSummary.CreatedAt,
				UpdatedAt: existingSummary.UpdatedAt,
			},
		}, nil
	}
}

func (s *DiaryEntry) GetDailySummary(
	ctx context.Context,
	req *g.GetDailySummaryRequest,
) (*g.GetDailySummaryResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// 日付から要約を直接取得
	date := time.Date(int(req.Date.Year), time.Month(req.Date.Month), int(req.Date.Day), 0, 0, 0, 0, time.UTC)

	// タスクの状態をチェック（日記要約タスクの状態確認）
	dateStr := date.Format("2006-01-02")
	taskKey := fmt.Sprintf("task:daily_summary:%s:%s", userID.String(), dateStr)
	taskStatus, taskErr := s.getTaskStatus(ctx, taskKey)

	// 既存の要約を取得
	summary, summaryErr := database.DiarySummaryDayByUserIDDate(ctx, s.DB, userID, date)

	// タスクが実行中の場合は状態メッセージを返す
	if taskErr == nil && (taskStatus == "queued" || taskStatus == "processing") {
		if summaryErr != nil {
			// 要約がまだ存在しない場合、状態メッセージのみ
			return &g.GetDailySummaryResponse{
				Summary: &g.DailySummary{
					Id:        "",
					DiaryId:   "",
					Date:      req.Date,
					Summary:   fmt.Sprintf("Summary generation is %s. Please check back later.", taskStatus),
					CreatedAt: 0,
					UpdatedAt: 0,
				},
			}, nil
		} else {
			// 既存の要約があるが更新中の場合、要約に状態を付加
			return &g.GetDailySummaryResponse{
				Summary: &g.DailySummary{
					Id:        summary.ID.String(),
					DiaryId:   "",
					Date:      req.Date,
					Summary:   fmt.Sprintf("%s (Updating)", summary.Summary),
					CreatedAt: summary.CreatedAt,
					UpdatedAt: summary.UpdatedAt,
				},
			}, nil
		}
	}

	// タスクが実行中でない場合、通常の要約取得処理
	if summaryErr != nil {
		return nil, status.Error(codes.NotFound, "Daily summary not found")
	}

	return &g.GetDailySummaryResponse{
		Summary: &g.DailySummary{
			Id:        summary.ID.String(),
			DiaryId:   "", // DiarySummaryDayにはdiaryIdがないので空文字
			Date:      req.Date,
			Summary:   summary.Summary,
			CreatedAt: summary.CreatedAt,
			UpdatedAt: summary.UpdatedAt,
		},
	}, nil
}

// TriggerDiaryHighlight 日記エントリのハイライト生成を非同期でトリガー
func (s *DiaryEntry) TriggerDiaryHighlight(
	ctx context.Context,
	req *g.TriggerDiaryHighlightRequest,
) (*g.TriggerDiaryHighlightResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// 日記IDをパース
	diaryID, err := uuid.Parse(req.DiaryId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid diary ID")
	}

	// 日記を取得
	diary, err := database.DiaryByID(ctx, s.DB, diaryID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Diary entry not found")
	}

	// 日記の所有者確認
	if diary.UserID != userID {
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	// 文字数チェック（最小500文字）
	if len([]rune(diary.Content)) < 500 {
		return nil, status.Error(codes.FailedPrecondition, "Content too short for highlight generation (minimum 500 characters)")
	}

	// ユーザーのLLMキーが設定されているかチェック
	_, err = database.UserLlmByUserIDLlmProvider(ctx, s.DB, userID, 1) // Gemini
	if err != nil {
		return nil, status.Error(codes.NotFound, "Gemini API key not configured")
	}

	// タスクキーを生成
	taskKey := fmt.Sprintf("task:diary_highlight:%s:%s", userID.String(), diaryID.String())

	// 既にタスクが実行中かチェック
	taskStatus, err := s.getTaskStatus(ctx, taskKey)
	if err == nil && (taskStatus == "queued" || taskStatus == "processing") {
		return &g.TriggerDiaryHighlightResponse{
			Queued:  true,
			Message: fmt.Sprintf("Highlight generation is already %s", taskStatus),
		}, nil
	}

	// タスクを「キューに追加済み」としてマーク
	timeout := getTaskTimeout()
	if err := s.setTaskStatus(ctx, taskKey, "queued", timeout); err != nil {
		return nil, status.Error(codes.Internal, "Failed to set task status")
	}

	// Redis Pub/Sub経由でハイライト生成を依頼
	message := DiaryHighlightGenerationMessage{
		Type:    "diary_highlight",
		UserID:  userID.String(),
		DiaryID: diaryID.String(),
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		return nil, status.Error(codes.Internal, "Failed to create highlight generation request")
	}

	// Redisにメッセージを送信
	publishCmd := s.Redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
	if err := s.Redis.Do(ctx, publishCmd).Error(); err != nil {
		// タスクステータスをクリア
		_ = s.deleteTaskStatus(ctx, taskKey)
		return nil, status.Error(codes.Internal, "Failed to queue highlight generation")
	}

	return &g.TriggerDiaryHighlightResponse{
		Queued:  true,
		Message: "Highlight generation has been queued",
	}, nil
}

// GetDiaryHighlight 日記エントリのハイライト情報を取得
func (s *DiaryEntry) GetDiaryHighlight(
	ctx context.Context,
	req *g.GetDiaryHighlightRequest,
) (*g.GetDiaryHighlightResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// 日記IDをパース
	diaryID, err := uuid.Parse(req.DiaryId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid diary ID")
	}

	// 日記を取得
	diary, err := database.DiaryByID(ctx, s.DB, diaryID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Diary entry not found")
	}

	// 日記の所有者確認
	if diary.UserID != userID {
		return nil, status.Error(codes.PermissionDenied, "Access denied")
	}

	// ハイライトを取得
	highlight, err := database.DiaryHighlightByDiaryID(ctx, s.DB, diaryID)
	if err != nil {
		return nil, status.Error(codes.NotFound, "Highlight not found")
	}

	// 注: 日記が更新された場合でもハイライトは返す
	// フロントエンドで diary.UpdatedAt と highlight.UpdatedAt を比較して古いかどうかを判断する
	// これにより、ユーザーは古いハイライトを確認しつつ再生成を選択できる

	// JSONBからハイライト情報を取得
	var highlightsRaw []map[string]any
	if err := json.Unmarshal(highlight.Highlights, &highlightsRaw); err != nil {
		return nil, status.Error(codes.Internal, "Failed to parse highlights")
	}

	// gRPCレスポンス形式に変換
	highlights := make([]*g.HighlightRange, 0, len(highlightsRaw))
	for _, h := range highlightsRaw {
		start, ok1 := h["start"].(float64)
		end, ok2 := h["end"].(float64)
		text, ok3 := h["text"].(string)
		if !ok1 || !ok2 || !ok3 {
			continue
		}

		highlights = append(highlights, &g.HighlightRange{
			Start: int32(start),
			End:   int32(end),
			Text:  text,
		})
	}

	return &g.GetDiaryHighlightResponse{
		Highlights: highlights,
		CreatedAt:  highlight.CreatedAt.Unix(),
		UpdatedAt:  highlight.UpdatedAt.Unix(),
	}, nil
}

// publishDiaryEmbeddingMessage は日記の埋め込みベクトル生成をRedis Pub/Sub経由でキューに追加する
// エラーはログに記録するのみで、レスポンスには影響しない
func (s *DiaryEntry) publishDiaryEmbeddingMessage(ctx context.Context, userID, diaryID string) {
	if s.Redis == nil {
		return
	}
	message := DiaryEmbeddingMessage{
		Type:    "diary_embedding",
		UserID:  userID,
		DiaryID: diaryID,
	}
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return
	}
	publishCmd := s.Redis.B().Publish().Channel("diary_events").Message(string(messageBytes)).Build()
	// エラーはログには出せないが、埋め込み生成の失敗は非クリティカル
	_ = s.Redis.Do(ctx, publishCmd).Error()
}

// SearchDiaryEntriesSemantic 自然言語クエリで日記を意味的に検索する
func (s *DiaryEntry) SearchDiaryEntriesSemantic(
	ctx context.Context,
	req *g.SearchDiaryEntriesSemanticRequest,
) (*g.SearchDiaryEntriesSemanticResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// クエリ検証
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "Query is required")
	}

	// ユーザーのAPIキーと設定を取得
	userLLM, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, userID, 1) // Gemini
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Gemini API key not found")
	}

	// 意味的検索が有効化されているか確認
	if !userLLM.SemanticSearchEnabled {
		return nil, status.Errorf(codes.FailedPrecondition, "Semantic search is not enabled. Please enable it in settings.")
	}

	// LLMファクトリーの確認
	if s.LLMFactory == nil {
		return nil, status.Error(codes.Internal, "LLM factory not configured")
	}

	// Geminiクライアント作成
	geminiClient, err := s.LLMFactory.CreateGeminiClient(ctx, userLLM.Key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create Gemini client")
	}
	defer func() {
		_ = geminiClient.Close()
	}()

	// クエリをベクトル化（クエリ用タスクタイプ）
	queryEmbedding, err := geminiClient.GenerateEmbedding(ctx, req.Query, false)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to generate query embedding: %v", err)
	}

	// 検索件数を決定
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// pgvectorでコサイン類似度ANN検索
	const similarityThreshold = 0.5
	searchResults, err := database.SearchDiaryEntriesByEmbedding(ctx, s.DB, userID, queryEmbedding, limit, similarityThreshold)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to search diary entries: %v", err)
	}

	// 結果を変換
	results := make([]*g.SemanticSearchResult, 0, len(searchResults))
	for _, sr := range searchResults {
		snippet := generateSnippet(sr.Content, 200)
		results = append(results, &g.SemanticSearchResult{
			DiaryId: sr.DiaryID.String(),
			Date: &g.YMD{
				Year:  uint32(sr.Date.Year()),
				Month: uint32(sr.Date.Month()),
				Day:   uint32(sr.Date.Day()),
			},
			Snippet:    snippet,
			Similarity: float32(sr.Similarity),
		})
	}

	return &g.SearchDiaryEntriesSemanticResponse{
		Results: results,
	}, nil
}

// RegenerateAllEmbeddings はembeddingが未生成の全日記をキューに追加する
func (s *DiaryEntry) RegenerateAllEmbeddings(
	ctx context.Context,
	_ *g.RegenerateAllEmbeddingsRequest,
) (*g.RegenerateAllEmbeddingsResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// ユーザーのAPIキーと設定を取得
	userLLM, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, userID, 1) // Gemini
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "Gemini API key not found")
	}

	// 意味的検索が有効化されているか確認
	if !userLLM.SemanticSearchEnabled {
		return nil, status.Errorf(codes.FailedPrecondition, "Semantic search is not enabled. Please enable it in settings.")
	}

	if s.Redis == nil {
		return nil, status.Error(codes.Internal, "Redis not configured")
	}

	// embedding未生成の日記ID一覧を取得
	query := `
		SELECT d.id
		FROM diaries d
		WHERE d.user_id = $1
		  AND NOT EXISTS (
		    SELECT 1 FROM diary_embeddings e WHERE e.diary_id = d.id
		  )
		ORDER BY d.date DESC
	`
	rows, err := s.DB.(*sql.DB).QueryContext(ctx, query, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to query diaries: %v", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			fmt.Printf("Failed to close rows: %v\n", closeErr)
		}
	}()

	// 各日記のembedding生成メッセージをキューに追加
	var count int32
	for rows.Next() {
		var diaryID string
		if err := rows.Scan(&diaryID); err != nil {
			continue
		}
		s.publishDiaryEmbeddingMessage(ctx, userIDStr, diaryID)
		count++
	}

	return &g.RegenerateAllEmbeddingsResponse{
		Success:     true,
		QueuedCount: count,
	}, nil
}

// GetDiaryEmbeddingStatus は指定された日記のRAGインデックス状態を返す
func (s *DiaryEntry) GetDiaryEmbeddingStatus(
	ctx context.Context,
	req *g.GetDiaryEmbeddingStatusRequest,
) (*g.GetDiaryEmbeddingStatusResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	diaryID, err := uuid.Parse(req.DiaryId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid diary ID")
	}

	// 日記が存在し、このユーザーのものかを確認
	diary, err := database.DiaryByID(ctx, s.DB, diaryID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Diary not found")
	}
	if diary.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "Permission denied")
	}

	embeddingStatus, err := database.GetDiaryEmbeddingStatus(ctx, s.DB, diaryID, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get embedding status: %v", err)
	}

	resp := &g.GetDiaryEmbeddingStatusResponse{
		Indexed: embeddingStatus.Indexed,
	}
	if embeddingStatus.Indexed {
		resp.ModelVersion = embeddingStatus.ModelVersion
		resp.CreatedAt = embeddingStatus.CreatedAt.Unix()
		resp.UpdatedAt = embeddingStatus.UpdatedAt.Unix()
		resp.EmbeddingValues = embeddingStatus.EmbeddingValues
	}

	return resp, nil
}

// generateSnippet はコンテンツから最大maxLen文字のスニペットを生成する
func generateSnippet(content string, maxLen int) string {
	runes := []rune(content)
	if len(runes) <= maxLen {
		return content
	}
	return string(runes[:maxLen]) + "..."
}
