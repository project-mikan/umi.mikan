import { error, redirect } from "@sveltejs/kit";
import { createYM, getDiaryEntriesByMonth } from "$lib/server/diary-api";
import { getUserInfo } from "$lib/server/auth-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ cookies, params }) => {
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw redirect(302, "/login");
	}

	const year = Number.parseInt(params.year, 10);
	const month = Number.parseInt(params.month, 10);

	if (Number.isNaN(year) || Number.isNaN(month) || month < 1 || month > 12) {
		throw error(400, "Invalid year or month");
	}

	try {
		const [entries, userInfo] = await Promise.all([
			getDiaryEntriesByMonth({
				month: createYM(year, month),
				accessToken: authResult.accessToken,
			}),
			getUserInfo({ accessToken: authResult.accessToken }),
		]);

		// Convert BigInt to Number for JSON serialization
		const serializedEntries = {
			...entries,
			entries: entries.entries.map((entry) => ({
				...entry,
				createdAt: Number(entry.createdAt),
				updatedAt: Number(entry.updatedAt),
				diaryEntities: entry.diaryEntities || [],
			})),
		};

		return {
			entries: serializedEntries,
			year,
			month,
			user: {
				name: userInfo.name,
				email: userInfo.email,
				llmKeys: userInfo.llmKeys || [],
			},
		};
	} catch (err) {
		console.error("Failed to load diary entries:", err);
		return {
			entries: { entries: [] }, // Empty GetDiaryEntriesByMonthResponse structure
			year,
			month,
			user: {
				name: "",
				email: "",
				llmKeys: [],
			},
		};
	}
};
