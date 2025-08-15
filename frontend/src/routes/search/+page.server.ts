import { fail, redirect } from "@sveltejs/kit";
import { searchDiaryEntries } from "$lib/server/diary-api.js";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ url, cookies }) => {
	const keyword = url.searchParams.get("q") || "";
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw redirect(302, "/login");
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
			accessToken: authResult.accessToken,
		});

		return {
			searchResults: searchResponse,
			keyword: keyword,
		};
	} catch (err) {
		console.error("Search error:", err);
		return {
			searchResults: null,
			keyword: keyword,
			error: "Failed to search diary entries",
		};
	}
};

export const actions: Actions = {
	search: async ({ request, cookies }) => {
		const authResult = await ensureValidAccessToken(cookies);
		if (!authResult.isAuthenticated || !authResult.accessToken) {
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
				accessToken: authResult.accessToken,
			});

			return {
				success: true,
				searchResults: searchResponse,
				keyword: keyword,
			};
		} catch (err) {
			console.error("Search error:", err);
			return fail(500, { error: "Failed to search diary entries" });
		}
	},
};
