import { error, redirect } from "@sveltejs/kit";
import {
	createDiaryEntry,
	createYMD,
	deleteDiaryEntry,
	getDiaryEntry,
	updateDiaryEntry,
} from "$lib/server/diary-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { getPastSameDates } from "$lib/utils/date-utils";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ params, cookies }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw redirect(302, "/login");
	}

	// params.id should be in format YYYY-MM-DD
	const dateMatch = params.id.match(/^(\d{4})-(\d{2})-(\d{2})$/);
	if (!dateMatch) {
		throw error(400, "Invalid date format");
	}

	const [, year, month, day] = dateMatch;
	const date = createYMD(
		Number.parseInt(year, 10),
		Number.parseInt(month, 10),
		Number.parseInt(day, 10),
	);

	// 過去の同日の日付を計算
	const pastDates = getPastSameDates({
		year: Number.parseInt(year, 10),
		month: Number.parseInt(month, 10),
		day: Number.parseInt(day, 10),
	});

	try {
		// メインの日記を取得
		const response = await getDiaryEntry({
			date,
			accessToken: authResult.accessToken,
		});

		// 過去の日記を並行して取得
		const pastEntriesPromises = [
			getDiaryEntry({
				date: createYMD(
					pastDates.oneWeekAgo.year,
					pastDates.oneWeekAgo.month,
					pastDates.oneWeekAgo.day,
				),
				accessToken: authResult.accessToken,
			}).catch(() => ({ entry: null })),
			getDiaryEntry({
				date: createYMD(
					pastDates.oneMonthAgo.year,
					pastDates.oneMonthAgo.month,
					pastDates.oneMonthAgo.day,
				),
				accessToken: authResult.accessToken,
			}).catch(() => ({ entry: null })),
			getDiaryEntry({
				date: createYMD(
					pastDates.oneYearAgo.year,
					pastDates.oneYearAgo.month,
					pastDates.oneYearAgo.day,
				),
				accessToken: authResult.accessToken,
			}).catch(() => ({ entry: null })),
			getDiaryEntry({
				date: createYMD(
					pastDates.twoYearsAgo.year,
					pastDates.twoYearsAgo.month,
					pastDates.twoYearsAgo.day,
				),
				accessToken: authResult.accessToken,
			}).catch(() => ({ entry: null })),
		];

		const pastEntriesResults = await Promise.all(pastEntriesPromises);

		// Return the entry if it exists, or null if it doesn't (allowing creation)
		return {
			entry: response.entry || null,
			date,
			pastEntries: {
				oneWeekAgo: {
					date: pastDates.oneWeekAgo,
					entry: pastEntriesResults[0].entry || null,
				},
				oneMonthAgo: {
					date: pastDates.oneMonthAgo,
					entry: pastEntriesResults[1].entry || null,
				},
				oneYearAgo: {
					date: pastDates.oneYearAgo,
					entry: pastEntriesResults[2].entry || null,
				},
				twoYearsAgo: {
					date: pastDates.twoYearsAgo,
					entry: pastEntriesResults[3].entry || null,
				},
			},
		};
	} catch (err) {
		if (err instanceof Response) {
			throw err;
		}

		// Handle gRPC NOT_FOUND error (code 2) - this is normal when no diary entry exists
		if (err && typeof err === "object" && "code" in err && err.code === 2) {
			// 過去の日記も取得（エラーでもnullを返す）
			const pastEntriesPromises = [
				getDiaryEntry({
					date: createYMD(
						pastDates.oneWeekAgo.year,
						pastDates.oneWeekAgo.month,
						pastDates.oneWeekAgo.day,
					),
					accessToken: authResult.accessToken,
				}).catch(() => ({ entry: null })),
				getDiaryEntry({
					date: createYMD(
						pastDates.oneMonthAgo.year,
						pastDates.oneMonthAgo.month,
						pastDates.oneMonthAgo.day,
					),
					accessToken: authResult.accessToken,
				}).catch(() => ({ entry: null })),
				getDiaryEntry({
					date: createYMD(
						pastDates.oneYearAgo.year,
						pastDates.oneYearAgo.month,
						pastDates.oneYearAgo.day,
					),
					accessToken: authResult.accessToken,
				}).catch(() => ({ entry: null })),
				getDiaryEntry({
					date: createYMD(
						pastDates.twoYearsAgo.year,
						pastDates.twoYearsAgo.month,
						pastDates.twoYearsAgo.day,
					),
					accessToken: authResult.accessToken,
				}).catch(() => ({ entry: null })),
			];

			const pastEntriesResults = await Promise.all(pastEntriesPromises);

			return {
				entry: null,
				date,
				pastEntries: {
					oneWeekAgo: {
						date: pastDates.oneWeekAgo,
						entry: pastEntriesResults[0].entry || null,
					},
					oneMonthAgo: {
						date: pastDates.oneMonthAgo,
						entry: pastEntriesResults[1].entry || null,
					},
					oneYearAgo: {
						date: pastDates.oneYearAgo,
						entry: pastEntriesResults[2].entry || null,
					},
					twoYearsAgo: {
						date: pastDates.twoYearsAgo,
						entry: pastEntriesResults[3].entry || null,
					},
				},
			};
		}

		console.error("Failed to load diary entry:", err);
		throw error(500, "Failed to load diary entry");
	}
};

export const actions: Actions = {
	save: async ({ request, cookies }) => {
		const authResult = await ensureValidAccessToken(cookies);

		if (!authResult.isAuthenticated || !authResult.accessToken) {
			throw error(401, "Unauthorized");
		}

		const data = await request.formData();
		const content = data.get("content") as string;
		const id = data.get("id") as string;
		const dateStr = data.get("date") as string;

		if (!content || !dateStr) {
			return {
				error: "Content and date are required",
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

			if (id) {
				// Update existing entry
				await updateDiaryEntry({
					id,
					title: "",
					content,
					date: ymd,
					accessToken: authResult.accessToken,
				});
			} else {
				// Create new entry
				await createDiaryEntry({
					content,
					date: ymd,
					accessToken: authResult.accessToken,
				});
			}
		} catch (err) {
			if (err instanceof Response) {
				throw err;
			}
			console.error("Failed to save diary entry:", err);
			return {
				error: "Failed to save diary entry",
			};
		}

		return {
			success: true,
		};
	},

	delete: async ({ params, cookies }) => {
		const authResult = await ensureValidAccessToken(cookies);

		if (!authResult.isAuthenticated || !authResult.accessToken) {
			throw error(401, "Unauthorized");
		}

		try {
			// First, get the current entry to get the ID
			const dateMatch = params.id.match(/^(\d{4})-(\d{2})-(\d{2})$/);
			if (!dateMatch) {
				throw error(400, "Invalid date format");
			}

			const [, year, month, day] = dateMatch;
			let currentResponse: Awaited<ReturnType<typeof getDiaryEntry>>;
			try {
				currentResponse = await getDiaryEntry({
					date: createYMD(
						Number.parseInt(year, 10),
						Number.parseInt(month, 10),
						Number.parseInt(day, 10),
					),
					accessToken: authResult.accessToken,
				});
			} catch (getDiaryErr) {
				// Handle gRPC NOT_FOUND error (code 2) - diary entry doesn't exist
				if (
					getDiaryErr &&
					typeof getDiaryErr === "object" &&
					"code" in getDiaryErr &&
					getDiaryErr.code === 2
				) {
					return {
						error: "Diary entry not found",
					};
				}
				throw getDiaryErr;
			}

			if (!currentResponse.entry) {
				return {
					error: "Diary entry not found",
				};
			}

			await deleteDiaryEntry({
				id: currentResponse.entry.id,
				accessToken: authResult.accessToken,
			});
		} catch (err) {
			if (err instanceof Response) {
				throw err;
			}
			console.error("Failed to delete diary entry:", err);
			return {
				error: "Failed to delete diary entry",
			};
		}

		throw redirect(303, "/");
	},
};
