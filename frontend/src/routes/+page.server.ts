import { error, redirect } from "@sveltejs/kit";
import {
	createDiaryEntry,
	createYMD,
	getDiaryEntry,
	updateDiaryEntry,
} from "$lib/server/diary-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import type { DiaryEntityInput } from "$lib/grpc/diary/diary_pb";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({
	cookies,
	setHeaders,
	depends,
}) => {
	// キャッシュを無効化して常に最新のデータを取得
	setHeaders({
		"cache-control": "no-store, no-cache, must-revalidate, max-age=0",
	});

	// 明示的な依存関係を設定してinvalidateAll()で確実に再読み込み
	depends("diary:home");

	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw redirect(302, "/login");
	}

	// TypeScript型アサーション: ここではaccessTokenは必ず存在する
	const accessToken: string = authResult.accessToken;

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

		// 直近7日分の日付情報を生成（古い日付から新しい日付の順）
		const recentDays = Array.from({ length: 7 }, (_, i) => {
			const date = new Date(now);
			date.setDate(date.getDate() - (6 - i)); // 6日前から今日まで
			const dayOfWeekArray = [
				"Sunday",
				"Monday",
				"Tuesday",
				"Wednesday",
				"Thursday",
				"Friday",
				"Saturday",
			] as const;
			return {
				date: `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`,
				ymd: createYMD(date.getFullYear(), date.getMonth() + 1, date.getDate()),
				dayOfWeek: dayOfWeekArray[date.getDay()],
				dayOfMonth: date.getDate(),
			};
		});

		// 3日分の日記と直近7日分の日記を並行して取得
		const [
			todayResult,
			yesterdayResult,
			dayBeforeYesterdayResult,
			...recentDaysResults
		] = await Promise.allSettled([
			getDiaryEntry({ date: today, accessToken }),
			getDiaryEntry({
				date: yesterdayYMD,
				accessToken,
			}),
			getDiaryEntry({
				date: dayBeforeYesterdayYMD,
				accessToken,
			}),
			...recentDays.map((day) => getDiaryEntry({ date: day.ymd, accessToken })),
		]);

		// 直近7日分のデータを整形
		const recentDaysWithEntry = recentDays.map((day, index) => ({
			date: day.date,
			hasEntry:
				recentDaysResults[index]?.status === "fulfilled" &&
				recentDaysResults[index].value.entry !== null,
			dayOfWeek: day.dayOfWeek,
			dayOfMonth: day.dayOfMonth,
		}));

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
			recentDays: recentDaysWithEntry,
		};
	} catch (err) {
		console.error("Failed to load diary entries:", err);
		const now = new Date();

		// エラー時も直近7日分のデータ構造を返す（古い日付から新しい日付の順）
		const recentDays = Array.from({ length: 7 }, (_, i) => {
			const date = new Date(now);
			date.setDate(date.getDate() - (6 - i)); // 6日前から今日まで
			const dayOfWeekArray = [
				"Sunday",
				"Monday",
				"Tuesday",
				"Wednesday",
				"Thursday",
				"Friday",
				"Saturday",
			] as const;
			return {
				date: `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`,
				hasEntry: false,
				dayOfWeek: dayOfWeekArray[date.getDay()],
				dayOfMonth: date.getDate(),
			};
		});

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
			recentDays,
		};
	}
};

export const actions: Actions = {
	saveToday: async ({ request, cookies }) => {
		const authResult = await ensureValidAccessToken(cookies);

		if (!authResult.isAuthenticated || !authResult.accessToken) {
			throw error(401, "Unauthorized");
		}

		const data = await request.formData();
		const content = data.get("content")?.toString();
		const dateStr = data.get("date")?.toString();
		const id = data.get("id")?.toString();
		const selectedEntitiesStr = data.get("selectedEntities")?.toString();

		if (!content || !dateStr) {
			return { error: "内容と日付は必須です" };
		}

		try {
			const [year, month, day] = dateStr.split("-").map(Number);
			const date = createYMD(year, month, day);

			// 明示的に選択されたエンティティのみを使用
			let diaryEntities: DiaryEntityInput[] = [];
			if (selectedEntitiesStr && selectedEntitiesStr !== "[]") {
				try {
					const selectedEntities = JSON.parse(selectedEntitiesStr) as {
						entityId: string;
						positions: { start: number; end: number }[];
					}[];

					const { create } = await import("@bufbuild/protobuf");
					const { DiaryEntityInputSchema } = await import(
						"$lib/grpc/diary/diary_pb"
					);
					const { PositionSchema } = await import("$lib/grpc/entity/entity_pb");

					diaryEntities = selectedEntities.map((se) => {
						const positionMessages = se.positions.map((pos) =>
							create(PositionSchema, {
								start: pos.start,
								end: pos.end,
							}),
						);

						return create(DiaryEntityInputSchema, {
							entityId: se.entityId,
							positions: positionMessages,
						});
					});
				} catch (parseErr) {
					console.error("Failed to parse selectedEntities:", parseErr);
					diaryEntities = [];
				}
			}

			if (id) {
				// 更新
				await updateDiaryEntry({
					id,
					title: "",
					content,
					date,
					diaryEntities,
					accessToken: authResult.accessToken,
				});
			} else {
				// 新規作成
				await createDiaryEntry({
					content,
					date,
					diaryEntities,
					accessToken: authResult.accessToken,
				});
			}

			return { success: true };
		} catch (err) {
			console.error("Failed to save diary entry:", err);
			return { error: "日記の保存に失敗しました" };
		}
	},
	saveYesterday: async ({ request, cookies }) => {
		const authResult = await ensureValidAccessToken(cookies);

		if (!authResult.isAuthenticated || !authResult.accessToken) {
			throw error(401, "Unauthorized");
		}

		const data = await request.formData();
		const content = data.get("content")?.toString();
		const dateStr = data.get("date")?.toString();
		const id = data.get("id")?.toString();
		const selectedEntitiesStr = data.get("selectedEntities")?.toString();

		if (!content || !dateStr) {
			return { error: "内容と日付は必須です" };
		}

		try {
			const [year, month, day] = dateStr.split("-").map(Number);
			const date = createYMD(year, month, day);

			// 明示的に選択されたエンティティのみを使用
			let diaryEntities: DiaryEntityInput[] = [];
			if (selectedEntitiesStr && selectedEntitiesStr !== "[]") {
				try {
					const selectedEntities = JSON.parse(selectedEntitiesStr) as {
						entityId: string;
						positions: { start: number; end: number }[];
					}[];

					const { create } = await import("@bufbuild/protobuf");
					const { DiaryEntityInputSchema } = await import(
						"$lib/grpc/diary/diary_pb"
					);
					const { PositionSchema } = await import("$lib/grpc/entity/entity_pb");

					diaryEntities = selectedEntities.map((se) => {
						const positionMessages = se.positions.map((pos) =>
							create(PositionSchema, {
								start: pos.start,
								end: pos.end,
							}),
						);

						return create(DiaryEntityInputSchema, {
							entityId: se.entityId,
							positions: positionMessages,
						});
					});
				} catch (parseErr) {
					console.error("Failed to parse selectedEntities:", parseErr);
					diaryEntities = [];
				}
			}

			if (id) {
				await updateDiaryEntry({
					id,
					title: "",
					content,
					date,
					diaryEntities,
					accessToken: authResult.accessToken,
				});
			} else {
				await createDiaryEntry({
					content,
					date,
					diaryEntities,
					accessToken: authResult.accessToken,
				});
			}

			return { success: true };
		} catch (err) {
			console.error("Failed to save diary entry:", err);
			return { error: "日記の保存に失敗しました" };
		}
	},
	saveDayBeforeYesterday: async ({ request, cookies }) => {
		const authResult = await ensureValidAccessToken(cookies);

		if (!authResult.isAuthenticated || !authResult.accessToken) {
			throw error(401, "Unauthorized");
		}

		const data = await request.formData();
		const content = data.get("content")?.toString();
		const dateStr = data.get("date")?.toString();
		const id = data.get("id")?.toString();
		const selectedEntitiesStr = data.get("selectedEntities")?.toString();

		if (!content || !dateStr) {
			return { error: "内容と日付は必須です" };
		}

		try {
			const [year, month, day] = dateStr.split("-").map(Number);
			const date = createYMD(year, month, day);

			// 明示的に選択されたエンティティのみを使用
			let diaryEntities: DiaryEntityInput[] = [];
			if (selectedEntitiesStr && selectedEntitiesStr !== "[]") {
				try {
					const selectedEntities = JSON.parse(selectedEntitiesStr) as {
						entityId: string;
						positions: { start: number; end: number }[];
					}[];

					const { create } = await import("@bufbuild/protobuf");
					const { DiaryEntityInputSchema } = await import(
						"$lib/grpc/diary/diary_pb"
					);
					const { PositionSchema } = await import("$lib/grpc/entity/entity_pb");

					diaryEntities = selectedEntities.map((se) => {
						const positionMessages = se.positions.map((pos) =>
							create(PositionSchema, {
								start: pos.start,
								end: pos.end,
							}),
						);

						return create(DiaryEntityInputSchema, {
							entityId: se.entityId,
							positions: positionMessages,
						});
					});
				} catch (parseErr) {
					console.error("Failed to parse selectedEntities:", parseErr);
					diaryEntities = [];
				}
			}

			if (id) {
				await updateDiaryEntry({
					id,
					title: "",
					content,
					date,
					diaryEntities,
					accessToken: authResult.accessToken,
				});
			} else {
				await createDiaryEntry({
					content,
					date,
					diaryEntities,
					accessToken: authResult.accessToken,
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
