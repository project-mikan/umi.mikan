import { create } from "@bufbuild/protobuf";
import {
	type AuthResponse,
	LoginByPasswordRequestSchema,
	RefreshAccessTokenRequestSchema,
	RegisterByPasswordRequestSchema,
} from "./auth/auth_pb.js";
import { authClient } from "./client.js";

export interface LoginCredentials {
	email: string;
	password: string;
}

export interface RegisterCredentials {
	email: string;
	password: string;
	name: string;
}

export class AuthClient {
	async login(credentials: LoginCredentials): Promise<AuthResponse> {
		const request = create(LoginByPasswordRequestSchema, {
			email: credentials.email,
			password: credentials.password,
		});

		const response = await authClient.loginByPassword(request);
		return response;
	}

	async register(credentials: RegisterCredentials): Promise<AuthResponse> {
		const request = create(RegisterByPasswordRequestSchema, {
			email: credentials.email,
			password: credentials.password,
			name: credentials.name,
		});

		const response = await authClient.registerByPassword(request);
		return response;
	}

	async refreshAccessToken(refreshToken: string): Promise<AuthResponse> {
		const request = create(RefreshAccessTokenRequestSchema, {
			refreshToken,
		});

		const response = await authClient.refreshAccessToken(request);
		return response;
	}
}

export const authService = new AuthClient();
