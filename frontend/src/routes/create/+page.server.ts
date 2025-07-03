import { createDiaryEntry, createYMD } from "$lib/server/diary-api";
import { error, redirect } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies, url }) => {
	const accessToken = cookies.get("accessToken");

	if (!accessToken) {
		throw redirect(302, "/login");
	}

	const dateParam = url.searchParams.get("date");
	return {
		defaultDate: dateParam || new Date().toISOString().split("T")[0],
	};
};

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
			// Parse date string directly to avoid timezone issues
			const dateMatch = dateStr.match(/^(\d{4})-(\d{2})-(\d{2})$/);
			if (!dateMatch) {
				return {
					error: "Invalid date format",
				};
			}
			const [, year, month, day] = dateMatch;
			const ymd = createYMD(
				Number.parseInt(year, 10),
				Number.parseInt(month, 10),
				Number.parseInt(day, 10),
			);

			await createDiaryEntry({
				content,
				date: ymd,
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
