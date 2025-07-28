import { error, redirect } from "@sveltejs/kit";
import { createYM, getDiaryEntriesByMonth } from "$lib/server/diary-api";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies, params }) => {
	const accessToken = cookies.get("accessToken");

	if (!accessToken) {
		throw redirect(302, "/login");
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

		return {
			entries,
			year,
			month,
		};
	} catch (err) {
		console.error("Failed to load diary entries:", err);
		return {
			entries: [],
			year,
			month,
		};
	}
};
