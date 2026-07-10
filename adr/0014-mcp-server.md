# ADR 0014: バックエンドMCPサーバーの追加

## ステータス

Accepted

## コンテキスト

AIクライアント（Claude Desktopなど）から自分の日記データを直接参照できるようにしたい。
具体的には以下3つの読み取り操作をMCP（Model Context Protocol）ツールとして提供する。

- 指定した日付範囲の日記取得
- キーワードによる全文検索
- 自然言語クエリによるあいまい（意味的）検索

## 決定事項

### トランスポート: MCP Streamable HTTP

バックエンドはDockerコンテナ上で稼働するリモートサービスであり、ローカルサブプロセスとして起動する
stdioトランスポートは適さない。既存の ConnectRPC エンドポイント（`:8013`）と同様に、`cmd/server` プロセス内で
追加の `net/http` サーバーを `:8014`（ホスト側 `2014`）で起動し、公式Go SDK
（`github.com/modelcontextprotocol/go-sdk`）の `StreamableHTTPHandler` を `Stateless: true` で使用する。
Statelessモードを選ぶことで、セッションの永続化やクリーンアップを実装する必要がなくなる
（各リクエストが独立したBearerトークンで認証されるため、セッション状態を持つメリットが薄い）。

### 認証: 既存のJWT Bearerトークンを再利用

新しい認証方式（APIキー等）を導入せず、gRPC/ConnectRPCの `AuthInterceptor` と同じロジック
（`Authorization: Bearer <accessToken>` を検証し `model.ParseAuthTokens` でユーザーIDを取得）を
`net/http` ミドルウェアとして実装し、MCPハンドラーの前段に挟む（`backend/infrastructure/mcpserver/auth.go`）。
`mcp.NewStreamableHTTPHandler` は `req.Context()` をそのままサーバーに渡すため、ミドルウェアで注入した
`context.WithValue` の値（`middleware.UserIDKey`）はツールハンドラー内でもそのまま参照できる。

アクセストークンは15分で失効するため、MCPクライアント側は既存の `RefreshAccessToken` エンドポイントで
更新する必要がある。長期間有効なトークン方式（APIキー等）の導入は将来の検討課題とする。

### ビジネスロジックの共有

全文検索・あいまい検索はエンティティ展開やLLM呼び出しを含む既存ロジック
（`DiaryEntry.SearchDiaryEntries` / `SearchDiaryEntriesSemantic`）と重複させないため、
ユーザーIDを直接引数に取るコア関数（`SearchDiaryEntriesByUserID` / `SearchDiaryEntriesSemanticByUserID`）
に切り出し、gRPCハンドラーとMCPツールの両方から呼び出す。日付範囲取得は新規に
`GetDiaryEntriesByDateRange`（日単位、両端含む）を追加し、`database.DiariesByUserIDAndDateRangeDays` を
経由してSQLを実行する。

## 影響

- 新規依存: `github.com/modelcontextprotocol/go-sdk`
- 新規ポート: `2014`（コンテナ内 `8014`）
- 既存のgRPC/ConnectRPCの認可モデルをそのまま踏襲するため、追加のセキュリティレビューコストが小さい
