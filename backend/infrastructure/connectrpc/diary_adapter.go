package connect

import (
	"context"

	"connectrpc.com/connect"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/grpc/grpcconnect"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
)

// DiaryServiceAdapter は diary.DiaryEntry を grpcconnect.DiaryServiceHandler としてラップするアダプター。
type DiaryServiceAdapter struct {
	svc *diary.DiaryEntry
}

// NewDiaryServiceAdapter は DiaryServiceAdapter を生成する。
func NewDiaryServiceAdapter(svc *diary.DiaryEntry) grpcconnect.DiaryServiceHandler {
	return &DiaryServiceAdapter{svc: svc}
}

func (a *DiaryServiceAdapter) CreateDiaryEntry(ctx context.Context, req *connect.Request[g.CreateDiaryEntryRequest]) (*connect.Response[g.CreateDiaryEntryResponse], error) {
	resp, err := a.svc.CreateDiaryEntry(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) UpdateDiaryEntry(ctx context.Context, req *connect.Request[g.UpdateDiaryEntryRequest]) (*connect.Response[g.UpdateDiaryEntryResponse], error) {
	resp, err := a.svc.UpdateDiaryEntry(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) DeleteDiaryEntry(ctx context.Context, req *connect.Request[g.DeleteDiaryEntryRequest]) (*connect.Response[g.DeleteDiaryEntryResponse], error) {
	resp, err := a.svc.DeleteDiaryEntry(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) GetDiaryEntry(ctx context.Context, req *connect.Request[g.GetDiaryEntryRequest]) (*connect.Response[g.GetDiaryEntryResponse], error) {
	resp, err := a.svc.GetDiaryEntry(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) GetDiaryEntries(ctx context.Context, req *connect.Request[g.GetDiaryEntriesRequest]) (*connect.Response[g.GetDiaryEntriesResponse], error) {
	resp, err := a.svc.GetDiaryEntries(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) GetDiaryEntriesByMonth(ctx context.Context, req *connect.Request[g.GetDiaryEntriesByMonthRequest]) (*connect.Response[g.GetDiaryEntriesByMonthResponse], error) {
	resp, err := a.svc.GetDiaryEntriesByMonth(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) SearchDiaryEntries(ctx context.Context, req *connect.Request[g.SearchDiaryEntriesRequest]) (*connect.Response[g.SearchDiaryEntriesResponse], error) {
	resp, err := a.svc.SearchDiaryEntries(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) GenerateMonthlySummary(ctx context.Context, req *connect.Request[g.GenerateMonthlySummaryRequest]) (*connect.Response[g.GenerateMonthlySummaryResponse], error) {
	resp, err := a.svc.GenerateMonthlySummary(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) GetMonthlySummary(ctx context.Context, req *connect.Request[g.GetMonthlySummaryRequest]) (*connect.Response[g.GetMonthlySummaryResponse], error) {
	resp, err := a.svc.GetMonthlySummary(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) GetLatestTrend(ctx context.Context, req *connect.Request[g.GetLatestTrendRequest]) (*connect.Response[g.GetLatestTrendResponse], error) {
	resp, err := a.svc.GetLatestTrend(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) TriggerLatestTrend(ctx context.Context, req *connect.Request[g.TriggerLatestTrendRequest]) (*connect.Response[g.TriggerLatestTrendResponse], error) {
	resp, err := a.svc.TriggerLatestTrend(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) SearchDiaryEntriesSemantic(ctx context.Context, req *connect.Request[g.SearchDiaryEntriesSemanticRequest]) (*connect.Response[g.SearchDiaryEntriesSemanticResponse], error) {
	resp, err := a.svc.SearchDiaryEntriesSemantic(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) TriggerDiaryHighlight(ctx context.Context, req *connect.Request[g.TriggerDiaryHighlightRequest]) (*connect.Response[g.TriggerDiaryHighlightResponse], error) {
	resp, err := a.svc.TriggerDiaryHighlight(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) GetDiaryHighlight(ctx context.Context, req *connect.Request[g.GetDiaryHighlightRequest]) (*connect.Response[g.GetDiaryHighlightResponse], error) {
	resp, err := a.svc.GetDiaryHighlight(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) RegenerateAllEmbeddings(ctx context.Context, req *connect.Request[g.RegenerateAllEmbeddingsRequest]) (*connect.Response[g.RegenerateAllEmbeddingsResponse], error) {
	resp, err := a.svc.RegenerateAllEmbeddings(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *DiaryServiceAdapter) GetDiaryEmbeddingStatus(ctx context.Context, req *connect.Request[g.GetDiaryEmbeddingStatusRequest]) (*connect.Response[g.GetDiaryEmbeddingStatusResponse], error) {
	resp, err := a.svc.GetDiaryEmbeddingStatus(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}
