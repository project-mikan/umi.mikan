import { refreshAccessToken } from "$lib/server/auth-api";
import { isTokenExpiringSoon } from "$lib/utils/token-utils";
import { ACCESS_TOKEN_COOKIE_OPTIONS, REFRESH_TOKEN_COOKIE_OPTIONS } from "$lib/utils/cookie-utils";
import { setCSRFToken, getCSRFToken } from "$lib/server/csrf";
import type { LayoutServerLoad } from "./$types";

export const load: LayoutServerLoad = async ({ cookies, url }) => {
	let accessToken = cookies.get("accessToken");
	const refreshToken = cookies.get("refreshToken");
	let isAuthenticated = !!accessToken;

	if (refreshToken && (!accessToken || isTokenExpiringSoon(accessToken))) {
		try {
			const response = await refreshAccessToken(refreshToken);

			cookies.set("accessToken", response.accessToken, ACCESS_TOKEN_COOKIE_OPTIONS);

			if (response.refreshToken) {
				cookies.set("refreshToken", response.refreshToken, REFRESH_TOKEN_COOKIE_OPTIONS);
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

	return {
		isAuthenticated,
		path: url.pathname,
		csrfToken,
	};
};
