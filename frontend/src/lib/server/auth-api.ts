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

const transport = createGrpcTransport({
	baseUrl: "http://backend:8080",
});

const authClient = createClient(AuthService, transport);

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
	const transport = createGrpcTransport({
		baseUrl: "http://backend:8080",
		interceptors: [
			(next) => (req) => {
				req.header.set("authorization", `Bearer ${params.accessToken}`);
				return next(req);
			},
		],
	});

	const userClient = createClient(UserService, transport);
	const request = create(UpdateUserNameRequestSchema, {
		newName: params.newName,
	});

	const response = await userClient.updateUserName(request);
	return response;
}

export interface ChangePasswordParams {
	currentPassword: string;
	newPassword: string;
	accessToken: string;
}

export async function changePassword(
	params: ChangePasswordParams,
): Promise<ChangePasswordResponse> {
	const transport = createGrpcTransport({
		baseUrl: "http://backend:8080",
		interceptors: [
			(next) => (req) => {
				req.header.set("authorization", `Bearer ${params.accessToken}`);
				return next(req);
			},
		],
	});

	const userClient = createClient(UserService, transport);
	const request = create(ChangePasswordRequestSchema, {
		currentPassword: params.currentPassword,
		newPassword: params.newPassword,
	});

	const response = await userClient.changePassword(request);
	return response;
}

export interface UpdateLLMKeyParams {
	llmProvider: number;
	key: string;
	accessToken: string;
}

export async function updateLLMKey(
	params: UpdateLLMKeyParams,
): Promise<UpdateLLMKeyResponse> {
	const transport = createGrpcTransport({
		baseUrl: "http://backend:8080",
		interceptors: [
			(next) => (req) => {
				req.header.set("authorization", `Bearer ${params.accessToken}`);
				return next(req);
			},
		],
	});

	const userClient = createClient(UserService, transport);
	const request = create(UpdateLLMKeyRequestSchema, {
		llmProvider: params.llmProvider,
		key: params.key,
	});

	const response = await userClient.updateLLMKey(request);
	return response;
}

export interface GetUserInfoParams {
	accessToken: string;
}

export async function getUserInfo(
	params: GetUserInfoParams,
): Promise<GetUserInfoResponse> {
	const transport = createGrpcTransport({
		baseUrl: "http://backend:8080",
		interceptors: [
			(next) => (req) => {
				req.header.set("authorization", `Bearer ${params.accessToken}`);
				return next(req);
			},
		],
	});

	const userClient = createClient(UserService, transport);
	const request = create(GetUserInfoRequestSchema, {});

	const response = await userClient.getUserInfo(request);
	return response;
}

export interface DeleteLLMKeyParams {
	llmProvider: number;
	accessToken: string;
}

export async function deleteLLMKey(
	params: DeleteLLMKeyParams,
): Promise<DeleteLLMKeyResponse> {
	const transport = createGrpcTransport({
		baseUrl: "http://backend:8080",
		interceptors: [
			(next) => (req) => {
				req.header.set("authorization", `Bearer ${params.accessToken}`);
				return next(req);
			},
		],
	});

	const userClient = createClient(UserService, transport);
	const request = create(DeleteLLMKeyRequestSchema, {
		llmProvider: params.llmProvider,
	});

	const response = await userClient.deleteLLMKey(request);
	return response;
}

export interface DeleteAccountParams {
	accessToken: string;
}

export async function deleteAccount(
	params: DeleteAccountParams,
): Promise<DeleteAccountResponse> {
	const transport = createGrpcTransport({
		baseUrl: "http://backend:8080",
		interceptors: [
			(next) => (req) => {
				req.header.set("authorization", `Bearer ${params.accessToken}`);
				return next(req);
			},
		],
	});

	const userClient = createClient(UserService, transport);
	const request = create(DeleteAccountRequestSchema, {});

	const response = await userClient.deleteAccount(request);
	return response;
}

export interface UpdateAutoSummarySettingsParams {
	llmProvider: number;
	autoSummaryDaily: boolean;
	autoSummaryMonthly: boolean;
	accessToken: string;
}

export async function updateAutoSummarySettings(
	params: UpdateAutoSummarySettingsParams,
): Promise<UpdateAutoSummarySettingsResponse> {
	const transport = createGrpcTransport({
		baseUrl: "http://backend:8080",
		interceptors: [
			(next) => (req) => {
				req.header.set("authorization", `Bearer ${params.accessToken}`);
				return next(req);
			},
		],
	});

	const userClient = createClient(UserService, transport);
	const request = create(UpdateAutoSummarySettingsRequestSchema, {
		llmProvider: params.llmProvider,
		autoSummaryDaily: params.autoSummaryDaily,
		autoSummaryMonthly: params.autoSummaryMonthly,
	});

	const response = await userClient.updateAutoSummarySettings(request);
	return response;
}
