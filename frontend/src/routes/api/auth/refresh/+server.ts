import { error, json } from "@sveltejs/kit";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async ({ cookies }) => {
	const refreshToken = cookies.get("refreshToken");

	if (!refreshToken) {
		throw error(401, "No refresh token found");
	}

	const result = await ensureValidAccessToken(cookies);

	if (!result.isAuthenticated || !result.accessToken) {
		throw error(401, "Token refresh failed");
	}

	return json({
		accessToken: result.accessToken,
		refreshToken: cookies.get("refreshToken"),
	});
};
