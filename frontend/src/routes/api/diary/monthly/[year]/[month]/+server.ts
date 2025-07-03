import { createYM, getDiaryEntriesByMonth } from "$lib/server/diary-api";
import { error, json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const GET: RequestHandler = async ({ cookies, params }) => {
	const accessToken = cookies.get("accessToken");

	if (!accessToken) {
		throw error(401, "Unauthorized");
	}

	const year = Number.parseInt(params.year);
	const month = Number.parseInt(params.month);

	if (Number.isNaN(year) || Number.isNaN(month) || month < 1 || month > 12) {
		throw error(400, "Invalid year or month");
	}

	try {
		const entries = await getDiaryEntriesByMonth({
			month: createYM(year, month),
			accessToken,
		});

		return json(entries);
	} catch (err) {
		console.error("Failed to load diary entries:", err);
		throw error(500, "Failed to load diary entries");
	}
};
