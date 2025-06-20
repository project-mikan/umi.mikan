import { createDiaryEntry, createYMD } from "$lib/server/diary-api";
import { error, redirect } from "@sveltejs/kit";
import type { Actions } from "./$types";

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");

		if (!accessToken) {
			throw error(401, "Unauthorized");
		}

		const data = await request.formData();
		const content = data.get("content") as string;
		const dateStr = data.get("date") as string;

		if (!content || !dateStr) {
			return {
				error: "コンテンツと日付は必須です",
			};
		}

		try {
			const date = new Date(dateStr);
			await createDiaryEntry({
				content,
				date: createYMD(
					date.getFullYear(),
					date.getMonth() + 1,
					date.getDate(),
				),
				accessToken,
			});
		} catch (err) {
			if (err instanceof Response) {
				throw err;
			}
			console.error("Failed to create diary entry:", err);
			return {
				error: "日記の作成に失敗しました",
			};
		}

		redirect(303, "/diary");
	},
};
