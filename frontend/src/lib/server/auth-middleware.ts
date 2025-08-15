import type { Cookies } from "@sveltejs/kit";
import { refreshAccessToken } from "$lib/server/auth-api";
import { isTokenExpiringSoon } from "$lib/utils/token-utils";

export interface AuthResult {
	accessToken: string | null;
	isAuthenticated: boolean;
}

/**
 * 共通のトークン更新処理
 * サーバーサイドのAPIエンドポイントで使用
 */
export async function ensureValidAccessToken(
	cookies: Cookies,
): Promise<AuthResult> {
	let accessToken = cookies.get("accessToken");
	const refreshToken = cookies.get("refreshToken");

	// Try to refresh token if access token is missing or expiring soon
	if (refreshToken && (!accessToken || isTokenExpiringSoon(accessToken))) {
		try {
			const response = await refreshAccessToken(refreshToken);

			// Update cookies with new tokens
			cookies.set("accessToken", response.accessToken, {
				path: "/",
				httpOnly: true,
				secure: false,
				sameSite: "strict",
				maxAge: 60 * 15, // 15 minutes
			});

			if (response.refreshToken) {
				cookies.set("refreshToken", response.refreshToken, {
					path: "/",
					httpOnly: true,
					secure: false,
					sameSite: "strict",
					maxAge: 60 * 60 * 24 * 30, // 30 days
				});
			}

			accessToken = response.accessToken;
		} catch (err) {
			console.error("Token refresh failed:", err);

			// Clear invalid tokens
			cookies.delete("accessToken", { path: "/" });
			cookies.delete("refreshToken", { path: "/" });

			return {
				accessToken: null,
				isAuthenticated: false,
			};
		}
	}

	return {
		accessToken: accessToken || null,
		isAuthenticated: !!accessToken,
	};
}
