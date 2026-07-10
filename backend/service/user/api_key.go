package user

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/project-mikan/umi.mikan/backend/domain/model"
	"github.com/project-mikan/umi.mikan/backend/infrastructure/database"
	g "github.com/project-mikan/umi.mikan/backend/infrastructure/grpc"
	"github.com/project-mikan/umi.mikan/backend/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// apiKeyNameMaxLength はAPIキー名の最大文字数（user_api_keys.nameのVARCHAR(100)に合わせる）
const apiKeyNameMaxLength = 100

// toApiKeyInfo はDB行をレスポンス用のApiKeyInfoに変換する
func toApiKeyInfo(key *database.UserAPIKey) *g.ApiKeyInfo {
	lastUsedAt := int64(0)
	if key.LastUsedAt.Valid {
		lastUsedAt = key.LastUsedAt.Int64
	}
	return &g.ApiKeyInfo{
		Id:         key.ID.String(),
		Name:       key.Name,
		KeyPrefix:  key.KeyPrefix,
		LastUsedAt: lastUsedAt,
		CreatedAt:  key.CreatedAt,
	}
}

// CreateApiKey はMCPサーバーなど外部クライアント向けのAPIキーを発行する。
// キー本体はこのレスポンスで一度だけ返し、DBにはSHA-256ハッシュのみ保存する。
func (s *UserEntry) CreateApiKey(ctx context.Context, req *g.CreateApiKeyRequest) (*g.CreateApiKeyResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalidUserId")
	}

	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "nameRequired")
	}
	if len([]rune(req.GetName())) > apiKeyNameMaxLength {
		return nil, status.Error(codes.InvalidArgument, "nameTooLong")
	}

	generated, err := model.GenerateAPIKey()
	if err != nil {
		return nil, status.Error(codes.Internal, "createFailed")
	}

	currentTime := time.Now().Unix()
	key := &database.UserAPIKey{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      req.GetName(),
		KeyHash:   generated.Hash,
		KeyPrefix: generated.DisplayPrefix,
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
	}
	if err := key.Insert(ctx, s.DB); err != nil {
		return nil, status.Error(codes.Internal, "createFailed")
	}

	return &g.CreateApiKeyResponse{
		ApiKey: generated.Key,
		Info:   toApiKeyInfo(key),
	}, nil
}

// ListApiKeys は発行済みAPIキーの一覧を作成日時の降順で返す（キー本体は含まれない）
func (s *UserEntry) ListApiKeys(ctx context.Context, _ *g.ListApiKeysRequest) (*g.ListApiKeysResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalidUserId")
	}

	keys, err := database.UserAPIKeysByUserID(ctx, s.DB, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "listFailed")
	}

	// 作成日時の降順（新しいキーが先頭）
	sort.Slice(keys, func(i, j int) bool { return keys[i].CreatedAt > keys[j].CreatedAt })

	infos := make([]*g.ApiKeyInfo, 0, len(keys))
	for _, key := range keys {
		infos = append(infos, toApiKeyInfo(key))
	}

	return &g.ListApiKeysResponse{ApiKeys: infos}, nil
}

// DeleteApiKey は指定されたAPIキーを失効させる。
// 他ユーザーのキーは存在しないものとして扱う（NotFound）。
func (s *UserEntry) DeleteApiKey(ctx context.Context, req *g.DeleteApiKeyRequest) (*g.DeleteApiKeyResponse, error) {
	userIDStr, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalidUserId")
	}

	keyID, err := uuid.Parse(req.GetId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalidKeyId")
	}

	key, err := database.UserAPIKeyByID(ctx, s.DB, keyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "keyNotFound")
		}
		return nil, status.Error(codes.Internal, "deleteFailed")
	}

	// 他ユーザーのキーは存在を悟らせないためNotFoundを返す
	if key.UserID != userID {
		return nil, status.Error(codes.NotFound, "keyNotFound")
	}

	if err := key.Delete(ctx, s.DB); err != nil {
		return nil, status.Error(codes.Internal, "deleteFailed")
	}

	return &g.DeleteApiKeyResponse{
		Success: true,
		Message: "apiKeyDeleted",
	}, nil
}
