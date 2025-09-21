import { error, redirect } from "@sveltejs/kit";
import { createYM, getDiaryEntriesByMonth } from "$lib/server/diary-api";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies, params }) => {
	const accessToken = cookies.get("accessToken");

	if (!accessToken) {
		throw redirect(302, "/login");
	}

	const year = Number.parseInt(params.year, 10);
	const month = Number.parseInt(params.month, 10);

	if (Number.isNaN(year) || Number.isNaN(month) || month < 1 || month > 12) {
		throw error(400, "Invalid year or month");
	}

	try {
		const entries = await getDiaryEntriesByMonth({
			month: createYM(year, month),
			accessToken,
		});

		// Convert BigInt to Number for JSON serialization
		const serializedEntries = {
			...entries,
			entries: entries.entries.map((entry) => ({
				...entry,
				createdAt: Number(entry.createdAt),
				updatedAt: Number(entry.updatedAt),
			})),
		};

		return {
			entries: serializedEntries,
			year,
			month,
		};
	} catch (err) {
		console.error("Failed to load diary entries:", err);
		return {
			entries: { entries: [] }, // Empty GetDiaryEntriesByMonthResponse structure
			year,
			month,
		};
	}
};
