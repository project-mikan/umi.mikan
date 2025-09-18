import { create } from "@bufbuild/protobuf";
import { createClient } from "@connectrpc/connect";
import { createGrpcTransport } from "@connectrpc/connect-node";
import {
	type AuthResponse,
	AuthService,
	LoginByPasswordRequestSchema,
	RefreshAccessTokenRequestSchema,
	RegisterByPasswordRequestSchema,
	UpdateUserNameRequestSchema,
	ChangePasswordRequestSchema,
	UpdateLLMTokenRequestSchema,
	GetUserInfoRequestSchema,
	DeleteLLMTokenRequestSchema,
	DeleteAccountRequestSchema,
	type UpdateUserNameResponse,
	type ChangePasswordResponse,
	type UpdateLLMTokenResponse,
	type GetUserInfoResponse,
	type DeleteLLMTokenResponse,
	type DeleteAccountResponse,
} from "$lib/grpc/auth/auth_pb.js";

const transport = createGrpcTransport({
	baseUrl: "http://backend:8080",
});

const authClient = createClient(AuthService, transport);

export interface LoginByPasswordParams {
	email: string;
	password: string;
}

export interface RegisterByPasswordParams {
	email: string;
	password: string;
	name: string;
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

	const authClient = createClient(AuthService, transport);
	const request = create(UpdateUserNameRequestSchema, {
		newName: params.newName,
	});

	const response = await authClient.updateUserName(request);
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

	const authClient = createClient(AuthService, transport);
	const request = create(ChangePasswordRequestSchema, {
		currentPassword: params.currentPassword,
		newPassword: params.newPassword,
	});

	const response = await authClient.changePassword(request);
	return response;
}

export interface UpdateLLMTokenParams {
	llmProvider: number;
	token: string;
	accessToken: string;
}

export async function updateLLMToken(
	params: UpdateLLMTokenParams,
): Promise<UpdateLLMTokenResponse> {
	const transport = createGrpcTransport({
		baseUrl: "http://backend:8080",
		interceptors: [
			(next) => (req) => {
				req.header.set("authorization", `Bearer ${params.accessToken}`);
				return next(req);
			},
		],
	});

	const authClient = createClient(AuthService, transport);
	const request = create(UpdateLLMTokenRequestSchema, {
		llmProvider: params.llmProvider,
		token: params.token,
	});

	const response = await authClient.updateLLMToken(request);
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

	const authClient = createClient(AuthService, transport);
	const request = create(GetUserInfoRequestSchema, {});

	const response = await authClient.getUserInfo(request);
	return response;
}

export interface DeleteLLMTokenParams {
	llmProvider: number;
	accessToken: string;
}

export async function deleteLLMToken(
	params: DeleteLLMTokenParams,
): Promise<DeleteLLMTokenResponse> {
	const transport = createGrpcTransport({
		baseUrl: "http://backend:8080",
		interceptors: [
			(next) => (req) => {
				req.header.set("authorization", `Bearer ${params.accessToken}`);
				return next(req);
			},
		],
	});

	const authClient = createClient(AuthService, transport);
	const request = create(DeleteLLMTokenRequestSchema, {
		llmProvider: params.llmProvider,
	});

	const response = await authClient.deleteLLMToken(request);
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

	const authClient = createClient(AuthService, transport);
	const request = create(DeleteAccountRequestSchema, {});

	const response = await authClient.deleteAccount(request);
	return response;
}
