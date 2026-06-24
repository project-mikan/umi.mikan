package connect

import (
	"context"

	"connectrpc.com/connect"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/grpc/grpcconnect"
	"github.com/project-mikan/umi.mikan/backend/service/auth"
)

// AuthServiceAdapter は auth.AuthEntry を grpcconnect.AuthServiceHandler としてラップするアダプター。
// ConnectRPC の connect.Request/Response ラッパーを剥がして既存の gRPC サービスへ委譲する。
type AuthServiceAdapter struct {
	svc *auth.AuthEntry
}

// NewAuthServiceAdapter は AuthServiceAdapter を生成する。
func NewAuthServiceAdapter(svc *auth.AuthEntry) grpcconnect.AuthServiceHandler {
	return &AuthServiceAdapter{svc: svc}
}

func (a *AuthServiceAdapter) GetRegistrationConfig(ctx context.Context, req *connect.Request[g.GetRegistrationConfigRequest]) (*connect.Response[g.GetRegistrationConfigResponse], error) {
	resp, err := a.svc.GetRegistrationConfig(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *AuthServiceAdapter) RegisterByPassword(ctx context.Context, req *connect.Request[g.RegisterByPasswordRequest]) (*connect.Response[g.AuthResponse], error) {
	resp, err := a.svc.RegisterByPassword(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *AuthServiceAdapter) LoginByPassword(ctx context.Context, req *connect.Request[g.LoginByPasswordRequest]) (*connect.Response[g.AuthResponse], error) {
	resp, err := a.svc.LoginByPassword(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *AuthServiceAdapter) RefreshAccessToken(ctx context.Context, req *connect.Request[g.RefreshAccessTokenRequest]) (*connect.Response[g.AuthResponse], error) {
	resp, err := a.svc.RefreshAccessToken(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}
