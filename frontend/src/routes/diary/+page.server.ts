import { getDiaryEntriesByMonth, createYM } from "$lib/server/diary-api";
import { error } from "@sveltejs/kit";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies }) => {
	const accessToken = cookies.get("accessToken");

	if (!accessToken) {
		throw error(401, "Unauthorized");
	}

	try {
		// 今月の日記エントリを取得
		const now = new Date();
		const entries = await getDiaryEntriesByMonth({
			month: createYM(now.getFullYear(), now.getMonth() + 1),
			accessToken
		});

		return {
			entries,
		};
	} catch (err) {
		console.error("Failed to load diary entries:", err);
		return {
			entries: [],
		};
	}
};
