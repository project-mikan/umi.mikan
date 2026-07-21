// Package mcpserver は、日記データをMCP（Model Context Protocol）ツールとして
// AIクライアント（Claude Desktopなど）に公開するサーバーを提供する。
// 認証は既存のgRPC/ConnectRPCと同じJWT Bearerトークン方式、およびOAuth 2.0
// Authorization Code + PKCEフロー（発行後の実体はAPIキー、adr/0016参照）を再利用する。
package mcpserver

import (
	"database/sql"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
	"github.com/project-mikan/umi.mikan/backend/service/user"
	"github.com/redis/rueidis"
)

const serverName = "umi-mikan-diary"

const serverVersion = "1.0.0"

// mcpPath はMCP Streamable HTTPハンドラー自体のパス
const mcpPath = "/"

// OAuth関連エンドポイントのパス。newAuthorizationServerMetadataHandlerが
// これらを絶対URLに組み立てて返すため、一箇所にまとめておく。
const (
	oauthProtectedResourceMetadataPath   = "/.well-known/oauth-protected-resource"
	oauthAuthorizationServerMetadataPath = "/.well-known/oauth-authorization-server"
	oauthRegisterPath                    = "/register"
	oauthAuthorizePath                   = "/oauth/authorize"
	oauthConsentPath                     = "/oauth/consent"
	oauthTokenPath                       = "/oauth/token"
)

// frontendConsentPath はフロントエンド（SvelteKit）側の同意画面のパス
const frontendConsentPath = "/oauth/authorize"

// NewServer は日記操作ツールを登録したMCPサーバーを作成する
func NewServer(diaryService *diary.DiaryEntry) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: serverName, Version: serverVersion}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_diary_entries_by_range",
		Description: "指定した日付範囲（開始日〜終了日、両端含む）の日記エントリを取得する",
	}, getDiaryEntriesByRangeHandler(diaryService))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_diary_entries_fulltext",
		Description: "キーワードで日記を全文検索する。登録済みの人物・エンティティ名の場合、関連する別名やエイリアスにも自動展開して検索される",
	}, searchDiaryEntriesFulltextHandler(diaryService))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_diary_entries_fuzzy",
		Description: "自然言語クエリで日記を意味的（あいまい）に検索する。ユーザーがセマンティック検索機能を有効化している場合のみ利用可能",
	}, searchDiaryEntriesFuzzyHandler(diaryService))

	return server
}

// NewHTTPHandler は認証（APIキーまたはJWT）付きのMCP Streamable HTTPハンドラーと、
// OAuth 2.0 Discovery・Dynamic Client Registration・Authorization Code + PKCEフローの
// 各エンドポイント（adr/0016参照）をまとめて登録したハンドラーを作成する。
//
//   - baseURL: このMCPサーバー自身の公開URL（例: https://umi-mikan-api.usuyuki.net）。
//     OAuth Discoveryメタデータ内のエンドポイントURL組み立てに使う。
//   - frontendBaseURL: フロントエンド（SvelteKit）の公開URL。/oauth/authorizeの
//     リダイレクト先（ログイン・同意画面）組み立てに使う。
func NewHTTPHandler(diaryService *diary.DiaryEntry, db *sql.DB, redisClient rueidis.Client, userService *user.UserEntry, baseURL, frontendBaseURL string) http.Handler {
	server := NewServer(diaryService)
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{Stateless: true})

	mux := http.NewServeMux()
	mux.Handle(mcpPath, AuthMiddleware(db, mcpHandler))
	mux.HandleFunc(oauthProtectedResourceMetadataPath, newProtectedResourceMetadataHandler(baseURL))
	mux.HandleFunc(oauthAuthorizationServerMetadataPath, newAuthorizationServerMetadataHandler(baseURL))
	mux.HandleFunc(oauthRegisterPath, newRegisterHandler())
	mux.HandleFunc(oauthAuthorizePath, newAuthorizeHandler(frontendBaseURL))
	mux.HandleFunc(oauthConsentPath, newConsentHandler(redisClient))
	mux.HandleFunc(oauthTokenPath, newTokenHandler(redisClient, userService))

	return mux
}
