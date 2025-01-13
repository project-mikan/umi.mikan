package diary

import (
	"context"

	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
)

type DiaryEntry struct {
	g.UnimplementedDiaryServiceServer
	DB database.DB
}

func (s *DiaryEntry) CreateDiaryEntry(
	ctx context.Context,
	message *g.CreateDiaryEntryRequest,
) (*g.CreateDiaryEntryResponse, error) {
	return &g.CreateDiaryEntryResponse{}, nil
}

func (s *DiaryEntry) GetDiaryEntry(
	ctx context.Context,
	message *g.GetDiaryEntryRequest,
) (*g.GetDiaryEntryResponse, error) {
	return &g.GetDiaryEntryResponse{}, nil
}

func (s *DiaryEntry) ListDiaryEntries(
	ctx context.Context,
	message *g.ListDiaryEntriesRequest,
) (*g.ListDiaryEntriesResponse, error) {
	return &g.ListDiaryEntriesResponse{}, nil
}

func (s *DiaryEntry) UpdateDiaryEntry(
	ctx context.Context,
	message *g.UpdateDiaryEntryRequest,
) (*g.UpdateDiaryEntryResponse, error) {
	return &g.UpdateDiaryEntryResponse{}, nil
}

func (s *DiaryEntry) DeleteDiaryEntry(
	ctx context.Context,
	message *g.DeleteDiaryEntryRequest,
) (*g.DeleteDiaryEntryResponse, error) {
	return &g.DeleteDiaryEntryResponse{}, nil
}

func (s *DiaryEntry) SearchDiaryEntries(
	ctx context.Context,
	message *g.SearchDiaryEntriesRequest,
) (*g.SearchDiaryEntriesResponse, error) {
	ds, err := database.DiariesByUserIDAndContent(ctx, s.DB, message.UserID, message.Keyword)
	if err != nil {
		return nil, err
	}
	entries := make([]*g.DiaryEntry, 0, len(ds))
	for _, d := range ds {
		entries = append(entries, &g.DiaryEntry{
			Id:      d.ID.String(),
			Content: d.Content,
			Date:    &g.Date{Year: uint32(d.Date.Year()), Month: uint32(d.Date.Month()), Day: uint32(d.Date.Day())},
		})
	}
	return &g.SearchDiaryEntriesResponse{
		Entries: entries,
	}, nil
}
