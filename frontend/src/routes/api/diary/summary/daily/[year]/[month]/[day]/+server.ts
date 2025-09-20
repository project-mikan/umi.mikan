import { error, json } from "@sveltejs/kit";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { getDailySummary, createYMD } from "$lib/server/diary-api";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({ cookies, params }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw error(401, "Unauthorized");
	}

	const { year, month, day } = params;

	// Validate date parameters
	const yearNum = Number.parseInt(year, 10);
	const monthNum = Number.parseInt(month, 10);
	const dayNum = Number.parseInt(day, 10);

	if (
		Number.isNaN(yearNum) ||
		Number.isNaN(monthNum) ||
		Number.isNaN(dayNum) ||
		monthNum < 1 ||
		monthNum > 12 ||
		dayNum < 1 ||
		dayNum > 31
	) {
		throw error(400, "Invalid date parameters");
	}

	try {
		const response = await getDailySummary({
			date: createYMD(yearNum, monthNum, dayNum),
			accessToken: authResult.accessToken,
		});

		return json(response);
	} catch (err) {
		if (err instanceof Response) {
			throw err;
		}

		console.error("Failed to get daily summary:", err);
		throw error(500, "Failed to get daily summary");
	}
};
