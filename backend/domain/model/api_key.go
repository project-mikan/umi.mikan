package model

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// APIKeyPrefix はAPIキーの先頭に付与する識別子。
// 認証時にJWTアクセストークンとAPIキーを区別するために使用する。
const APIKeyPrefix = "umi_"

// apiKeyRandomBytes はAPIキーの乱数部分のバイト長（hex化で64文字になる）
const apiKeyRandomBytes = 32

// apiKeyDisplayPrefixLen は一覧表示用に保存するキー先頭部分の長さ（例: umi_a1b2c3d4）
const apiKeyDisplayPrefixLen = 12

// GeneratedAPIKey は発行されたAPIキー一式
type GeneratedAPIKey struct {
	// Key はキー本体（発行時に一度だけユーザーに返す）
	Key string
	// Hash はDBに保存するSHA-256ハッシュ（hex）
	Hash string
	// DisplayPrefix は一覧表示用のキー先頭部分
	DisplayPrefix string
}

// GenerateAPIKey は暗号論的乱数から新しいAPIキーを生成する
func GenerateAPIKey() (*GeneratedAPIKey, error) {
	buf := make([]byte, apiKeyRandomBytes)
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("failed to generate api key: %w", err)
	}
	key := APIKeyPrefix + hex.EncodeToString(buf)
	return &GeneratedAPIKey{
		Key:           key,
		Hash:          HashAPIKey(key),
		DisplayPrefix: key[:apiKeyDisplayPrefixLen],
	}, nil
}

// HashAPIKey はAPIキーのSHA-256ハッシュ（hex）を返す。
// キー本体はDBに保存せず、このハッシュで照合する。
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// IsAPIKey はトークンがAPIキー形式（umi_プレフィックス）かどうかを判定する
func IsAPIKey(token string) bool {
	return strings.HasPrefix(token, APIKeyPrefix)
}
