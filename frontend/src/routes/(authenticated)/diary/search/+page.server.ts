import { searchDiaryEntries } from "$lib/server/diary-api.js";
import { error, fail } from "@sveltejs/kit";
import type { Actions, PageServerLoad } from "./$types.ts";

export const load: PageServerLoad = async ({ url, cookies }) => {
	const keyword = url.searchParams.get("q") || "";
	const accessToken = cookies.get("accessToken");

	if (!accessToken) {
		throw error(401, "Unauthorized");
	}

	if (!keyword) {
		return {
			searchResults: null,
			keyword: "",
		};
	}

	try {
		const searchResponse = await searchDiaryEntries({
			keyword: keyword,
			accessToken: accessToken,
		});

		return {
			searchResults: searchResponse,
			keyword: keyword,
		};
	} catch (_err) {
		// Log error for debugging but don't expose details to client
		return {
			searchResults: null,
			keyword: keyword,
			error: "Failed to search diary entries",
		};
	}
};

export const actions: Actions = {
	search: async ({ request, cookies }) => {
		const accessToken = cookies.get("accessToken");
		if (!accessToken) {
			return fail(401, { error: "Unauthorized" });
		}

		const data = await request.formData();
		const keyword = data.get("keyword")?.toString();

		if (!keyword) {
			return fail(400, { error: "Keyword is required" });
		}

		try {
			const searchResponse = await searchDiaryEntries({
				keyword: keyword,
				accessToken: accessToken,
			});

			return {
				success: true,
				searchResults: searchResponse,
				keyword: keyword,
			};
		} catch (_err) {
			// Log error for debugging but don't expose details to client
			return fail(500, { error: "Failed to search diary entries" });
		}
	},
};
