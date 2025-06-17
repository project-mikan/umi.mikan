import {
	type LoginCredentials,
	type RegisterCredentials,
	authService,
} from "./auth-client.js";
import type { AuthResponse } from "./auth/auth_pb.js";
import { getAccessToken, setAccessToken } from "./client.js";
import {
	type CreateDiaryEntryParams,
	type SearchDiaryEntriesParams,
	type UpdateDiaryEntryParams,
	createYM,
	createYMD,
	diaryService,
} from "./diary-client.js";

export class ApiClient {
	private refreshToken: string | null = null;

	async login(credentials: LoginCredentials): Promise<AuthResponse> {
		const response = await authService.login(credentials);
		this.setTokens(response);
		return response;
	}

	async register(credentials: RegisterCredentials): Promise<AuthResponse> {
		const response = await authService.register(credentials);
		this.setTokens(response);
		return response;
	}

	async refreshAccessToken(): Promise<AuthResponse | null> {
		if (!this.refreshToken) {
			throw new Error("No refresh token available");
		}

		try {
			const response = await authService.refreshAccessToken(this.refreshToken);
			this.setTokens(response);
			return response;
		} catch (error) {
			this.clearTokens();
			throw error;
		}
	}

	logout() {
		this.clearTokens();
	}

	private setTokens(authResponse: AuthResponse) {
		setAccessToken(authResponse.accessToken);
		this.refreshToken = authResponse.refreshToken;

		if (typeof window !== "undefined") {
			localStorage.setItem("refreshToken", authResponse.refreshToken);
		}
	}

	private clearTokens() {
		setAccessToken(null);
		this.refreshToken = null;

		if (typeof window !== "undefined") {
			localStorage.removeItem("refreshToken");
		}
	}

	initializeFromStorage() {
		if (typeof window !== "undefined") {
			this.refreshToken = localStorage.getItem("refreshToken");
		}
	}

	isAuthenticated(): boolean {
		return getAccessToken() !== null;
	}

	get diary() {
		return diaryService;
	}
}

export const apiClient = new ApiClient();

export {
	createYMD,
	createYM,
	type LoginCredentials,
	type RegisterCredentials,
	type CreateDiaryEntryParams,
	type UpdateDiaryEntryParams,
	type SearchDiaryEntriesParams,
};

export * from "./auth/auth_pb.js";
export * from "./diary/diary_pb.js";
