package user

import (
	"context"
	"strings"
	"testing"

	"github.com/project-mikan/umi.mikan/backend/domain/model"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func TestUserEntry_CreateApiKey(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "api-key-create@example.com", "APIKeyCreateUser")
	svc := &UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	t.Run("正常系: APIキーが発行されキー本体が一度だけ返る", func(t *testing.T) {
		resp, err := svc.CreateApiKey(ctx, &g.CreateApiKeyRequest{Name: "Claude Desktop"})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if !strings.HasPrefix(resp.ApiKey, model.APIKeyPrefix) {
			t.Errorf("キーが %q で始まらない: %q", model.APIKeyPrefix, resp.ApiKey)
		}
		if resp.Info == nil {
			t.Fatal("Infoがnil")
		}
		if resp.Info.Name != "Claude Desktop" {
			t.Errorf("Name: 期待 %q, 実際 %q", "Claude Desktop", resp.Info.Name)
		}
		if !strings.HasPrefix(resp.ApiKey, resp.Info.KeyPrefix) {
			t.Errorf("KeyPrefix %q がキー本体の先頭と一致しない", resp.Info.KeyPrefix)
		}
		if resp.Info.LastUsedAt != 0 {
			t.Errorf("未使用キーのLastUsedAtは0を期待したが %d", resp.Info.LastUsedAt)
		}
	})

	t.Run("異常系: 名前が空の場合はエラー", func(t *testing.T) {
		_, err := svc.CreateApiKey(ctx, &g.CreateApiKeyRequest{Name: ""})
		if err == nil {
			t.Fatal("名前が空でエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: 名前が長すぎる場合はエラー", func(t *testing.T) {
		_, err := svc.CreateApiKey(ctx, &g.CreateApiKeyRequest{Name: strings.Repeat("あ", 101)})
		if err == nil {
			t.Fatal("名前が長すぎる場合にエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: 未認証の場合はエラー", func(t *testing.T) {
		_, err := svc.CreateApiKey(context.Background(), &g.CreateApiKeyRequest{Name: "key"})
		if err == nil {
			t.Fatal("未認証でエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: ユーザーIDがUUID形式でない場合はエラー", func(t *testing.T) {
		badCtx := context.WithValue(context.Background(), middleware.UserIDKey, "not-a-uuid")
		_, err := svc.CreateApiKey(badCtx, &g.CreateApiKeyRequest{Name: "key"})
		if err == nil {
			t.Fatal("不正なユーザーIDでエラーを期待したがnilが返った")
		}
	})
}

func TestUserEntry_ListApiKeys(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "api-key-list@example.com", "APIKeyListUser")
	svc := &UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	t.Run("正常系: キーがない場合は空配列を返す", func(t *testing.T) {
		resp, err := svc.ListApiKeys(ctx, &g.ListApiKeysRequest{})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(resp.ApiKeys) != 0 {
			t.Errorf("期待件数 0 に対して %d 件取得", len(resp.ApiKeys))
		}
	})

	t.Run("正常系: 発行済みキーが一覧に含まれキー本体は含まれない", func(t *testing.T) {
		created, err := svc.CreateApiKey(ctx, &g.CreateApiKeyRequest{Name: "一覧テスト用キー"})
		if err != nil {
			t.Fatalf("キー発行に失敗: %v", err)
		}

		resp, err := svc.ListApiKeys(ctx, &g.ListApiKeysRequest{})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(resp.ApiKeys) != 1 {
			t.Fatalf("期待件数 1 に対して %d 件取得", len(resp.ApiKeys))
		}
		info := resp.ApiKeys[0]
		if info.Id != created.Info.Id {
			t.Errorf("Id: 期待 %v, 実際 %v", created.Info.Id, info.Id)
		}
		if info.Name != "一覧テスト用キー" {
			t.Errorf("Name: 期待 %q, 実際 %q", "一覧テスト用キー", info.Name)
		}
		if info.KeyPrefix != created.Info.KeyPrefix {
			t.Errorf("KeyPrefix: 期待 %q, 実際 %q", created.Info.KeyPrefix, info.KeyPrefix)
		}
	})

	t.Run("正常系: 他ユーザーのキーは一覧に含まれない", func(t *testing.T) {
		otherUserID := testutil.CreateTestUser(t, db, "api-key-list-other@example.com", "APIKeyOtherUser")
		otherCtx := testutil.CreateAuthenticatedContext(otherUserID)
		resp, err := svc.ListApiKeys(otherCtx, &g.ListApiKeysRequest{})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if len(resp.ApiKeys) != 0 {
			t.Errorf("他ユーザーのキーが返らないことを期待したが %d 件取得", len(resp.ApiKeys))
		}
	})

	t.Run("異常系: 未認証の場合はエラー", func(t *testing.T) {
		_, err := svc.ListApiKeys(context.Background(), &g.ListApiKeysRequest{})
		if err == nil {
			t.Fatal("未認証でエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: ユーザーIDがUUID形式でない場合はエラー", func(t *testing.T) {
		badCtx := context.WithValue(context.Background(), middleware.UserIDKey, "not-a-uuid")
		_, err := svc.ListApiKeys(badCtx, &g.ListApiKeysRequest{})
		if err == nil {
			t.Fatal("不正なユーザーIDでエラーを期待したがnilが返った")
		}
	})
}

func TestUserEntry_DeleteApiKey(t *testing.T) {
	db := testutil.SetupTestDB(t)
	userID := testutil.CreateTestUser(t, db, "api-key-delete@example.com", "APIKeyDeleteUser")
	svc := &UserEntry{DB: db}
	ctx := testutil.CreateAuthenticatedContext(userID)

	t.Run("正常系: 自分のキーを削除できる", func(t *testing.T) {
		created, err := svc.CreateApiKey(ctx, &g.CreateApiKeyRequest{Name: "削除テスト用キー"})
		if err != nil {
			t.Fatalf("キー発行に失敗: %v", err)
		}

		resp, err := svc.DeleteApiKey(ctx, &g.DeleteApiKeyRequest{Id: created.Info.Id})
		if err != nil {
			t.Fatalf("予期しないエラー: %v", err)
		}
		if !resp.Success {
			t.Error("Success=trueを期待した")
		}

		// 削除後は一覧に含まれない
		listResp, err := svc.ListApiKeys(ctx, &g.ListApiKeysRequest{})
		if err != nil {
			t.Fatalf("一覧取得に失敗: %v", err)
		}
		for _, info := range listResp.ApiKeys {
			if info.Id == created.Info.Id {
				t.Error("削除したキーが一覧に残っている")
			}
		}
	})

	t.Run("異常系: 存在しないキーIDはNotFound", func(t *testing.T) {
		_, err := svc.DeleteApiKey(ctx, &g.DeleteApiKeyRequest{Id: "00000000-0000-0000-0000-000000000001"})
		if err == nil {
			t.Fatal("存在しないキーでエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: 他ユーザーのキーは削除できない（NotFound）", func(t *testing.T) {
		created, err := svc.CreateApiKey(ctx, &g.CreateApiKeyRequest{Name: "他ユーザー削除防止テスト"})
		if err != nil {
			t.Fatalf("キー発行に失敗: %v", err)
		}

		otherUserID := testutil.CreateTestUser(t, db, "api-key-delete-other@example.com", "APIKeyDelOther")
		otherCtx := testutil.CreateAuthenticatedContext(otherUserID)
		_, err = svc.DeleteApiKey(otherCtx, &g.DeleteApiKeyRequest{Id: created.Info.Id})
		if err == nil {
			t.Fatal("他ユーザーのキー削除でエラーを期待したがnilが返った")
		}

		// キーは残っている
		listResp, err := svc.ListApiKeys(ctx, &g.ListApiKeysRequest{})
		if err != nil {
			t.Fatalf("一覧取得に失敗: %v", err)
		}
		found := false
		for _, info := range listResp.ApiKeys {
			if info.Id == created.Info.Id {
				found = true
			}
		}
		if !found {
			t.Error("他ユーザーの削除試行でキーが消えた")
		}
	})

	t.Run("異常系: キーIDがUUID形式でない場合はエラー", func(t *testing.T) {
		_, err := svc.DeleteApiKey(ctx, &g.DeleteApiKeyRequest{Id: "not-a-uuid"})
		if err == nil {
			t.Fatal("不正なキーIDでエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: 未認証の場合はエラー", func(t *testing.T) {
		_, err := svc.DeleteApiKey(context.Background(), &g.DeleteApiKeyRequest{Id: "00000000-0000-0000-0000-000000000001"})
		if err == nil {
			t.Fatal("未認証でエラーを期待したがnilが返った")
		}
	})

	t.Run("異常系: ユーザーIDがUUID形式でない場合はエラー", func(t *testing.T) {
		badCtx := context.WithValue(context.Background(), middleware.UserIDKey, "not-a-uuid")
		_, err := svc.DeleteApiKey(badCtx, &g.DeleteApiKeyRequest{Id: "00000000-0000-0000-0000-000000000001"})
		if err == nil {
			t.Fatal("不正なユーザーIDでエラーを期待したがnilが返った")
		}
	})
}
