import { error, json } from "@sveltejs/kit";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { generateDailySummary } from "$lib/server/diary-api";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async ({ cookies, request }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw error(401, "Unauthorized");
	}

	let body: {
		diaryId: string;
		content: string;
		date: { year: number; month: number; day: number };
	};
	try {
		body = await request.json();
	} catch {
		throw error(400, "Invalid JSON body");
	}

	const { diaryId, content, date } = body;

	if (!diaryId || !content || !date) {
		throw error(400, "Missing required fields: diaryId, content, date");
	}

	if (content.length < 1000) {
		throw error(
			400,
			"Content too short for summary generation (minimum 1000 characters)",
		);
	}

	// Check if the diary date is not today (only allow summary generation for past entries)
	const today = new Date();
	const diaryDate = new Date(date.year, date.month - 1, date.day);
	if (diaryDate >= today) {
		throw error(
			400,
			"Summary generation is only allowed for past diary entries",
		);
	}

	try {
		// Call backend gRPC service for daily summary generation
		const response = await generateDailySummary({
			diaryId,
			accessToken: authResult.accessToken,
		});

		if (!response.summary) {
			throw error(500, "Failed to generate summary");
		}

		return json({
			id: response.summary.id,
			diaryId: response.summary.diaryId,
			date: {
				year: response.summary.date?.year || 0,
				month: response.summary.date?.month || 0,
				day: response.summary.date?.day || 0,
			},
			summary: response.summary.summary,
			createdAt: Number(response.summary.createdAt) * 1000,
			updatedAt: Number(response.summary.updatedAt) * 1000,
		});
	} catch (err) {
		if (err instanceof Response) {
			throw err;
		}

		console.error("Failed to generate daily summary:", err);

		// Handle specific gRPC errors
		if (err && typeof err === "object" && "code" in err) {
			if (err.code === 7) {
				// PERMISSION_DENIED
				throw error(403, "Permission denied");
			}
			if (err.code === 5) {
				// NOT_FOUND
				throw error(404, "Diary entry not found");
			}
			if (err.code === 3) {
				// INVALID_ARGUMENT
				throw error(400, "Invalid request parameters");
			}
		}

		// Check for LLM API key related errors
		if (
			(err as Error)?.message?.includes("API key") ||
			(err as Error)?.message?.includes("Gemini")
		) {
			throw error(400, { message: "Gemini API key not configured" });
		}

		throw error(500, "Failed to generate daily summary");
	}
};
