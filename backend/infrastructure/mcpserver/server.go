// Package mcpserver は、日記データをMCP（Model Context Protocol）ツールとして
// AIクライアント（Claude Desktopなど）に公開するサーバーを提供する。
// 認証は既存のgRPC/ConnectRPCと同じJWT Bearerトークン方式を再利用する。
package mcpserver

import (
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/project-mikan/umi.mikan/backend/service/diary"
)

const serverName = "umi-mikan-diary"

const serverVersion = "1.0.0"

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

// NewHTTPHandler はJWT認証付きのMCP Streamable HTTPハンドラーを作成する
func NewHTTPHandler(diaryService *diary.DiaryEntry) http.Handler {
	server := NewServer(diaryService)
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{Stateless: true})

	return AuthMiddleware(mcpHandler)
}
