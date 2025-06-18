import { getDiaryEntriesByMonth, createYM } from "$lib/server/diary-api";
import { error } from "@sveltejs/kit";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies, params }) => {
	const accessToken = cookies.get("accessToken");

	if (!accessToken) {
		throw error(401, "Unauthorized");
	}

	const year = parseInt(params.year);
	const month = parseInt(params.month);

	if (isNaN(year) || isNaN(month) || month < 1 || month > 12) {
		throw error(400, "Invalid year or month");
	}

	try {
		const entries = await getDiaryEntriesByMonth({
			month: createYM(year, month),
			accessToken
		});

		return {
			entries,
			year,
			month
		};
	} catch (err) {
		console.error("Failed to load diary entries:", err);
		return {
			entries: [],
			year,
			month
		};
	}
};