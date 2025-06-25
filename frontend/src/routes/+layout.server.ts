import { refreshAccessToken } from "$lib/server/auth-api";
import { isTokenExpiringSoon } from "$lib/utils/token-utils";
import type { LayoutServerLoad } from "./$types";

export const load: LayoutServerLoad = async ({ cookies, url }) => {
	let accessToken = cookies.get("accessToken");
	const refreshToken = cookies.get("refreshToken");
	let isAuthenticated = !!accessToken;

	if (refreshToken && (!accessToken || isTokenExpiringSoon(accessToken))) {
		try {
			const response = await refreshAccessToken(refreshToken);

			cookies.set("accessToken", response.accessToken, {
				path: "/",
				httpOnly: true,
				secure: false,
				sameSite: "strict",
				maxAge: 60 * 15,
			});

			cookies.set("refreshToken", response.refreshToken, {
				path: "/",
				httpOnly: true,
				secure: false,
				sameSite: "strict",
				maxAge: 60 * 60 * 24 * 30,
			});

			accessToken = response.accessToken;
			isAuthenticated = true;
		} catch (error) {
			console.error("Token refresh failed:", error);
			cookies.delete("accessToken", { path: "/" });
			cookies.delete("refreshToken", { path: "/" });
			isAuthenticated = false;
		}
	}

	return {
		isAuthenticated,
		path: url.pathname,
	};
};
