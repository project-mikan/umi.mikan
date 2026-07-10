package connect

import (
	"context"

	"connectrpc.com/connect"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/grpc/grpcconnect"
	"github.com/project-mikan/umi.mikan/backend/service/user"
)

// UserServiceAdapter は user.UserEntry を grpcconnect.UserServiceHandler としてラップするアダプター。
type UserServiceAdapter struct {
	svc *user.UserEntry
}

// NewUserServiceAdapter は UserServiceAdapter を生成する。
func NewUserServiceAdapter(svc *user.UserEntry) grpcconnect.UserServiceHandler {
	return &UserServiceAdapter{svc: svc}
}

func (a *UserServiceAdapter) UpdateUserName(ctx context.Context, req *connect.Request[g.UpdateUserNameRequest]) (*connect.Response[g.UpdateUserNameResponse], error) {
	resp, err := a.svc.UpdateUserName(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *UserServiceAdapter) ChangePassword(ctx context.Context, req *connect.Request[g.ChangePasswordRequest]) (*connect.Response[g.ChangePasswordResponse], error) {
	resp, err := a.svc.ChangePassword(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *UserServiceAdapter) UpdateLLMKey(ctx context.Context, req *connect.Request[g.UpdateLLMKeyRequest]) (*connect.Response[g.UpdateLLMKeyResponse], error) {
	resp, err := a.svc.UpdateLLMKey(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *UserServiceAdapter) GetUserInfo(ctx context.Context, req *connect.Request[g.GetUserInfoRequest]) (*connect.Response[g.GetUserInfoResponse], error) {
	resp, err := a.svc.GetUserInfo(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *UserServiceAdapter) DeleteLLMKey(ctx context.Context, req *connect.Request[g.DeleteLLMKeyRequest]) (*connect.Response[g.DeleteLLMKeyResponse], error) {
	resp, err := a.svc.DeleteLLMKey(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *UserServiceAdapter) DeleteAccount(ctx context.Context, req *connect.Request[g.DeleteAccountRequest]) (*connect.Response[g.DeleteAccountResponse], error) {
	resp, err := a.svc.DeleteAccount(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *UserServiceAdapter) UpdateAutoSummarySettings(ctx context.Context, req *connect.Request[g.UpdateAutoSummarySettingsRequest]) (*connect.Response[g.UpdateAutoSummarySettingsResponse], error) {
	resp, err := a.svc.UpdateAutoSummarySettings(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *UserServiceAdapter) GetAutoSummarySettings(ctx context.Context, req *connect.Request[g.GetAutoSummarySettingsRequest]) (*connect.Response[g.GetAutoSummarySettingsResponse], error) {
	resp, err := a.svc.GetAutoSummarySettings(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *UserServiceAdapter) GetPubSubMetrics(ctx context.Context, req *connect.Request[g.GetPubSubMetricsRequest]) (*connect.Response[g.GetPubSubMetricsResponse], error) {
	resp, err := a.svc.GetPubSubMetrics(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *UserServiceAdapter) CreateApiKey(ctx context.Context, req *connect.Request[g.CreateApiKeyRequest]) (*connect.Response[g.CreateApiKeyResponse], error) {
	resp, err := a.svc.CreateApiKey(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *UserServiceAdapter) ListApiKeys(ctx context.Context, req *connect.Request[g.ListApiKeysRequest]) (*connect.Response[g.ListApiKeysResponse], error) {
	resp, err := a.svc.ListApiKeys(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}

func (a *UserServiceAdapter) DeleteApiKey(ctx context.Context, req *connect.Request[g.DeleteApiKeyRequest]) (*connect.Response[g.DeleteApiKeyResponse], error) {
	resp, err := a.svc.DeleteApiKey(ctx, req.Msg)
	if err != nil {
		return nil, grpcStatusToConnectError(err)
	}
	return connect.NewResponse(resp), nil
}
