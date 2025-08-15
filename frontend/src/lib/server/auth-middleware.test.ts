import { describe, it, expect, vi, beforeEach } from "vitest";
import { ensureValidAccessToken } from "./auth-middleware";
import type { Cookies } from "@sveltejs/kit";

// Mock dependencies
vi.mock("$lib/server/auth-api", () => ({
	refreshAccessToken: vi.fn(),
}));

vi.mock("$lib/utils/token-utils", () => ({
	isTokenExpiringSoon: vi.fn(),
}));

const { refreshAccessToken } = await import("$lib/server/auth-api");
const { isTokenExpiringSoon } = await import("$lib/utils/token-utils");

// Create mock cookies object
function createMockCookies(
	tokens: { accessToken?: string; refreshToken?: string } = {},
) {
	const cookies: Partial<Cookies> = {
		get: vi.fn((name: string) => {
			if (name === "accessToken") return tokens.accessToken;
			if (name === "refreshToken") return tokens.refreshToken;
			return undefined;
		}),
		set: vi.fn(),
		delete: vi.fn(),
	};
	return cookies as Cookies;
}

describe("auth-middleware", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe("ensureValidAccessToken", () => {
		it("should return existing valid access token without refresh", async () => {
			const cookies = createMockCookies({
				accessToken: "valid-token",
				refreshToken: "refresh-token",
			});
			vi.mocked(isTokenExpiringSoon).mockReturnValue(false);

			const result = await ensureValidAccessToken(cookies);

			expect(result).toEqual({
				accessToken: "valid-token",
				isAuthenticated: true,
			});
			expect(refreshAccessToken).not.toHaveBeenCalled();
			expect(cookies.set).not.toHaveBeenCalled();
		});

		it("should refresh token when access token is missing", async () => {
			const cookies = createMockCookies({
				refreshToken: "refresh-token",
			});
			vi.mocked(refreshAccessToken).mockResolvedValue({
				$typeName: "auth.AuthResponse",
				accessToken: "new-access-token",
				tokenType: "Bearer",
				expiresIn: 900,
				refreshToken: "new-refresh-token",
			});

			const result = await ensureValidAccessToken(cookies);

			expect(result).toEqual({
				accessToken: "new-access-token",
				isAuthenticated: true,
			});
			expect(refreshAccessToken).toHaveBeenCalledWith("refresh-token");
			expect(cookies.set).toHaveBeenCalledWith(
				"accessToken",
				"new-access-token",
				{
					path: "/",
					httpOnly: true,
					secure: false,
					sameSite: "strict",
					maxAge: 60 * 15,
				},
			);
			expect(cookies.set).toHaveBeenCalledWith(
				"refreshToken",
				"new-refresh-token",
				{
					path: "/",
					httpOnly: true,
					secure: false,
					sameSite: "strict",
					maxAge: 60 * 60 * 24 * 30,
				},
			);
		});

		it("should refresh token when access token is expiring soon", async () => {
			const cookies = createMockCookies({
				accessToken: "expiring-token",
				refreshToken: "refresh-token",
			});
			vi.mocked(isTokenExpiringSoon).mockReturnValue(true);
			vi.mocked(refreshAccessToken).mockResolvedValue({
				$typeName: "auth.AuthResponse",
				accessToken: "new-access-token",
				tokenType: "Bearer",
				expiresIn: 900,
				refreshToken: "",
			});

			const result = await ensureValidAccessToken(cookies);

			expect(result).toEqual({
				accessToken: "new-access-token",
				isAuthenticated: true,
			});
			expect(refreshAccessToken).toHaveBeenCalledWith("refresh-token");
			expect(cookies.set).toHaveBeenCalledWith(
				"accessToken",
				"new-access-token",
				expect.any(Object),
			);
		});

		it("should not set refresh token cookie if not returned in response", async () => {
			const cookies = createMockCookies({
				refreshToken: "refresh-token",
			});
			vi.mocked(refreshAccessToken).mockResolvedValue({
				$typeName: "auth.AuthResponse",
				accessToken: "new-access-token",
				tokenType: "Bearer",
				expiresIn: 900,
				refreshToken: "",
				// No refreshToken in response
			});

			const result = await ensureValidAccessToken(cookies);

			expect(result).toEqual({
				accessToken: "new-access-token",
				isAuthenticated: true,
			});
			expect(cookies.set).toHaveBeenCalledTimes(1); // Only access token
			expect(cookies.set).toHaveBeenCalledWith(
				"accessToken",
				"new-access-token",
				expect.any(Object),
			);
		});

		it("should return unauthenticated when no refresh token available", async () => {
			const cookies = createMockCookies({}); // No tokens

			const result = await ensureValidAccessToken(cookies);

			expect(result).toEqual({
				accessToken: null,
				isAuthenticated: false,
			});
			expect(refreshAccessToken).not.toHaveBeenCalled();
		});

		it("should return unauthenticated when refresh token call fails", async () => {
			const cookies = createMockCookies({
				refreshToken: "invalid-refresh-token",
			});
			vi.mocked(refreshAccessToken).mockRejectedValue(
				new Error("Refresh failed"),
			);

			const result = await ensureValidAccessToken(cookies);

			expect(result).toEqual({
				accessToken: null,
				isAuthenticated: false,
			});
			expect(refreshAccessToken).toHaveBeenCalledWith("invalid-refresh-token");
			expect(cookies.delete).toHaveBeenCalledWith("accessToken", { path: "/" });
			expect(cookies.delete).toHaveBeenCalledWith("refreshToken", {
				path: "/",
			});
		});

		it("should handle valid access token that is not expiring", async () => {
			const cookies = createMockCookies({
				accessToken: "valid-token",
				refreshToken: "refresh-token",
			});
			vi.mocked(isTokenExpiringSoon).mockReturnValue(false);

			const result = await ensureValidAccessToken(cookies);

			expect(result).toEqual({
				accessToken: "valid-token",
				isAuthenticated: true,
			});
			expect(refreshAccessToken).not.toHaveBeenCalled();
			expect(isTokenExpiringSoon).toHaveBeenCalledWith("valid-token");
		});

		it("should handle undefined access token correctly", async () => {
			const cookies = createMockCookies({
				refreshToken: "refresh-token",
				// accessToken is undefined
			});
			vi.mocked(refreshAccessToken).mockResolvedValue({
				$typeName: "auth.AuthResponse",
				accessToken: "new-token",
				tokenType: "Bearer",
				expiresIn: 900,
				refreshToken: "",
			});

			const result = await ensureValidAccessToken(cookies);

			expect(result.accessToken).toBe("new-token");
			expect(result.isAuthenticated).toBe(true);
		});

		it("should log error when token refresh fails", async () => {
			const consoleSpy = vi
				.spyOn(console, "error")
				.mockImplementation(() => {});
			const cookies = createMockCookies({
				refreshToken: "refresh-token",
			});
			const error = new Error("Network error");
			vi.mocked(refreshAccessToken).mockRejectedValue(error);

			await ensureValidAccessToken(cookies);

			expect(consoleSpy).toHaveBeenCalledWith("Token refresh failed:", error);
			consoleSpy.mockRestore();
		});
	});
});
