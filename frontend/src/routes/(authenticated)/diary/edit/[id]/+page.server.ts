import {
	createYMD,
	deleteDiaryEntry,
	getDiaryEntry,
	updateDiaryEntry,
} from "$lib/server/diary-api";
import { error, redirect } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types.ts";

const DATE_REGEX = /^(\d{4})-(\d{2})-(\d{2})$/;

export const load: PageServerLoad = async ({ params, cookies }) => {
	const accessToken = cookies.get("accessToken");

	if (!accessToken) {
		throw error(401, "Unauthorized");
	}

	try {
		// params.id should be in format YYYY-MM-DD
		const dateMatch = params.id.match(DATE_REGEX);
		if (!dateMatch) {
			throw error(400, "Invalid date format");
		}

		const [, year, month, day] = dateMatch;
		const response = await getDiaryEntry({
			date: createYMD(
				Number.parseInt(year, 10),
				Number.parseInt(month, 10),
				Number.parseInt(day, 10),
			),
			accessToken,
		});

		if (!response.entry) {
			throw error(404, "Diary entry not found");
		}

		return {
			entry: response.entry,
		};
	} catch (err) {
		if (err instanceof Response) {
			throw err;
		}
		// Log error for debugging but don't expose details to client
		throw error(500, "Failed to load diary entry");
	}
};

export const actions: Actions = {
	update: async ({ request, params, cookies }) => {
		const accessToken = cookies.get("accessToken");

		if (!accessToken) {
			throw error(401, "Unauthorized");
		}

		const data = await request.formData();
		const content = data.get("content") as string;
		const title = (data.get("title") as string) || "";
		const dateStr = data.get("date") as string;

		if (!(content && dateStr)) {
			return {
				error: "コンテンツと日付は必須です",
			};
		}

		try {
			// First, get the current entry to get the ID
			const dateMatch = params.id.match(DATE_REGEX);
			if (!dateMatch) {
				throw error(400, "Invalid date format");
			}

			const [, year, month, day] = dateMatch;
			const currentResponse = await getDiaryEntry({
				date: createYMD(
					Number.parseInt(year, 10),
					Number.parseInt(month, 10),
					Number.parseInt(day, 10),
				),
				accessToken,
			});

			if (!currentResponse.entry) {
				throw error(404, "Diary entry not found");
			}

			const date = new Date(dateStr);
			await updateDiaryEntry({
				id: currentResponse.entry.id,
				title,
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
			return {
				error: "日記の更新に失敗しました",
			};
		}

		redirect(303, "/diary");
	},

	delete: async ({ params, cookies }) => {
		const accessToken = cookies.get("accessToken");

		if (!accessToken) {
			throw error(401, "Unauthorized");
		}

		try {
			// First, get the current entry to get the ID
			const dateMatch = params.id.match(DATE_REGEX);
			if (!dateMatch) {
				throw error(400, "Invalid date format");
			}

			const [, year, month, day] = dateMatch;
			const currentResponse = await getDiaryEntry({
				date: createYMD(
					Number.parseInt(year, 10),
					Number.parseInt(month, 10),
					Number.parseInt(day, 10),
				),
				accessToken,
			});

			if (!currentResponse.entry) {
				throw error(404, "Diary entry not found");
			}

			await deleteDiaryEntry({
				id: currentResponse.entry.id,
				accessToken,
			});
		} catch (err) {
			if (err instanceof Response) {
				throw err;
			}
			return {
				error: "日記の削除に失敗しました",
			};
		}

		redirect(303, "/diary");
	},
};
