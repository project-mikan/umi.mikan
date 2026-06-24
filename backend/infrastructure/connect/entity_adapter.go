package connect

import (
	"context"

	"connectrpc.com/connect"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/grpc/grpcconnect"
	"github.com/project-mikan/umi.mikan/backend/service/entity"
)

// EntityServiceAdapter は entity.EntityEntry を grpcconnect.EntityServiceHandler としてラップするアダプター。
type EntityServiceAdapter struct {
	svc *entity.EntityEntry
}

// NewEntityServiceAdapter は EntityServiceAdapter を生成する。
func NewEntityServiceAdapter(svc *entity.EntityEntry) grpcconnect.EntityServiceHandler {
	return &EntityServiceAdapter{svc: svc}
}

func (a *EntityServiceAdapter) CreateEntity(ctx context.Context, req *connect.Request[g.CreateEntityRequest]) (*connect.Response[g.CreateEntityResponse], error) {
	resp, err := a.svc.CreateEntity(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *EntityServiceAdapter) UpdateEntity(ctx context.Context, req *connect.Request[g.UpdateEntityRequest]) (*connect.Response[g.UpdateEntityResponse], error) {
	resp, err := a.svc.UpdateEntity(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *EntityServiceAdapter) DeleteEntity(ctx context.Context, req *connect.Request[g.DeleteEntityRequest]) (*connect.Response[g.DeleteEntityResponse], error) {
	resp, err := a.svc.DeleteEntity(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *EntityServiceAdapter) GetEntity(ctx context.Context, req *connect.Request[g.GetEntityRequest]) (*connect.Response[g.GetEntityResponse], error) {
	resp, err := a.svc.GetEntity(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *EntityServiceAdapter) ListEntities(ctx context.Context, req *connect.Request[g.ListEntitiesRequest]) (*connect.Response[g.ListEntitiesResponse], error) {
	resp, err := a.svc.ListEntities(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *EntityServiceAdapter) CreateEntityAlias(ctx context.Context, req *connect.Request[g.CreateEntityAliasRequest]) (*connect.Response[g.CreateEntityAliasResponse], error) {
	resp, err := a.svc.CreateEntityAlias(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *EntityServiceAdapter) UpdateEntityAlias(ctx context.Context, req *connect.Request[g.UpdateEntityAliasRequest]) (*connect.Response[g.UpdateEntityAliasResponse], error) {
	resp, err := a.svc.UpdateEntityAlias(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *EntityServiceAdapter) DeleteEntityAlias(ctx context.Context, req *connect.Request[g.DeleteEntityAliasRequest]) (*connect.Response[g.DeleteEntityAliasResponse], error) {
	resp, err := a.svc.DeleteEntityAlias(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *EntityServiceAdapter) SearchEntities(ctx context.Context, req *connect.Request[g.SearchEntitiesRequest]) (*connect.Response[g.SearchEntitiesResponse], error) {
	resp, err := a.svc.SearchEntities(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}
