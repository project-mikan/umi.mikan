package diary

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RedisInterface defines the methods needed for caching diary counts
type RedisInterface interface {
	SetDiaryCount(ctx context.Context, userID string, count uint32) error
	GetDiaryCount(ctx context.Context, userID string) (uint32, error)
	UpdateDiaryCount(ctx context.Context, userID string, delta int) error
	DeleteDiaryCount(ctx context.Context, userID string) error
	Close() error
}

type DiaryEntry struct {
	g.UnimplementedDiaryServiceServer
	DB    database.DB
	Redis RedisInterface
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

	// 日記が作成されたのでキャッシュの値を非同期で+1する
	go func(userIDStr string) {
		// 新しいcontextを使用（元のcontextがキャンセルされても実行継続）
		bgCtx := context.Background()
		if err := s.safeUpdateDiaryCount(bgCtx, userIDStr, 1); err != nil {
			log.Printf("Warning: failed to update diary count cache after creation for user %s: %v", userIDStr, err)
		}
	}(userID.String())

	return &g.CreateDiaryEntryResponse{
		Entry: &g.DiaryEntry{
			Id:      diary.ID.String(),
			Date:    message.Date,
			Content: diary.Content,
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
			Id:      diary.ID.String(),
			Date:    message.Date,
			Content: diary.Content,
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
			Id:      diary.ID.String(),
			Date:    dateMsg,
			Content: diary.Content,
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
			Id:      diary.ID.String(),
			Date:    &g.YMD{Year: uint32(diary.Date.Year()), Month: uint32(diary.Date.Month()), Day: uint32(diary.Date.Day())},
			Content: diary.Content,
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

	return &g.UpdateDiaryEntryResponse{
		Entry: &g.DiaryEntry{
			Id:      diary.ID.String(),
			Date:    &g.YMD{Year: uint32(diary.Date.Year()), Month: uint32(diary.Date.Month()), Day: uint32(diary.Date.Day())},
			Content: diary.Content,
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

	// 日記が削除されたのでキャッシュの値を非同期で-1する
	go func(userIDStr string) {
		// 新しいcontextを使用（元のcontextがキャンセルされても実行継続）
		bgCtx := context.Background()
		if err := s.safeUpdateDiaryCount(bgCtx, userIDStr, -1); err != nil {
			log.Printf("Warning: failed to update diary count cache after deletion for user %s: %v", userIDStr, err)
		}
	}(userID.String())

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
			Id:      d.ID.String(),
			Content: d.Content,
			Date:    &g.YMD{Year: uint32(d.Date.Year()), Month: uint32(d.Date.Month()), Day: uint32(d.Date.Day())},
		})
	}
	return &g.SearchDiaryEntriesResponse{
		SearchedKeyword: message.Keyword,
		Entries:         entries,
	}, nil
}

func (s *DiaryEntry) GetDiaryCount(
	ctx context.Context,
	message *g.GetDiaryCountRequest,
) (*g.GetDiaryCountResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// まずキャッシュから取得を試行
	if count, err := s.Redis.GetDiaryCount(ctx, userID.String()); err == nil {
		return &g.GetDiaryCountResponse{Count: count}, nil
	}

	// キャッシュにない場合はDBから取得
	count, err := database.CountDiariesByUserID(ctx, s.DB, userID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get diary count: %w", err)
	}

	// 結果をキャッシュに保存
	if err := s.Redis.SetDiaryCount(ctx, userID.String(), uint32(count)); err != nil {
		log.Printf("Warning: failed to cache diary count for user %s: %v", userID.String(), err)
	}

	return &g.GetDiaryCountResponse{Count: uint32(count)}, nil
}

// safeUpdateDiaryCount safely updates the diary count cache, initializing from DB if cache doesn't exist
func (s *DiaryEntry) safeUpdateDiaryCount(ctx context.Context, userID string, delta int) error {
	// Check if cache key exists first
	_, err := s.Redis.GetDiaryCount(ctx, userID)
	if err != nil {
		// Cache miss - initialize from database with the current count (after DB change)
		count, err := database.CountDiariesByUserID(ctx, s.DB, userID)
		if err != nil {
			return fmt.Errorf("failed to get diary count from DB for cache initialization: %w", err)
		}

		// Set the cache to current DB value (which already reflects the change)
		// No need to add delta since DB already has the updated count
		if err := s.Redis.SetDiaryCount(ctx, userID, uint32(count)); err != nil {
			return fmt.Errorf("failed to initialize diary count cache: %w", err)
		}
		return nil // Cache is already set to correct value
	}

	// Cache exists - just update it
	return s.Redis.UpdateDiaryCount(ctx, userID, delta)
}
