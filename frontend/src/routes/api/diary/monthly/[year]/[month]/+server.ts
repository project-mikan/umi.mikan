import { error, json } from "@sveltejs/kit";
import { createYM, getDiaryEntriesByMonth } from "$lib/server/diary-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({ cookies, params }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw error(401, "Unauthorized");
	}

	const year = Number.parseInt(params.year, 10);
	const month = Number.parseInt(params.month, 10);

	if (Number.isNaN(year) || Number.isNaN(month) || month < 1 || month > 12) {
		throw error(400, "Invalid year or month");
	}

	try {
		const entries = await getDiaryEntriesByMonth({
			month: createYM(year, month),
			accessToken: authResult.accessToken,
		});

		return json(entries);
	} catch (err) {
		console.error("Failed to load diary entries:", err);
		throw error(500, "Failed to load diary entries");
	}
};
