package diary

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/redis/rueidis"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DiaryEntry struct {
	g.UnimplementedDiaryServiceServer
	DB    database.DB
	Redis rueidis.Client
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

	// トランザクション内でdiaryとdiary_entitiesを保存
	err = database.RwTransaction(ctx, s.DB.(*sql.DB), func(tx *sql.Tx) error {
		if err := diary.Insert(ctx, tx); err != nil {
			return err
		}

		// diary_entitiesを保存
		if err := s.saveDiaryEntities(ctx, tx, diary.ID, message.DiaryEntities, currentTime); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

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

// diary_entitiesを取得してDiaryEntityOutputに変換
func (s *DiaryEntry) getDiaryEntityOutputs(ctx context.Context, diaryID uuid.UUID) ([]*g.DiaryEntityOutput, error) {
	diaryEntities, err := database.DiaryEntitiesByDiaryID(ctx, s.DB, diaryID)
	if err != nil {
		return nil, err
	}

	diaryEntityOutputs := make([]*g.DiaryEntityOutput, 0, len(diaryEntities))
	for _, de := range diaryEntities {
		// positionsをJSONからデコード（alias_idを含む）
		var positionsRaw []map[string]interface{}
		if err := json.Unmarshal(de.Positions, &positionsRaw); err != nil {
			return nil, err
		}

		// map[string]interface{}から*g.Positionに変換
		positions := make([]*g.Position, 0, len(positionsRaw))
		for _, posRaw := range positionsRaw {
			pos := &g.Position{
				Start: uint32(posRaw["start"].(float64)),
				End:   uint32(posRaw["end"].(float64)),
			}
			// alias_idがあれば設定
			if aliasID, ok := posRaw["alias_id"].(string); ok && aliasID != "" {
				pos.AliasId = aliasID
			}
			positions = append(positions, pos)
		}

		diaryEntityOutputs = append(diaryEntityOutputs, &g.DiaryEntityOutput{
			EntityId:  de.EntityID.String(),
			Positions: positions,
		})
	}

	return diaryEntityOutputs, nil
}

// getDiaryEntityOutputsForDiaries 複数の日記に対してdiary_entitiesを一括取得（N+1問題を回避）
func (s *DiaryEntry) getDiaryEntityOutputsForDiaries(ctx context.Context, diaryIDs []uuid.UUID) (map[string][]*g.DiaryEntityOutput, error) {
	if len(diaryIDs) == 0 {
		return make(map[string][]*g.DiaryEntityOutput), nil
	}

	// diary_entitiesを一括取得
	query := `
		SELECT id, diary_id, entity_id, created_at, updated_at, positions
		FROM diary_entities
		WHERE diary_id = ANY($1)
		ORDER BY diary_id, created_at
	`
	rows, err := s.DB.(*sql.DB).QueryContext(ctx, query, pq.Array(diaryIDs))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	// diary_idごとにグループ化
	entityMap := make(map[string][]*g.DiaryEntityOutput)
	for rows.Next() {
		var de database.DiaryEntity
		if err := rows.Scan(&de.ID, &de.DiaryID, &de.EntityID, &de.CreatedAt, &de.UpdatedAt, &de.Positions); err != nil {
			return nil, err
		}

		// positionsをJSONからデコード
		var positionsRaw []map[string]interface{}
		if err := json.Unmarshal(de.Positions, &positionsRaw); err != nil {
			return nil, err
		}

		// map[string]interface{}から*g.Positionに変換
		positions := make([]*g.Position, 0, len(positionsRaw))
		for _, posRaw := range positionsRaw {
			pos := &g.Position{
				Start: uint32(posRaw["start"].(float64)),
				End:   uint32(posRaw["end"].(float64)),
			}
			// alias_idがあれば設定
			if aliasID, ok := posRaw["alias_id"].(string); ok && aliasID != "" {
				pos.AliasId = aliasID
			}
			positions = append(positions, pos)
		}

		diaryIDStr := de.DiaryID.String()
		entityMap[diaryIDStr] = append(entityMap[diaryIDStr], &g.DiaryEntityOutput{
			EntityId:  de.EntityID.String(),
			Positions: positions,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return entityMap, nil
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

	// diary_entitiesを取得
	diaryEntityOutputs, err := s.getDiaryEntityOutputs(ctx, diary.ID)
	if err != nil {
		return nil, err
	}

	return &g.GetDiaryEntryResponse{
		Entry: &g.DiaryEntry{
			Id:            diary.ID.String(),
			Date:          message.Date,
			Content:       diary.Content,
			CreatedAt:     diary.CreatedAt,
			UpdatedAt:     diary.UpdatedAt,
			DiaryEntities: diaryEntityOutputs,
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

	// diary_entitiesを一括取得（N+1問題を回避）
	diaryIDs := make([]uuid.UUID, 0, len(diariesWithDates))
	for _, dwd := range diariesWithDates {
		diaryIDs = append(diaryIDs, dwd.diary.ID)
	}
	entityMap, err := s.getDiaryEntityOutputsForDiaries(ctx, diaryIDs)
	if err != nil {
		return nil, err
	}

	entries := make([]*g.DiaryEntry, 0, len(diariesWithDates))
	for _, dwd := range diariesWithDates {
		diaryEntityOutputs := entityMap[dwd.diary.ID.String()]
		if diaryEntityOutputs == nil {
			diaryEntityOutputs = []*g.DiaryEntityOutput{}
		}

		entries = append(entries, &g.DiaryEntry{
			Id:            dwd.diary.ID.String(),
			Date:          dwd.dateMsg,
			Content:       dwd.diary.Content,
			CreatedAt:     dwd.diary.CreatedAt,
			UpdatedAt:     dwd.diary.UpdatedAt,
			DiaryEntities: diaryEntityOutputs,
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

	// diary_entitiesを一括取得（N+1問題を回避）
	diaryIDs := make([]uuid.UUID, 0, len(diaries))
	for _, d := range diaries {
		diaryIDs = append(diaryIDs, d.ID)
	}
	entityMap, err := s.getDiaryEntityOutputsForDiaries(ctx, diaryIDs)
	if err != nil {
		return nil, err
	}

	entries := make([]*g.DiaryEntry, 0, len(diaries))
	for _, diary := range diaries {
		diaryEntityOutputs := entityMap[diary.ID.String()]
		if diaryEntityOutputs == nil {
			diaryEntityOutputs = []*g.DiaryEntityOutput{}
		}

		entries = append(entries, &g.DiaryEntry{
			Id:            diary.ID.String(),
			Date:          &g.YMD{Year: uint32(diary.Date.Year()), Month: uint32(diary.Date.Month()), Day: uint32(diary.Date.Day())},
			Content:       diary.Content,
			CreatedAt:     diary.CreatedAt,
			UpdatedAt:     diary.UpdatedAt,
			DiaryEntities: diaryEntityOutputs,
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

		// 既存のdiary_entitiesを削除してから新しいものを保存
		if err := s.deleteDiaryEntities(ctx, tx, diary.ID); err != nil {
			return err
		}

		// 新しいdiary_entitiesを保存
		if err := s.saveDiaryEntities(ctx, tx, diary.ID, message.DiaryEntities, currentTime); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

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

	// diary_entitiesを一括取得（N+1問題を回避）
	diaryIDs := make([]uuid.UUID, 0, len(ds))
	for _, d := range ds {
		diaryIDs = append(diaryIDs, d.ID)
	}
	entityMap, err := s.getDiaryEntityOutputsForDiaries(ctx, diaryIDs)
	if err != nil {
		return nil, err
	}

	entries := make([]*g.DiaryEntry, 0, len(ds))
	for _, d := range ds {
		diaryEntityOutputs := entityMap[d.ID.String()]
		if diaryEntityOutputs == nil {
			diaryEntityOutputs = []*g.DiaryEntityOutput{}
		}

		entries = append(entries, &g.DiaryEntry{
			Id:            d.ID.String(),
			Content:       d.Content,
			Date:          &g.YMD{Year: uint32(d.Date.Year()), Month: uint32(d.Date.Month()), Day: uint32(d.Date.Day())},
			CreatedAt:     d.CreatedAt,
			UpdatedAt:     d.UpdatedAt,
			DiaryEntities: diaryEntityOutputs,
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

	// タスクを「キューに追加済み」としてマーク（10分の有効期限）
	if err := s.setTaskStatus(ctx, taskKey, "queued", 600); err != nil {
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

	// タスクを「キューに追加済み」としてマーク（10分の有効期限）
	if err := s.setTaskStatus(ctx, taskKey, "queued", 600); err != nil {
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

// saveDiaryEntities diary_entitiesを保存
func (s *DiaryEntry) saveDiaryEntities(ctx context.Context, tx *sql.Tx, diaryID uuid.UUID, entities []*g.DiaryEntityInput, currentTime int64) error {
	if len(entities) == 0 {
		return nil
	}

	for _, entity := range entities {
		entityID, err := uuid.Parse(entity.EntityId)
		if err != nil {
			return status.Errorf(codes.InvalidArgument, "invalid entity ID: %s", entity.EntityId)
		}

		// positionsをJSONBに変換（alias_idも含む）
		positions := make([]map[string]interface{}, 0, len(entity.Positions))
		for _, pos := range entity.Positions {
			posMap := map[string]interface{}{
				"start": pos.Start,
				"end":   pos.End,
			}
			// alias_idが空文字列でない場合のみ含める
			if pos.AliasId != "" {
				posMap["alias_id"] = pos.AliasId
			}
			positions = append(positions, posMap)
		}
		positionsJSON, err := json.Marshal(positions)
		if err != nil {
			return status.Errorf(codes.Internal, "failed to marshal positions")
		}

		// diary_entityを作成
		diaryEntity := &database.DiaryEntity{
			ID:        uuid.New(),
			DiaryID:   diaryID,
			EntityID:  entityID,
			Positions: positionsJSON,
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
		}

		if err := diaryEntity.Insert(ctx, tx); err != nil {
			return err
		}
	}

	return nil
}

// deleteDiaryEntities 特定の日記に紐づくdiary_entitiesを削除
func (s *DiaryEntry) deleteDiaryEntities(ctx context.Context, tx *sql.Tx, diaryID uuid.UUID) error {
	query := "DELETE FROM diary_entities WHERE diary_id = $1"
	_, err := tx.ExecContext(ctx, query, diaryID)
	return err
}
