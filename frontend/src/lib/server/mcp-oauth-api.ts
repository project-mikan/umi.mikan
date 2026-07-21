// umi.mikan MCPサーバー（backend/infrastructure/mcpserver）が提供する
// OAuth 2.0 Authorization Code + PKCEフローのうち、フロントエンドが担当する
// 「ログイン済みユーザーの同意」部分を仲介するサーバーサイド関数。
// 実際のauthorization code発行はバックエンドの POST /oauth/consent が行う
// （adr/0016参照）。gRPCではなく素のHTTP/JSONエンドポイントのため、
// 他のgRPCクライアント（auth-api.ts）とは別ファイルに分離している。

// バックエンドのMCPサーバーはDockerネットワーク内では :8014 で待ち受けている
// （gRPCの "http://backend:8080" と同じ命名規則）。
const MCP_SERVER_BASE_URL = "http://backend:8014";

export interface ConsentParams {
  accessToken: string;
  clientId: string;
  redirectUri: string;
  codeChallenge: string;
  codeChallengeMethod: string;
  state: string;
}

export interface ConsentResult {
  redirectUrl: string;
}

/**
 * POST /oauth/consent がエラーレスポンスを返した際に投げる例外。
 * バックエンドのHTTPステータスコードを保持することで、呼び出し元
 * （+page.server.tsのconsentアクション）が401（トークン失効など再ログインで
 * 解決する可能性がある場合）とそれ以外のエラーを区別してユーザーに案内できるようにする。
 */
export class McpOAuthConsentError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = "McpOAuthConsentError";
    this.status = status;
  }
}

/**
 * ユーザーがMCP接続の同意操作をした後に呼び出す。
 * バックエンドの POST /oauth/consent にJWTアクセストークンを添えてリクエストし、
 * authorization codeを含んだMCPクライアント側のredirect_uriを取得する。
 */
export async function issueMcpOAuthConsent(
  params: ConsentParams,
): Promise<ConsentResult> {
  const response = await fetch(`${MCP_SERVER_BASE_URL}/oauth/consent`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${params.accessToken}`,
    },
    body: JSON.stringify({
      client_id: params.clientId,
      redirect_uri: params.redirectUri,
      code_challenge: params.codeChallenge,
      code_challenge_method: params.codeChallengeMethod,
      state: params.state,
    }),
  });

  if (!response.ok) {
    const errorBody = await response.text();
    throw new McpOAuthConsentError(
      response.status,
      `MCP OAuth consent failed: ${response.status} ${errorBody}`,
    );
  }

  const body = (await response.json()) as { redirect_url: string };
  return { redirectUrl: body.redirect_url };
}
