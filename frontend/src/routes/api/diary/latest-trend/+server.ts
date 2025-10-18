import { error, json } from "@sveltejs/kit";
import { getLatestTrend } from "$lib/server/diary-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({ cookies }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw error(401, "Unauthorized");
	}

	try {
		const trendResponse = await getLatestTrend({
			accessToken: authResult.accessToken,
		});

		return json({
			analysis: trendResponse.analysis,
			periodStart: trendResponse.periodStart,
			periodEnd: trendResponse.periodEnd,
			generatedAt: trendResponse.generatedAt,
		});
	} catch (err) {
		console.error("Failed to get latest trend:", err);
		console.error("Error details:", {
			message: (err as Error)?.message,
			code: (err as { code?: string })?.code,
			stack: (err as Error)?.stack,
		});

		if (
			(err as { code?: string })?.code === "NOT_FOUND" ||
			(err as { code?: number })?.code === 5 || // gRPC NOT_FOUND code
			(err as Error)?.message?.includes("not found")
		) {
			throw error(404, "Latest trend analysis not found");
		}
		throw error(500, "Failed to get latest trend");
	}
};
