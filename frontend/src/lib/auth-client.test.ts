import { describe, it, expect, vi, beforeEach } from "vitest";
import { refreshAccessToken, authenticatedFetch } from "./auth-client";

// Mock dependencies
vi.mock("$app/environment", () => ({
	browser: true,
}));

vi.mock("$app/navigation", () => ({
	goto: vi.fn(),
}));

// Mock global fetch
const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

describe("auth-client", () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	describe("refreshAccessToken", () => {
		it("should return access token on successful refresh", async () => {
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({
					accessToken: "new-token",
					refreshToken: "new-refresh-token",
				}),
			});

			const result = await refreshAccessToken();

			expect(result).toBe("new-token");
			expect(mockFetch).toHaveBeenCalledWith("/api/auth/refresh", {
				method: "POST",
				credentials: "include",
			});
		});

		it("should return null on 401 response", async () => {
			const { goto } = await import("$app/navigation");
			mockFetch.mockResolvedValueOnce({
				ok: false,
				status: 401,
			});

			const result = await refreshAccessToken();

			expect(result).toBeNull();
			expect(goto).toHaveBeenCalledWith("/login");
		});

		it("should return null on other errors", async () => {
			mockFetch.mockResolvedValueOnce({
				ok: false,
				status: 500,
			});

			const result = await refreshAccessToken();

			expect(result).toBeNull();
		});
	});

	describe("authenticatedFetch", () => {
		it("should return response directly if not 401", async () => {
			const mockResponse = { status: 200, ok: true };
			mockFetch.mockResolvedValueOnce(mockResponse);

			const result = await authenticatedFetch("/test");

			expect(result).toBe(mockResponse);
			expect(mockFetch).toHaveBeenCalledTimes(1);
		});

		it("should retry request after successful token refresh on 401", async () => {
			// First call returns 401
			mockFetch.mockResolvedValueOnce({ status: 401, ok: false });
			// Refresh token call succeeds
			mockFetch.mockResolvedValueOnce({
				ok: true,
				json: async () => ({ accessToken: "new-token" }),
			});
			// Retry original request succeeds
			const successResponse = { status: 200, ok: true };
			mockFetch.mockResolvedValueOnce(successResponse);

			const result = await authenticatedFetch("/test");

			expect(result).toBe(successResponse);
			expect(mockFetch).toHaveBeenCalledTimes(3);
		});

		it("should return 401 response if token refresh fails", async () => {
			// First call returns 401
			const unauthorizedResponse = { status: 401, ok: false };
			mockFetch.mockResolvedValueOnce(unauthorizedResponse);
			// Refresh token call fails
			mockFetch.mockResolvedValueOnce({ status: 401, ok: false });

			const result = await authenticatedFetch("/test");

			expect(result).toBe(unauthorizedResponse);
			expect(mockFetch).toHaveBeenCalledTimes(2);
		});
	});
});
