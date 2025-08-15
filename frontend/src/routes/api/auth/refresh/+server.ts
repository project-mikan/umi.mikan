import { error, json } from "@sveltejs/kit";
import { refreshAccessToken } from "$lib/server/auth-api";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async ({ cookies }) => {
	const refreshToken = cookies.get("refreshToken");

	if (!refreshToken) {
		throw error(401, "No refresh token found");
	}

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

		return json({
			accessToken: response.accessToken,
			refreshToken: response.refreshToken,
		});
	} catch (err) {
		console.error("Token refresh failed:", err);

		// Clear invalid tokens
		cookies.delete("accessToken", { path: "/" });
		cookies.delete("refreshToken", { path: "/" });

		throw error(401, "Token refresh failed");
	}
};
