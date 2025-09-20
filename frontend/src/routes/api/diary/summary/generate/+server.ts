import { error, json } from "@sveltejs/kit";
import { createYM, generateMonthlySummary } from "$lib/server/diary-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async ({ cookies, request }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw error(401, "Unauthorized");
	}

	let body: { year: number; month: number };
	try {
		body = await request.json();
	} catch {
		throw error(400, "Invalid JSON body");
	}

	const { year, month } = body;

	if (
		typeof year !== "number" ||
		typeof month !== "number" ||
		month < 1 ||
		month > 12
	) {
		throw error(400, "Invalid year or month");
	}

	try {
		const summaryResponse = await generateMonthlySummary({
			month: createYM(year, month),
			accessToken: authResult.accessToken,
		});

		if (!summaryResponse.summary) {
			throw error(500, "Failed to generate summary");
		}

		return json({
			id: summaryResponse.summary.id,
			month: {
				year: summaryResponse.summary.month?.year,
				month: summaryResponse.summary.month?.month,
			},
			summary: summaryResponse.summary.summary,
			// Convert Unix timestamp (seconds) to JavaScript timestamp (milliseconds)
			createdAt: Number(summaryResponse.summary.createdAt) * 1000,
			updatedAt: Number(summaryResponse.summary.updatedAt) * 1000,
		});
	} catch (err) {
		console.error("Failed to generate monthly summary:", err);

		if ((err as Error)?.message?.includes("API key")) {
			throw error(400, { message: "Gemini API key not configured" });
		}
		if (
			(err as { code?: string })?.code === "NOT_FOUND" ||
			(err as Error)?.message?.includes("no diary entries")
		) {
			throw error(404, "No diary entries found for the specified month");
		}
		throw error(500, "Failed to generate monthly summary");
	}
};
