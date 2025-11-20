import { error, json } from "@sveltejs/kit";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { triggerDiaryHighlight } from "$lib/server/diary-api";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async ({ cookies, request }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw error(401, "Unauthorized");
	}

	let body: {
		diaryId: string;
	};
	try {
		body = await request.json();
	} catch {
		throw error(400, "Invalid JSON body");
	}

	const { diaryId } = body;

	if (!diaryId) {
		throw error(400, "Missing required field: diaryId");
	}

	try {
		// Call backend gRPC service for highlight generation trigger
		const response = await triggerDiaryHighlight({
			diaryId,
			accessToken: authResult.accessToken,
		});

		return json({
			queued: response.queued,
			message: response.message,
		});
	} catch (err) {
		if (err instanceof Response) {
			throw err;
		}

		console.error("Failed to trigger highlight generation:", err);

		// Handle specific gRPC errors
		if (err && typeof err === "object" && "code" in err) {
			if (err.code === 7) {
				// PERMISSION_DENIED
				throw error(403, "Permission denied");
			}
			if (err.code === 5) {
				// NOT_FOUND
				throw error(404, "Diary entry or LLM API key not found");
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

		throw error(500, "Failed to trigger highlight generation");
	}
};
