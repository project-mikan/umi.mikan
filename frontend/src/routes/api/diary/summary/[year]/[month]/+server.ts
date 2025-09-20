import { error, json } from "@sveltejs/kit";
import { createYM, getMonthlySummary } from "$lib/server/diary-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { unixToMilliseconds } from "$lib/utils/token-utils";
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
		const summaryResponse = await getMonthlySummary({
			month: createYM(year, month),
			accessToken: authResult.accessToken,
		});

		if (!summaryResponse.summary) {
			throw error(404, "Summary not found");
		}

		return json({
			id: summaryResponse.summary.id,
			month: {
				year: summaryResponse.summary.month?.year,
				month: summaryResponse.summary.month?.month,
			},
			summary: summaryResponse.summary.summary,
			createdAt: unixToMilliseconds(summaryResponse.summary.createdAt),
			updatedAt: unixToMilliseconds(summaryResponse.summary.updatedAt),
		});
	} catch (err) {
		console.error("Failed to load monthly summary:", err);
		if ((err as { code?: string })?.code === "NOT_FOUND") {
			throw error(404, "Summary not found");
		}
		throw error(500, "Failed to load monthly summary");
	}
};
