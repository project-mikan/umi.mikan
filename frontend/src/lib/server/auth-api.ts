import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createGrpcTransport } from "@connectrpc/connect-node";
import {
  type AuthResponse,
  type GetRegistrationConfigResponse,
  AuthService,
  GetRegistrationConfigRequestSchema,
  LoginByPasswordRequestSchema,
  RefreshAccessTokenRequestSchema,
  RegisterByPasswordRequestSchema,
} from "$lib/grpc/auth/auth_pb";
import {
  UserService,
  UpdateUserNameRequestSchema,
  ChangePasswordRequestSchema,
  UpdateLLMKeyRequestSchema,
  GetUserInfoRequestSchema,
  DeleteLLMKeyRequestSchema,
  DeleteAccountRequestSchema,
  UpdateAutoSummarySettingsRequestSchema,
  type UpdateUserNameResponse,
  type ChangePasswordResponse,
  type UpdateLLMKeyResponse,
  type GetUserInfoResponse,
  type DeleteLLMKeyResponse,
  type DeleteAccountResponse,
  type UpdateAutoSummarySettingsResponse,
} from "$lib/grpc/user/user_pb";

// モジュールレベルで Transport と Client を共有する（リクエストごとの生成コストを排除）
const transport = createGrpcTransport({
  baseUrl: "http://backend:8080",
});

const authClient = createClient(AuthService, transport);
const userClient = createClient(UserService, transport);

function authHeader(accessToken: string) {
  return { headers: { authorization: `Bearer ${accessToken}` } };
}

export async function getRegistrationConfig(): Promise<GetRegistrationConfigResponse> {
  const request = create(GetRegistrationConfigRequestSchema, {});
  const response = await authClient.getRegistrationConfig(request);
  return response;
}

export interface LoginByPasswordParams {
  email: string;
  password: string;
}

export interface RegisterByPasswordParams {
  email: string;
  password: string;
  name: string;
  registerKey: string;
}

export async function loginByPassword(
  params: LoginByPasswordParams,
): Promise<AuthResponse> {
  const request = create(LoginByPasswordRequestSchema, {
    email: params.email,
    password: params.password,
  });

  const response = await authClient.loginByPassword(request);
  return response;
}

export async function registerByPassword(
  params: RegisterByPasswordParams,
): Promise<AuthResponse> {
  const request = create(RegisterByPasswordRequestSchema, {
    email: params.email,
    password: params.password,
    name: params.name,
    registerKey: params.registerKey,
  });

  const response = await authClient.registerByPassword(request);
  return response;
}

export async function refreshAccessToken(
  refreshToken: string,
): Promise<AuthResponse> {
  const request = create(RefreshAccessTokenRequestSchema, {
    refreshToken,
  });

  const response = await authClient.refreshAccessToken(request);
  return response;
}

export interface UpdateUserNameParams {
  newName: string;
  accessToken: string;
}

export async function updateUserName(
  params: UpdateUserNameParams,
): Promise<UpdateUserNameResponse> {
  const request = create(UpdateUserNameRequestSchema, {
    newName: params.newName,
  });

  return await userClient.updateUserName(
    request,
    authHeader(params.accessToken),
  );
}

export interface ChangePasswordParams {
  currentPassword: string;
  newPassword: string;
  accessToken: string;
}

export async function changePassword(
  params: ChangePasswordParams,
): Promise<ChangePasswordResponse> {
  const request = create(ChangePasswordRequestSchema, {
    currentPassword: params.currentPassword,
    newPassword: params.newPassword,
  });

  return await userClient.changePassword(
    request,
    authHeader(params.accessToken),
  );
}

export interface UpdateLLMKeyParams {
  llmProvider: number;
  key: string;
  accessToken: string;
}

export async function updateLLMKey(
  params: UpdateLLMKeyParams,
): Promise<UpdateLLMKeyResponse> {
  const request = create(UpdateLLMKeyRequestSchema, {
    llmProvider: params.llmProvider,
    key: params.key,
  });

  return await userClient.updateLLMKey(request, authHeader(params.accessToken));
}

export interface GetUserInfoParams {
  accessToken: string;
}

export async function getUserInfo(
  params: GetUserInfoParams,
): Promise<GetUserInfoResponse> {
  const request = create(GetUserInfoRequestSchema, {});

  return await userClient.getUserInfo(request, authHeader(params.accessToken));
}

export interface DeleteLLMKeyParams {
  llmProvider: number;
  accessToken: string;
}

export async function deleteLLMKey(
  params: DeleteLLMKeyParams,
): Promise<DeleteLLMKeyResponse> {
  const request = create(DeleteLLMKeyRequestSchema, {
    llmProvider: params.llmProvider,
  });

  return await userClient.deleteLLMKey(request, authHeader(params.accessToken));
}

export interface DeleteAccountParams {
  accessToken: string;
}

export async function deleteAccount(
  params: DeleteAccountParams,
): Promise<DeleteAccountResponse> {
  const request = create(DeleteAccountRequestSchema, {});

  return await userClient.deleteAccount(
    request,
    authHeader(params.accessToken),
  );
}

export interface UpdateAutoSummarySettingsParams {
  llmProvider: number;
  autoSummaryMonthly: boolean;
  autoLatestTrendEnabled: boolean;
  semanticSearchEnabled: boolean;
  accessToken: string;
}

export async function updateAutoSummarySettings(
  params: UpdateAutoSummarySettingsParams,
): Promise<UpdateAutoSummarySettingsResponse> {
  const request = create(UpdateAutoSummarySettingsRequestSchema, {
    llmProvider: params.llmProvider,
    autoSummaryMonthly: params.autoSummaryMonthly,
    autoLatestTrendEnabled: params.autoLatestTrendEnabled,
    semanticSearchEnabled: params.semanticSearchEnabled,
  });

  return await userClient.updateAutoSummarySettings(
    request,
    authHeader(params.accessToken),
  );
}
