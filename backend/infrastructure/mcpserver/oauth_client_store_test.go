package mcpserver

import "testing"

func TestStoreAndIsRegisteredRedirectURI(t *testing.T) {
	t.Run("正常系: 登録したredirect_uriは一致判定でtrueになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		if err := storeClientRegistration(t.Context(), redisClient, "client-1", []string{"https://claude.ai/callback"}); err != nil {
			t.Fatalf("storeClientRegistration失敗: %v", err)
		}

		ok, err := isRegisteredRedirectURI(t.Context(), redisClient, "client-1", "https://claude.ai/callback")
		if err != nil {
			t.Fatalf("isRegisteredRedirectURI失敗: %v", err)
		}
		if !ok {
			t.Error("登録済みredirect_uriなのにokがfalseになった")
		}
	})

	t.Run("異常系: 登録時と異なるredirect_uriを指定すると、authorization code横取り攻撃を防ぐためfalseになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		if err := storeClientRegistration(t.Context(), redisClient, "client-1", []string{"https://claude.ai/callback"}); err != nil {
			t.Fatalf("storeClientRegistration失敗: %v", err)
		}

		ok, err := isRegisteredRedirectURI(t.Context(), redisClient, "client-1", "https://evil.example.com/collect")
		if err != nil {
			t.Fatalf("isRegisteredRedirectURI失敗: %v", err)
		}
		if ok {
			t.Error("未登録のredirect_uriなのにokがtrueになった")
		}
	})

	t.Run("異常系: 未登録のclient_idを指定するとfalseになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		ok, err := isRegisteredRedirectURI(t.Context(), redisClient, "nonexistent-client", "https://claude.ai/callback")
		if err != nil {
			t.Fatalf("isRegisteredRedirectURI失敗: %v", err)
		}
		if ok {
			t.Error("未登録のclient_idなのにokがtrueになった")
		}
	})

	t.Run("正常系: 複数のredirect_urisを登録した場合、そのいずれかと一致すればtrueになる", func(t *testing.T) {
		redisClient := setupTestRedisForOAuthStoreTest(t)
		if err := storeClientRegistration(t.Context(), redisClient, "client-1", []string{
			"https://claude.ai/callback",
			"https://claude.ai/api/mcp/callback",
		}); err != nil {
			t.Fatalf("storeClientRegistration失敗: %v", err)
		}

		ok, err := isRegisteredRedirectURI(t.Context(), redisClient, "client-1", "https://claude.ai/api/mcp/callback")
		if err != nil {
			t.Fatalf("isRegisteredRedirectURI失敗: %v", err)
		}
		if !ok {
			t.Error("登録済みredirect_uriなのにokがfalseになった")
		}
	})
}
