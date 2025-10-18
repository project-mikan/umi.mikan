import { refreshAccessToken, getUserInfo } from "$lib/server/auth-api";
import { isTokenExpiringSoon } from "$lib/utils/token-utils";
import {
	ACCESS_TOKEN_COOKIE_OPTIONS,
	REFRESH_TOKEN_COOKIE_OPTIONS,
} from "$lib/utils/cookie-utils";
import { setCSRFToken, getCSRFToken } from "$lib/server/csrf";
import type { LayoutServerLoad } from "./$types";

export const load: LayoutServerLoad = async ({ cookies, url }) => {
	let accessToken = cookies.get("accessToken");
	const refreshToken = cookies.get("refreshToken");
	let isAuthenticated = !!accessToken;

	if (refreshToken && (!accessToken || isTokenExpiringSoon(accessToken))) {
		try {
			const response = await refreshAccessToken(refreshToken);

			cookies.set(
				"accessToken",
				response.accessToken,
				ACCESS_TOKEN_COOKIE_OPTIONS,
			);

			if (response.refreshToken) {
				cookies.set(
					"refreshToken",
					response.refreshToken,
					REFRESH_TOKEN_COOKIE_OPTIONS,
				);
			}

			accessToken = response.accessToken;
			isAuthenticated = true;
		} catch (error) {
			console.error("Token refresh failed:", error);
			cookies.delete("accessToken", { path: "/" });
			cookies.delete("refreshToken", { path: "/" });
			isAuthenticated = false;
		}
	}

	// CSRFトークンを設定・取得
	let csrfToken = getCSRFToken(cookies);
	if (!csrfToken) {
		csrfToken = setCSRFToken(cookies);
	}

	// ユーザー情報を取得
	let userName: string | null = null;
	let autoLatestTrendEnabled = false;
	if (isAuthenticated && accessToken) {
		try {
			const userInfo = await getUserInfo({ accessToken });
			userName = userInfo.name;
			// LLMキー情報から autoLatestTrendEnabled を取得（Gemini provider=1のみ）
			const geminiKey = userInfo.llmKeys?.find((key) => key.llmProvider === 1);
			autoLatestTrendEnabled = geminiKey?.autoLatestTrendEnabled || false;
		} catch (error) {
			console.error("Failed to get user info:", error);
		}
	}

	return {
		isAuthenticated,
		path: url.pathname,
		csrfToken,
		userName,
		autoLatestTrendEnabled,
	};
};
