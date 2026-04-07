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

  // getUserInfo と検索クエリを並列で発火（キーワードがある場合）
  const userInfoPromise = getUserInfo({
    accessToken: authResult.accessToken,
  }).catch(() => null);

  const keywordPromise =
    keyword && mode !== "semantic"
      ? searchDiaryEntries({
          keyword,
          accessToken: authResult.accessToken,
        }).catch(() => null)
      : Promise.resolve(null);

  const semanticPromise =
    keyword && mode === "semantic"
      ? searchDiaryEntriesSemantic({
          query: keyword,
          limit: 10,
          accessToken: authResult.accessToken,
        }).catch(() => null)
      : Promise.resolve(null);

  const [userInfo, keywordResponse, semanticResponse] = await Promise.all([
    userInfoPromise,
    keywordPromise,
    semanticPromise,
  ]);

  const geminiKey = userInfo?.llmKeys?.find((k) => k.llmProvider === 1);
  const semanticSearchEnabled = geminiKey?.semanticSearchEnabled ?? false;

  return {
    searchResults: keywordResponse,
    semanticResults: semanticResponse,
    keyword,
    expandedKeywords: keywordResponse?.expandedKeywords ?? [],
    mode,
    semanticSearchEnabled,
  };
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
