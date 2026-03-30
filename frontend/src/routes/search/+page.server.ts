import { fail, redirect } from "@sveltejs/kit";
import {
	searchDiaryEntries,
	searchDiaryEntriesSemantic,
} from "$lib/server/diary-api.js";
import { getUserInfo } from "$lib/server/auth-api";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import type { Actions, PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ url, cookies }) => {
	const keyword = url.searchParams.get("q") || "";
	const mode = url.searchParams.get("mode") || "keyword";
	const authResult = await ensureValidAccessToken(cookies);

	if (!authResult.isAuthenticated || !authResult.accessToken) {
		throw redirect(302, "/login");
	}

	// ユーザーの意味的検索設定を取得
	let semanticSearchEnabled = false;
	try {
		const userInfo = await getUserInfo({ accessToken: authResult.accessToken });
		const geminiKey = userInfo.llmKeys?.find((k) => k.llmProvider === 1);
		semanticSearchEnabled = geminiKey?.semanticSearchEnabled ?? false;
	} catch {
		// ユーザー情報取得失敗時はデフォルト無効
	}

	if (!keyword) {
		return {
			searchResults: null,
			semanticResults: null,
			keyword: "",
			expandedKeywords: [] as string[],
			mode,
			semanticSearchEnabled,
		};
	}

	if (mode === "semantic") {
		try {
			const searchResponse = await searchDiaryEntriesSemantic({
				query: keyword,
				limit: 10,
				accessToken: authResult.accessToken,
			});

			return {
				searchResults: null,
				semanticResults: searchResponse,
				keyword,
				mode,
				semanticSearchEnabled,
			};
		} catch (err) {
			console.error("Semantic search error:", err);
			return {
				searchResults: null,
				semanticResults: null,
				keyword,
				mode,
				semanticSearchEnabled,
				error: "semanticSearchFailed",
			};
		}
	}

	try {
		const searchResponse = await searchDiaryEntries({
			keyword: keyword,
			accessToken: authResult.accessToken,
		});

		return {
			searchResults: searchResponse,
			semanticResults: null,
			keyword,
			expandedKeywords: searchResponse.expandedKeywords ?? [],
			mode,
			semanticSearchEnabled,
		};
	} catch (err) {
		console.error("Search error:", err);
		return {
			searchResults: null,
			semanticResults: null,
			keyword,
			expandedKeywords: [] as string[],
			mode,
			semanticSearchEnabled,
			error: "searchFailed",
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
