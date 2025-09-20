package diary

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/llm"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DiaryEntry struct {
	g.UnimplementedDiaryServiceServer
	DB database.DB
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

	if err := diary.Insert(ctx, s.DB); err != nil {
		return nil, err
	}

	// 自動要約生成のチェック（非同期で実行）
	go s.tryGenerateAutoSummary(context.Background(), userID, date, message.Content)

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

	entries := make([]*g.DiaryEntry, 0)

	for _, dateMsg := range message.Dates {
		date := time.Date(int(dateMsg.Year), time.Month(dateMsg.Month), int(dateMsg.Day), 0, 0, 0, 0, time.UTC)
		diary, err := database.DiaryByUserIDDate(ctx, s.DB, userID, date)
		if err != nil {
			continue // Skip entries that don't exist
		}

		entries = append(entries, &g.DiaryEntry{
			Id:        diary.ID.String(),
			Date:      dateMsg,
			Content:   diary.Content,
			CreatedAt: diary.CreatedAt,
			UpdatedAt: diary.UpdatedAt,
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
	entries := make([]*g.DiaryEntry, 0)

	// Get the number of days in the month
	daysInMonth := time.Date(int(message.Month.Year), time.Month(message.Month.Month)+1, 0, 0, 0, 0, 0, time.UTC).Day()

	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(int(message.Month.Year), time.Month(message.Month.Month), day, 0, 0, 0, 0, time.UTC)
		diary, err := database.DiaryByUserIDDate(ctx, s.DB, userID, date)
		if err != nil {
			continue // Skip entries that don't exist
		}

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
		diary.UpdatedAt = time.Now().Unix()

		if err := diary.Update(ctx, tx); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 自動要約生成のチェック（非同期で実行）
	go s.tryGenerateAutoSummary(context.Background(), userID, diary.Date, message.Content)

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

	ds, err := database.DiariesByUserIDAndContent(ctx, s.DB, userID.String(), message.Keyword)
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
		SearchedKeyword: message.Keyword,
		Entries:         entries,
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

	// Get user's Gemini API key (LLM provider 1 = Gemini)
	userLLM, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, userID, 1)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Gemini API key not found for user")
	}

	// Get all diary entries for the specified month
	entries := make([]*database.Diary, 0)
	daysInMonth := time.Date(int(message.Month.Year), time.Month(message.Month.Month)+1, 0, 0, 0, 0, 0, time.UTC).Day()

	for day := 1; day <= daysInMonth; day++ {
		date := time.Date(int(message.Month.Year), time.Month(message.Month.Month), day, 0, 0, 0, 0, time.UTC)
		diary, err := database.DiaryByUserIDDate(ctx, s.DB, userID, date)
		if err != nil {
			continue // Skip entries that don't exist
		}
		entries = append(entries, diary)
	}

	if len(entries) == 0 {
		return nil, status.Errorf(codes.NotFound, "no diary entries found for the specified month")
	}

	// Combine all diary contents
	var contentBuilder strings.Builder
	for _, entry := range entries {
		contentBuilder.WriteString(fmt.Sprintf("【%d日】\n%s\n\n", entry.Date.Day(), entry.Content))
	}
	combinedContent := contentBuilder.String()

	// Generate summary using Gemini
	geminiClient, err := llm.NewGeminiClient(ctx, userLLM.Key)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create Gemini client: %v", err)
	}
	defer func() {
		if closeErr := geminiClient.Close(); closeErr != nil {
			// Log the close error but don't affect the main function result
			fmt.Printf("Warning: failed to close Gemini client: %v\n", closeErr)
		}
	}()

	summary, err := geminiClient.GenerateSummary(ctx, combinedContent)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate summary: %v", err)
	}

	// Save or update the summary in database
	currentTime := time.Now().Unix()

	// Check if summary already exists for this month
	existingSummary, err := database.DiarySummaryMonthByUserIDYearMonth(ctx, s.DB, userID, int(message.Month.Year), int(message.Month.Month))
	if err != nil {
		// Create new summary
		summaryID := uuid.New()
		newSummary := &database.DiarySummaryMonth{
			ID:        summaryID,
			UserID:    userID,
			Year:      int(message.Month.Year),
			Month:     int(message.Month.Month),
			Summary:   summary,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
		}

		if err := newSummary.Insert(ctx, s.DB); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to save summary: %v", err)
		}

		return &g.GenerateMonthlySummaryResponse{
			Summary: &g.MonthlySummary{
				Id:        newSummary.ID.String(),
				Month:     message.Month,
				Summary:   newSummary.Summary,
				CreatedAt: newSummary.CreatedAt,
				UpdatedAt: newSummary.UpdatedAt,
			},
		}, nil
	} else {
		// Update existing summary
		existingSummary.Summary = summary
		existingSummary.UpdatedAt = currentTime

		if err := existingSummary.Update(ctx, s.DB); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to update summary: %v", err)
		}

		return &g.GenerateMonthlySummaryResponse{
			Summary: &g.MonthlySummary{
				Id:        existingSummary.ID.String(),
				Month:     message.Month,
				Summary:   existingSummary.Summary,
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

	// Get the summary for the specified month
	summary, err := database.DiarySummaryMonthByUserIDYearMonth(ctx, s.DB, userID, int(message.Month.Year), int(message.Month.Month))
	if err != nil {
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

// tryGenerateAutoSummary は1000文字以上の日記に対して自動要約を生成する
func (s *DiaryEntry) tryGenerateAutoSummary(ctx context.Context, userID uuid.UUID, date time.Time, content string) {
	// 1000文字未満の場合は要約を生成しない
	if len([]rune(content)) < 1000 {
		return
	}

	// ユーザーの自動要約設定を取得（Gemini API使用）
	userLLM, err := database.UserLlmByUserIDLlmProvider(ctx, s.DB, userID, 1)
	if err != nil {
		// LLMキーが設定されていない場合は何もしない
		return
	}

	// 日毎の自動要約が無効の場合は何もしない
	if !userLLM.AutoSummaryDaily {
		return
	}

	// 既に要約が存在するかチェック
	_, err = database.DiarySummaryDayByUserIDDate(ctx, s.DB, userID, date)
	if err == nil {
		// 既に要約が存在する場合は再生成
		s.regenerateDailySummary(ctx, userID, date, content, userLLM.Key)
		return
	}

	// 新規要約を生成
	s.generateDailySummary(ctx, userID, date, content, userLLM.Key)
}

// generateDailySummary は新規の日記要約を生成する
func (s *DiaryEntry) generateDailySummary(ctx context.Context, userID uuid.UUID, date time.Time, content string, apiKey string) {
	// Geminiクライアントを作成
	geminiClient, err := llm.NewGeminiClient(ctx, apiKey)
	if err != nil {
		return // エラーの場合は何もしない（ログ出力は省略）
	}
	defer func() {
		_ = geminiClient.Close()
	}()

	// 要約を生成
	summary, err := geminiClient.GenerateSummary(ctx, content)
	if err != nil {
		return // エラーの場合は何もしない
	}

	// データベースに保存
	currentTime := time.Now().Unix()
	summaryID := uuid.New()

	newSummary := &database.DiarySummaryDay{
		ID:        summaryID,
		UserID:    userID,
		Date:      date,
		Summary:   summary,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}

	// 保存に失敗してもエラーを無視（非同期処理のため）
	_ = newSummary.Insert(ctx, s.DB)
}

// regenerateDailySummary は既存の日記要約を更新する
func (s *DiaryEntry) regenerateDailySummary(ctx context.Context, userID uuid.UUID, date time.Time, content string, apiKey string) {
	// Geminiクライアントを作成
	geminiClient, err := llm.NewGeminiClient(ctx, apiKey)
	if err != nil {
		return // エラーの場合は何もしない
	}
	defer func() {
		_ = geminiClient.Close()
	}()

	// 要約を生成
	summary, err := geminiClient.GenerateSummary(ctx, content)
	if err != nil {
		return // エラーの場合は何もしない
	}

	// 既存の要約を取得して更新
	existingSummary, err := database.DiarySummaryDayByUserIDDate(ctx, s.DB, userID, date)
	if err != nil {
		return // 取得に失敗した場合は何もしない
	}

	existingSummary.Summary = summary
	existingSummary.UpdatedAt = time.Now().Unix()

	// 更新に失敗してもエラーを無視（非同期処理のため）
	_ = existingSummary.Update(ctx, s.DB)
}
