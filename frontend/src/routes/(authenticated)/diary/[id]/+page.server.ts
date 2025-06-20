import { createYMD, getDiaryEntry } from "$lib/server/diary-api";
import { error } from "@sveltejs/kit";
import type { PageServerLoad } from "./$types";

const DATE_REGEX = /^(\d{4})-(\d{2})-(\d{2})$/;

export const load: PageServerLoad = async ({ params, cookies }) => {
	const accessToken = cookies.get("accessToken");

	if (!accessToken) {
		throw error(401, "Unauthorized");
	}

	try {
		// params.id should be in format YYYY-MM-DD
		const dateMatch = params.id.match(DATE_REGEX);
		if (!dateMatch) {
			throw error(400, "Invalid date format");
		}

		const [, year, month, day] = dateMatch;
		const response = await getDiaryEntry({
			date: createYMD(
				Number.parseInt(year, 10),
				Number.parseInt(month, 10),
				Number.parseInt(day, 10),
			),
			accessToken,
		});

		if (!response.entry) {
			throw error(404, "Diary entry not found");
		}

		return {
			entry: response.entry,
		};
	} catch (err) {
		if (err instanceof Response) {
			throw err;
		}
		// Log error for debugging but don't expose details to client
		throw error(500, "Failed to load diary entry");
	}
};
