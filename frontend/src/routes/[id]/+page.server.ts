import {
	createDiaryEntry,
	createYMD,
	getDiaryEntry,
	updateDiaryEntry,
} from "$lib/server/diary-api";
import { error, redirect } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ params, cookies }) => {
	const accessToken = cookies.get("accessToken");

	if (!accessToken) {
		throw redirect(302, "/login");
	}

	try {
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

		const response = await getDiaryEntry({
			date,
			accessToken,
		});

		// Return the entry if it exists, or null if it doesn't (allowing creation)
		return {
			entry: response.entry || null,
			date,
		};
	} catch (err) {
		if (err instanceof Response) {
			throw err;
		}
		console.error("Failed to load diary entry:", err);
		throw error(500, "Failed to load diary entry");
	}
};

export const actions: Actions = {
	save: async ({ request, params, cookies }) => {
		const accessToken = cookies.get("accessToken");

		if (!accessToken) {
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
			const date = new Date(dateStr);
			const ymd = createYMD(
				date.getFullYear(),
				date.getMonth() + 1,
				date.getDate(),
			);

			if (id) {
				// Update existing entry
				await updateDiaryEntry({
					id,
					title: "",
					content,
					date: ymd,
					accessToken,
				});
			} else {
				// Create new entry
				await createDiaryEntry({
					content,
					date: ymd,
					accessToken,
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
};
