import { error } from "@sveltejs/kit";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
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
		// TODO: Call backend gRPC service to get daily summary
		// For now, return mock data indicating no summary exists

		// Check if summary exists in storage (mock implementation)
		// In the future, this should query the backend for daily summaries

		throw error(404, "Daily summary not found");
	} catch (err) {
		if (err instanceof Response) {
			throw err;
		}

		console.error("Failed to get daily summary:", err);
		throw error(500, "Failed to get daily summary");
	}
};
