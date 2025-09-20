import { error, json } from "@sveltejs/kit";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
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
		// TODO: Call backend gRPC service for daily summary generation
		// For now, return a mock response indicating backend integration is needed

		// Simulate processing time
		await new Promise((resolve) => setTimeout(resolve, 1000));

		// Temporary mock response - should be replaced with actual gRPC call
		const summary = `${date.year}年${date.month}月${date.day}日の日記要約

主な出来事:
- ${content.substring(0, 50)}...

登場人物:
- （要約生成機能は開発中です）`;

		return json({
			id: `summary-${diaryId}-${Date.now()}`,
			diaryId,
			date,
			summary,
			createdAt: Date.now(),
		});
	} catch (err) {
		console.error("Failed to generate daily summary:", err);

		if ((err as Error)?.message?.includes("API key")) {
			throw error(400, { message: "Gemini API key not configured" });
		}

		throw error(500, "Failed to generate daily summary");
	}
};
