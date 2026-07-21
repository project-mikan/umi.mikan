import { beforeEach, describe, expect, it, vi } from "vitest";
import { issueMcpOAuthConsent } from "./mcp-oauth-api";

const mockFetch = vi.fn();
globalThis.fetch = mockFetch;

describe("mcp-oauth-api", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("issueMcpOAuthConsent", () => {
    it("正常系: バックエンドが200を返すとredirect_urlを含む結果を返す", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          redirect_url: "https://claude.ai/callback?code=abc&state=xyz",
        }),
      });

      const result = await issueMcpOAuthConsent({
        accessToken: "token-1",
        clientId: "client-1",
        redirectUri: "https://claude.ai/callback",
        codeChallenge: "challenge",
        codeChallengeMethod: "S256",
        state: "xyz",
      });

      expect(result.redirectUrl).toBe(
        "https://claude.ai/callback?code=abc&state=xyz",
      );
      expect(mockFetch).toHaveBeenCalledWith(
        "http://backend:8014/oauth/consent",
        expect.objectContaining({
          method: "POST",
          headers: expect.objectContaining({
            Authorization: "Bearer token-1",
          }),
        }),
      );
    });

    it("異常系: バックエンドがエラーレスポンスを返すと例外を投げるので、呼び出し元（+page.server.tsのconsentアクション）がfail()でユーザーにエラー表示できる", async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        text: async () => '{"error":"invalid_request"}',
      });

      await expect(
        issueMcpOAuthConsent({
          accessToken: "token-1",
          clientId: "client-1",
          redirectUri: "https://claude.ai/callback",
          codeChallenge: "challenge",
          codeChallengeMethod: "S256",
          state: "",
        }),
      ).rejects.toThrow(/MCP OAuth consent failed/);
    });
  });
});
