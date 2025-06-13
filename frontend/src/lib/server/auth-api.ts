// gRPC APIを使用したバックエンドとの通信
import { createAuthClient, promisifyGrpcCall } from './grpc-client';
import { LoginByPasswordRequest } from '../grpc/auth/auth_pb';
import { RegisterByPasswordRequest } from '../grpc/auth/auth_pb';
import { RefreshAccessTokenRequest } from '../grpc/auth/auth_pb';
import { AuthResponse } from '../grpc/auth/auth_pb';
import * as grpc from '@grpc/grpc-js';

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

export async function loginByPassword(request: LoginRequest): Promise<AuthResult> {
	try {
		const client = createAuthClient();
		
		const grpcRequest = new LoginByPasswordRequest();
		grpcRequest.setEmail(request.email);
		grpcRequest.setPassword(request.password);

		const response = await promisifyGrpcCall<LoginByPasswordRequest, AuthResponse>(
			client,
			'loginByPassword',
			grpcRequest
		);

		// client.close(); // Close method not available on generated client

		return {
			access_token: response.getAccessToken(),
			refresh_token: response.getRefreshToken(),
			token_type: response.getTokenType(),
			expires_in: response.getExpiresIn()
		};
	} catch (error) {
		if (error instanceof Error && 'code' in error) {
			const grpcError = error as grpc.ServiceError;
			
			if (grpcError.code === grpc.status.UNAUTHENTICATED) {
				throw new Error('Invalid credentials');
			} else if (grpcError.code === grpc.status.UNAVAILABLE) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Login failed: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

export async function registerByPassword(request: RegisterRequest): Promise<AuthResult> {
	try {
		const client = createAuthClient();
		
		const grpcRequest = new RegisterByPasswordRequest();
		grpcRequest.setEmail(request.email);
		grpcRequest.setPassword(request.password);
		grpcRequest.setName(request.name);

		const response = await promisifyGrpcCall<RegisterByPasswordRequest, AuthResponse>(
			client,
			'registerByPassword',
			grpcRequest
		);

		// client.close(); // Close method not available on generated client

		return {
			access_token: response.getAccessToken(),
			refresh_token: response.getRefreshToken(),
			token_type: response.getTokenType(),
			expires_in: response.getExpiresIn()
		};
	} catch (error) {
		if (error instanceof Error && 'code' in error) {
			const grpcError = error as grpc.ServiceError;
			
			if (grpcError.code === grpc.status.ALREADY_EXISTS) {
				throw new Error('Email already registered');
			} else if (grpcError.code === grpc.status.INVALID_ARGUMENT) {
				throw new Error('Invalid registration data');
			} else if (grpcError.code === grpc.status.UNAVAILABLE) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Registration failed: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}

export async function refreshAccessToken(refreshToken: string): Promise<AuthResult> {
	try {
		const client = createAuthClient();
		
		const grpcRequest = new RefreshAccessTokenRequest();
		grpcRequest.setRefreshToken(refreshToken);

		const response = await promisifyGrpcCall<RefreshAccessTokenRequest, AuthResponse>(
			client,
			'refreshAccessToken',
			grpcRequest
		);

		// client.close(); // Close method not available on generated client

		return {
			access_token: response.getAccessToken(),
			refresh_token: response.getRefreshToken(),
			token_type: response.getTokenType(),
			expires_in: response.getExpiresIn()
		};
	} catch (error) {
		if (error instanceof Error && 'code' in error) {
			const grpcError = error as grpc.ServiceError;
			
			if (grpcError.code === grpc.status.UNAUTHENTICATED) {
				throw new Error('Invalid refresh token');
			} else if (grpcError.code === grpc.status.UNAVAILABLE) {
				throw new Error('Backend service unavailable');
			}
		}
		
		throw new Error('Token refresh failed: ' + (error instanceof Error ? error.message : 'Unknown error'));
	}
}