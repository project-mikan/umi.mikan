import { getUserInfo } from "$lib/server/auth-api";
import { isTokenExpiringSoon } from "$lib/utils/token-utils";
import { setCSRFToken, getCSRFToken } from "$lib/server/csrf";
import type { LayoutServerLoad } from "./$types";

export const load: LayoutServerLoad = async ({ cookies, url }) => {
  // ここではリフレッシュを行わない。
  // invalidateAll() で +layout と +page が並行実行されるとき、
  // 両方が refreshAccessToken() を呼ぶと同じ refreshToken を二重消費して競合が起きる。
  // リフレッシュは各 +page.server.ts の ensureValidAccessToken に任せる。
  const accessToken = cookies.get("accessToken");
  const refreshToken = cookies.get("refreshToken");

  // accessToken がなくても refreshToken があれば認証済みとみなす。
  // 実際のトークン更新は +page.server.ts の ensureValidAccessToken が担う。
  const isAuthenticated = !!(accessToken || refreshToken);

  // CSRFトークンを設定・取得
  let csrfToken = getCSRFToken(cookies);
  if (!csrfToken) {
    csrfToken = setCSRFToken(cookies);
  }

  // ユーザー情報を取得（accessToken が有効な場合のみ）
  // accessToken が期限切れ・未取得のケースは null を返してレイアウトだけ壊さないようにする
  let userName: string | null = null;
  let autoLatestTrendEnabled = false;
  const hasValidAccessToken =
    !!accessToken && !isTokenExpiringSoon(accessToken);
  if (hasValidAccessToken) {
    try {
      const userInfo = await getUserInfo({
        accessToken: accessToken as string,
      });
      userName = userInfo.name;
      // LLMキー情報から autoLatestTrendEnabled を取得（Gemini provider=1のみ）
      const geminiKey = userInfo.llmKeys?.find((key) => key.llmProvider === 1);
      autoLatestTrendEnabled = geminiKey?.autoLatestTrendEnabled || false;
    } catch (error) {
      // バックエンド一時エラーでログアウトさせないようエラーは吸収する
      console.error("Failed to get user info:", error);
    }
  }

  return {
    isAuthenticated,
    path: url.pathname,
    csrfToken,
    userName,
    autoLatestTrendEnabled,
  };
};
