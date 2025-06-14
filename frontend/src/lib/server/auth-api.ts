// gRPC APIを使用したバックエンドとの通信
import { authService } from "./grpc-client";
import type { LoginByPasswordRequest, RegisterByPasswordRequest, RefreshAccessTokenRequest, AuthResponse } from "../grpc/auth/auth_pb";

interface LoginRequest {
	email: string;
	password: string;
}

interface RegisterRequest {
	name: string;
	email: string;
	password: string;
}

interface AuthResult {
	access_token: string;
	refresh_token: string;
	token_type: string;
	expires_in: number;
}

export async function loginByPassword(
	request: LoginRequest,
): Promise<AuthResult> {
	try {
		const response = await authService.loginByPassword({
			email: request.email,
			password: request.password,
		});

		return {
			access_token: response.accessToken,
			refresh_token: response.refreshToken,
			token_type: response.tokenType,
			expires_in: response.expiresIn,
		};
	} catch (error) {
		// Handle specific gRPC errors
		if (error instanceof Error) {
			if (error.message.includes('401') || error.message.includes('UNAUTHENTICATED')) {
				throw new Error("Invalid credentials");
			} else if (error.message.includes('503') || error.message.includes('UNAVAILABLE')) {
				throw new Error("Backend service unavailable");
			}
		}

		throw new Error(
			"Login failed: " +
				(error instanceof Error ? error.message : "Unknown error"),
		);
	}
}

export async function registerByPassword(
	request: RegisterRequest,
): Promise<AuthResult> {
	try {
		const response = await authService.registerByPassword({
			email: request.email,
			password: request.password,
			name: request.name,
		});

		return {
			access_token: response.accessToken,
			refresh_token: response.refreshToken,
			token_type: response.tokenType,
			expires_in: response.expiresIn,
		};
	} catch (error) {
		// Handle specific gRPC errors
		if (error instanceof Error) {
			if (error.message.includes('409') || error.message.includes('ALREADY_EXISTS')) {
				throw new Error("Email already registered");
			} else if (error.message.includes('400') || error.message.includes('INVALID_ARGUMENT')) {
				throw new Error("Invalid registration data");
			} else if (error.message.includes('503') || error.message.includes('UNAVAILABLE')) {
				throw new Error("Backend service unavailable");
			}
		}

		throw new Error(
			"Registration failed: " +
				(error instanceof Error ? error.message : "Unknown error"),
		);
	}
}

export async function refreshAccessToken(
	refreshToken: string,
): Promise<AuthResult> {
	try {
		const response = await authService.refreshAccessToken({
			refreshToken: refreshToken,
		});

		return {
			access_token: response.accessToken,
			refresh_token: response.refreshToken,
			token_type: response.tokenType,
			expires_in: response.expiresIn,
		};
	} catch (error) {
		// Handle specific gRPC errors
		if (error instanceof Error) {
			if (error.message.includes('401') || error.message.includes('UNAUTHENTICATED')) {
				throw new Error("Invalid refresh token");
			} else if (error.message.includes('503') || error.message.includes('UNAVAILABLE')) {
				throw new Error("Backend service unavailable");
			}
		}

		throw new Error(
			"Token refresh failed: " +
				(error instanceof Error ? error.message : "Unknown error"),
		);
	}
}
