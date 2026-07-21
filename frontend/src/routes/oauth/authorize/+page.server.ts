import { fail } from "@sveltejs/kit";
import { ensureValidAccessToken } from "$lib/server/auth-middleware";
import { validateCSRFToken } from "$lib/server/csrf";
import { issueMcpOAuthConsent } from "$lib/server/mcp-oauth-api";
import type { Actions, PageServerLoad } from "./$types";

// MCPサーバー（backend/infrastructure/mcpserver）の /oauth/authorize が
// このページにリダイレクトしてくる際に付与するクエリパラメータをそのまま受け取り、
// ログイン状態に応じてログイン誘導 or 同意画面を出し分ける。
// パラメータの検証（redirect_uriのスキームなど）はバックエンド側で既に行っているため、
// ここでは必須パラメータの有無のみを確認する。
export const load: PageServerLoad = async ({ url, cookies }) => {
  const authResult = await ensureValidAccessToken(cookies);

  const clientId = url.searchParams.get("client_id") ?? "";
  const redirectUri = url.searchParams.get("redirect_uri") ?? "";
  const codeChallenge = url.searchParams.get("code_challenge") ?? "";
  const codeChallengeMethod =
    url.searchParams.get("code_challenge_method") ?? "";
  const state = url.searchParams.get("state") ?? "";

  if (!clientId || !redirectUri || !codeChallenge) {
    return {
      isAuthenticated: authResult.isAuthenticated,
      invalidRequest: true,
    };
  }

  return {
    isAuthenticated: authResult.isAuthenticated,
    invalidRequest: false,
    clientId,
    redirectUri,
    codeChallenge,
    codeChallengeMethod,
    state,
  };
};

export const actions: Actions = {
  // ユーザーが同意ボタンを押した際に呼ばれる。バックエンドにauthorization code発行を
  // リクエストし、MCPクライアント（Claude.aiなど）のredirect_uriへ遷移させるURLを
  // フロントエンド（+page.svelte）に返す。
  consent: async ({ request, cookies }) => {
    const authResult = await ensureValidAccessToken(cookies);
    if (!authResult.isAuthenticated || !authResult.accessToken) {
      return fail(401, { error: "unauthorized" });
    }

    const data = await request.formData();
    const csrfToken = data.get("csrfToken") as string;
    if (!validateCSRFToken(cookies, csrfToken)) {
      return fail(403, { error: "invalidCsrfToken" });
    }

    const clientId = data.get("client_id") as string;
    const redirectUri = data.get("redirect_uri") as string;
    const codeChallenge = data.get("code_challenge") as string;
    const codeChallengeMethod = data.get("code_challenge_method") as string;
    const state = (data.get("state") as string) ?? "";

    if (!clientId || !redirectUri || !codeChallenge) {
      return fail(400, { error: "invalidRequest" });
    }

    try {
      const result = await issueMcpOAuthConsent({
        accessToken: authResult.accessToken,
        clientId,
        redirectUri,
        codeChallenge,
        codeChallengeMethod,
        state,
      });
      return { success: true, redirectUrl: result.redirectUrl };
    } catch (error) {
      console.error("Failed to issue MCP OAuth consent:", error);
      return fail(500, { error: "consentFailed" });
    }
  },
};
