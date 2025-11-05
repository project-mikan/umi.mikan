import { error, json } from "@sveltejs/kit";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { getDiaryHighlight, deleteDiaryHighlight } from "$lib/server/diary-api";
import { unixToMilliseconds } from "$lib/utils/token-utils";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({ cookies, params }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw error(401, "Unauthorized");
	}

	const { diaryId } = params;

	if (!diaryId) {
		throw error(400, "Missing required parameter: diaryId");
	}

	try {
		// Call backend gRPC service to get highlight
		const response = await getDiaryHighlight({
			diaryId,
			accessToken: authResult.accessToken,
		});

		return json({
			highlights: response.highlights.map((h) => ({
				start: h.start,
				end: h.end,
				text: h.text,
			})),
			createdAt: unixToMilliseconds(response.createdAt),
			updatedAt: unixToMilliseconds(response.updatedAt),
		});
	} catch (err) {
		if (err instanceof Response) {
			throw err;
		}

		console.error("Failed to get diary highlight:", err);

		// Handle specific gRPC errors
		if (err && typeof err === "object" && "code" in err) {
			if (err.code === 7) {
				// PERMISSION_DENIED
				throw error(403, "Permission denied");
			}
			if (err.code === 5) {
				// NOT_FOUND
				throw error(404, "Highlight not found or diary has been updated");
			}
			if (err.code === 3) {
				// INVALID_ARGUMENT
				throw error(400, "Invalid request parameters");
			}
		}

		throw error(500, "Failed to get diary highlight");
	}
};

export const DELETE: RequestHandler = async ({ cookies, params }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw error(401, "Unauthorized");
	}

	const { diaryId } = params;

	if (!diaryId) {
		throw error(400, "Missing required parameter: diaryId");
	}

	try {
		// Call backend gRPC service to delete highlight
		const response = await deleteDiaryHighlight({
			diaryId,
			accessToken: authResult.accessToken,
		});

		return json({
			success: response.success,
		});
	} catch (err) {
		if (err instanceof Response) {
			throw err;
		}

		console.error("Failed to delete diary highlight:", err);

		// Handle specific gRPC errors
		if (err && typeof err === "object" && "code" in err) {
			if (err.code === 7) {
				// PERMISSION_DENIED
				throw error(403, "Permission denied");
			}
			if (err.code === 5) {
				// NOT_FOUND
				throw error(404, "Highlight not found");
			}
			if (err.code === 3) {
				// INVALID_ARGUMENT
				throw error(400, "Invalid request parameters");
			}
		}

		throw error(500, "Failed to delete diary highlight");
	}
};
