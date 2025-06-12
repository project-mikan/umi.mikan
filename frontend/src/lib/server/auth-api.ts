// gRPC APIを使用したバックエンドとの通信
import { createAuthClient, promisifyGrpcCall } from './grpc-client';
import type { LoginByPasswordRequest } from '../grpc/auth/LoginByPasswordRequest';
import type { RegisterByPasswordRequest } from '../grpc/auth/RegisterByPasswordRequest';
import type { AuthResponse__Output } from '../grpc/auth/AuthResponse';
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

interface AuthResponse {
	access_token: string;
	refresh_token: string;
	token_type: string;
	expires_in: number;
}

export async function loginByPassword(request: LoginRequest): Promise<AuthResponse> {
	try {
		const client = createAuthClient();
		
		const grpcRequest: LoginByPasswordRequest = {
			email: request.email,
			password: request.password
		};

		const response = await promisifyGrpcCall<LoginByPasswordRequest, AuthResponse__Output>(
			client,
			'loginByPassword',
			grpcRequest
		);

		client.close();

		return {
			access_token: response.access_token,
			refresh_token: response.refresh_token,
			token_type: response.token_type,
			expires_in: response.expires_in
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

export async function registerByPassword(request: RegisterRequest): Promise<AuthResponse> {
	try {
		const client = createAuthClient();
		
		const grpcRequest: RegisterByPasswordRequest = {
			email: request.email,
			password: request.password,
			name: request.name
		};

		const response = await promisifyGrpcCall<RegisterByPasswordRequest, AuthResponse__Output>(
			client,
			'registerByPassword',
			grpcRequest
		);

		client.close();

		return {
			access_token: response.access_token,
			refresh_token: response.refresh_token,
			token_type: response.token_type,
			expires_in: response.expires_in
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