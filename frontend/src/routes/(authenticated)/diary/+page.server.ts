import { createYM, getDiaryEntriesByMonth } from "$lib/server/diary-api";
import { error, redirect } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";

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
			accessToken,
		});

		return {
			entries,
		};
	} catch (err) {
		// Log error for debugging but don't expose details to client
		return {
			entries: [],
		};
	}
};

export const actions: Actions = {
	logout: ({ cookies }) => {
		cookies.delete("accessToken", { path: "/" });
		cookies.delete("refreshToken", { path: "/" });
		throw redirect(302, "/login");
	},
};
