import { error, json } from "@sveltejs/kit";
import { createYM, getDiaryEntriesByMonth } from "$lib/server/diary-api";
import { refreshAccessToken } from "$lib/server/auth-api";
import { isTokenExpiringSoon } from "$lib/utils/token-utils";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({ cookies, params }) => {
	let accessToken = cookies.get("accessToken");
	const refreshToken = cookies.get("refreshToken");

	// Try to refresh token if access token is missing or expiring soon
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

			if (response.refreshToken) {
				cookies.set("refreshToken", response.refreshToken, {
					path: "/",
					httpOnly: true,
					secure: false,
					sameSite: "strict",
					maxAge: 60 * 60 * 24 * 30,
				});
			}

			accessToken = response.accessToken;
		} catch (err) {
			console.error("Token refresh failed:", err);
			cookies.delete("accessToken", { path: "/" });
			cookies.delete("refreshToken", { path: "/" });
			throw error(401, "Unauthorized");
		}
	}

	if (!accessToken) {
		throw error(401, "Unauthorized");
	}

	const year = Number.parseInt(params.year);
	const month = Number.parseInt(params.month);

	if (Number.isNaN(year) || Number.isNaN(month) || month < 1 || month > 12) {
		throw error(400, "Invalid year or month");
	}

	try {
		const entries = await getDiaryEntriesByMonth({
			month: createYM(year, month),
			accessToken,
		});

		return json(entries);
	} catch (err) {
		console.error("Failed to load diary entries:", err);
		throw error(500, "Failed to load diary entries");
	}
};
