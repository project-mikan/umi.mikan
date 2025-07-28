import { error, redirect } from "@sveltejs/kit";
import {
	createDiaryEntry,
	createYMD,
	getDiaryEntry,
	updateDiaryEntry,
} from "$lib/server/diary-api";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies }) => {
	const accessToken = cookies.get("accessToken");

	if (!accessToken) {
		throw redirect(302, "/login");
	}

	try {
		const now = new Date();
		const today = createYMD(
			now.getFullYear(),
			now.getMonth() + 1,
			now.getDate(),
		);

		const yesterday = new Date(now);
		yesterday.setDate(yesterday.getDate() - 1);
		const yesterdayYMD = createYMD(
			yesterday.getFullYear(),
			yesterday.getMonth() + 1,
			yesterday.getDate(),
		);

		const dayBeforeYesterday = new Date(now);
		dayBeforeYesterday.setDate(dayBeforeYesterday.getDate() - 2);
		const dayBeforeYesterdayYMD = createYMD(
			dayBeforeYesterday.getFullYear(),
			dayBeforeYesterday.getMonth() + 1,
			dayBeforeYesterday.getDate(),
		);

		// 3日分の日記を並行して取得
		const [todayResult, yesterdayResult, dayBeforeYesterdayResult] =
			await Promise.allSettled([
				getDiaryEntry({ date: today, accessToken }),
				getDiaryEntry({ date: yesterdayYMD, accessToken }),
				getDiaryEntry({ date: dayBeforeYesterdayYMD, accessToken }),
			]);

		return {
			today: {
				date: today,
				entry:
					todayResult.status === "fulfilled" ? todayResult.value.entry : null,
			},
			yesterday: {
				date: yesterdayYMD,
				entry:
					yesterdayResult.status === "fulfilled"
						? yesterdayResult.value.entry
						: null,
			},
			dayBeforeYesterday: {
				date: dayBeforeYesterdayYMD,
				entry:
					dayBeforeYesterdayResult.status === "fulfilled"
						? dayBeforeYesterdayResult.value.entry
						: null,
			},
		};
	} catch (err) {
		console.error("Failed to load diary entries:", err);
		const now = new Date();
		return {
			today: {
				date: createYMD(now.getFullYear(), now.getMonth() + 1, now.getDate()),
				entry: null,
			},
			yesterday: {
				date: createYMD(
					now.getFullYear(),
					now.getMonth() + 1,
					now.getDate() - 1,
				),
				entry: null,
			},
			dayBeforeYesterday: {
				date: createYMD(
					now.getFullYear(),
					now.getMonth() + 1,
					now.getDate() - 2,
				),
				entry: null,
			},
		};
	}
};

export const actions: Actions = {
	saveToday: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");

		if (!accessToken) {
			throw error(401, "Unauthorized");
		}

		const data = await request.formData();
		const content = data.get("content")?.toString();
		const dateStr = data.get("date")?.toString();
		const id = data.get("id")?.toString();

		if (!content || !dateStr) {
			return { error: "内容と日付は必須です" };
		}

		try {
			const [year, month, day] = dateStr.split("-").map(Number);
			const date = createYMD(year, month, day);

			if (id) {
				// 更新
				await updateDiaryEntry({
					id,
					title: "",
					content,
					date,
					accessToken,
				});
			} else {
				// 新規作成
				await createDiaryEntry({
					content,
					date,
					accessToken,
				});
			}

			return { success: true };
		} catch (err) {
			console.error("Failed to save diary entry:", err);
			return { error: "日記の保存に失敗しました" };
		}
	},
	saveYesterday: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");

		if (!accessToken) {
			throw error(401, "Unauthorized");
		}

		const data = await request.formData();
		const content = data.get("content")?.toString();
		const dateStr = data.get("date")?.toString();
		const id = data.get("id")?.toString();

		if (!content || !dateStr) {
			return { error: "内容と日付は必須です" };
		}

		try {
			const [year, month, day] = dateStr.split("-").map(Number);
			const date = createYMD(year, month, day);

			if (id) {
				await updateDiaryEntry({
					id,
					title: "",
					content,
					date,
					accessToken,
				});
			} else {
				await createDiaryEntry({
					content,
					date,
					accessToken,
				});
			}

			return { success: true };
		} catch (err) {
			console.error("Failed to save diary entry:", err);
			return { error: "日記の保存に失敗しました" };
		}
	},
	saveDayBeforeYesterday: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");

		if (!accessToken) {
			throw error(401, "Unauthorized");
		}

		const data = await request.formData();
		const content = data.get("content")?.toString();
		const dateStr = data.get("date")?.toString();
		const id = data.get("id")?.toString();

		if (!content || !dateStr) {
			return { error: "内容と日付は必須です" };
		}

		try {
			const [year, month, day] = dateStr.split("-").map(Number);
			const date = createYMD(year, month, day);

			if (id) {
				await updateDiaryEntry({
					id,
					title: "",
					content,
					date,
					accessToken,
				});
			} else {
				await createDiaryEntry({
					content,
					date,
					accessToken,
				});
			}

			return { success: true };
		} catch (err) {
			console.error("Failed to save diary entry:", err);
			return { error: "日記の保存に失敗しました" };
		}
	},
	logout: async ({ cookies }) => {
		cookies.delete("accessToken", { path: "/" });
		cookies.delete("refreshToken", { path: "/" });
		throw redirect(302, "/login");
	},
};
