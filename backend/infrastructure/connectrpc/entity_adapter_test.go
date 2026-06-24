package connect

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/service/entity"
	"github.com/project-mikan/umi.mikan/backend/testutil"
)

func setupEntityAdapter(t *testing.T) *EntityServiceAdapter {
	t.Helper()
	db := testutil.SetupTestDB(t)
	svc := &entity.EntityEntry{DB: db}
	return &EntityServiceAdapter{svc: svc}
}

func TestEntityServiceAdapter_CreateEntity_Error(t *testing.T) {
	tests := []struct {
		name         string
		request      *g.CreateEntityRequest
		expectedCode connect.Code
	}{
		{
			name: "異常系: 空のエンティティ名はInvalidArgumentエラーになる",
			request: &g.CreateEntityRequest{
				Name: "",
			},
			expectedCode: connect.CodeInvalidArgument,
		},
		{
			name: "異常系: スペースのみの名前はInvalidArgumentエラーになる",
			request: &g.CreateEntityRequest{
				Name: "   ",
			},
			expectedCode: connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupEntityAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(tt.request)
			_, err := adapter.CreateEntity(ctx, req)

			if err == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}
			connectErr, ok := err.(*connect.Error)
			if !ok {
				t.Fatalf("*connect.Error を期待したが %T が返った", err)
			}
			if connectErr.Code() != tt.expectedCode {
				t.Errorf("エラーコード: 期待 %v, 実際 %v", tt.expectedCode, connectErr.Code())
			}
		})
	}
}

func TestEntityServiceAdapter_UpdateEntity_Error(t *testing.T) {
	tests := []struct {
		name         string
		request      *g.UpdateEntityRequest
		expectedCode connect.Code
	}{
		{
			name: "異常系: 不正なUUIDはInvalidArgumentエラーになる",
			request: &g.UpdateEntityRequest{
				Id:   "not-a-valid-uuid",
				Name: "Valid Name",
			},
			expectedCode: connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupEntityAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(tt.request)
			_, err := adapter.UpdateEntity(ctx, req)

			if err == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}
			connectErr, ok := err.(*connect.Error)
			if !ok {
				t.Fatalf("*connect.Error を期待したが %T が返った", err)
			}
			if connectErr.Code() != tt.expectedCode {
				t.Errorf("エラーコード: 期待 %v, 実際 %v", tt.expectedCode, connectErr.Code())
			}
		})
	}
}

func TestEntityServiceAdapter_DeleteEntity_Error(t *testing.T) {
	tests := []struct {
		name         string
		request      *g.DeleteEntityRequest
		expectedCode connect.Code
	}{
		{
			name: "異常系: 不正なUUIDはInvalidArgumentエラーになる",
			request: &g.DeleteEntityRequest{
				Id: "not-a-valid-uuid",
			},
			expectedCode: connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupEntityAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(tt.request)
			_, err := adapter.DeleteEntity(ctx, req)

			if err == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}
			connectErr, ok := err.(*connect.Error)
			if !ok {
				t.Fatalf("*connect.Error を期待したが %T が返った", err)
			}
			if connectErr.Code() != tt.expectedCode {
				t.Errorf("エラーコード: 期待 %v, 実際 %v", tt.expectedCode, connectErr.Code())
			}
		})
	}
}

func TestEntityServiceAdapter_GetEntity_Error(t *testing.T) {
	tests := []struct {
		name         string
		request      *g.GetEntityRequest
		expectedCode connect.Code
	}{
		{
			name: "異常系: 不正なUUIDはInvalidArgumentエラーになる",
			request: &g.GetEntityRequest{
				Id: "not-a-valid-uuid",
			},
			expectedCode: connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupEntityAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(tt.request)
			_, err := adapter.GetEntity(ctx, req)

			if err == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}
			connectErr, ok := err.(*connect.Error)
			if !ok {
				t.Fatalf("*connect.Error を期待したが %T が返った", err)
			}
			if connectErr.Code() != tt.expectedCode {
				t.Errorf("エラーコード: 期待 %v, 実際 %v", tt.expectedCode, connectErr.Code())
			}
		})
	}
}

func TestEntityServiceAdapter_ListEntities(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "正常系: 存在しないユーザーのエンティティ一覧は空リストを返す",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupEntityAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.ListEntitiesRequest{})
			resp, err := adapter.ListEntities(ctx, req)

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp == nil {
				t.Fatal("レスポンスがnilだった")
			}
		})
	}
}

func TestEntityServiceAdapter_CreateEntityAlias_Error(t *testing.T) {
	tests := []struct {
		name         string
		request      *g.CreateEntityAliasRequest
		expectedCode connect.Code
	}{
		{
			name: "異常系: 不正なエンティティUUIDはInvalidArgumentエラーになる",
			request: &g.CreateEntityAliasRequest{
				EntityId: "not-a-valid-uuid",
				Alias:    "test-alias",
			},
			expectedCode: connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupEntityAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(tt.request)
			_, err := adapter.CreateEntityAlias(ctx, req)

			if err == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}
			connectErr, ok := err.(*connect.Error)
			if !ok {
				t.Fatalf("*connect.Error を期待したが %T が返った", err)
			}
			if connectErr.Code() != tt.expectedCode {
				t.Errorf("エラーコード: 期待 %v, 実際 %v", tt.expectedCode, connectErr.Code())
			}
		})
	}
}

func TestEntityServiceAdapter_DeleteEntityAlias_Error(t *testing.T) {
	tests := []struct {
		name         string
		request      *g.DeleteEntityAliasRequest
		expectedCode connect.Code
	}{
		{
			name: "異常系: 不正なUUIDはInvalidArgumentエラーになる",
			request: &g.DeleteEntityAliasRequest{
				Id: "not-a-valid-uuid",
			},
			expectedCode: connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupEntityAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(tt.request)
			_, err := adapter.DeleteEntityAlias(ctx, req)

			if err == nil {
				t.Fatal("エラーを期待したがnilが返った")
			}
			connectErr, ok := err.(*connect.Error)
			if !ok {
				t.Fatalf("*connect.Error を期待したが %T が返った", err)
			}
			if connectErr.Code() != tt.expectedCode {
				t.Errorf("エラーコード: 期待 %v, 実際 %v", tt.expectedCode, connectErr.Code())
			}
		})
	}
}

func TestEntityServiceAdapter_SearchEntities(t *testing.T) {
	tests := []struct {
		name    string
		keyword string
	}{
		{
			name:    "正常系: キーワード検索は空リストを返す（ユーザーが存在しない）",
			keyword: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := setupEntityAdapter(t)
			ctx := createAuthContext(uuid.New().String())
			req := connect.NewRequest(&g.SearchEntitiesRequest{Query: tt.keyword})
			resp, err := adapter.SearchEntities(ctx, req)

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if resp == nil {
				t.Fatal("レスポンスがnilだった")
			}
		})
	}
}
