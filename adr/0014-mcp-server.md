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

### 認証: APIキー（推奨）とJWT Bearerトークンの2方式

`Authorization: Bearer <トークン>` ヘッダーを検証する `net/http` ミドルウェアをMCPハンドラーの
前段に挟む（`backend/infrastructure/mcpserver/auth.go`）。トークンは2種類を受け付ける。

1. **APIキー（`umi_` プレフィックス、MCPクライアント向けの推奨方式）**
   - JWTアクセストークンは15分で失効するため、Claude Desktopのような設定ファイルにトークンを
     書くタイプのMCPクライアントでは実用にならない。長期間有効なAPIキーを導入する。
   - キーはWebフロントエンドの設定ページ（`/settings` のAPI連携セクション）から発行・削除する。
     バックエンドは `UserService.CreateApiKey` / `ListApiKeys` / `DeleteApiKey` を提供する。
   - キー本体（`umi_` + 乱数32バイトのhex）は発行時に一度だけ返し、DBには SHA-256 ハッシュのみを
     保存する（`user_api_keys` テーブル）。認証時はハッシュで照合し、`last_used_at` を更新する。
   - 失効はDBの行削除（`DeleteApiKey`）で即時反映される。
2. **JWTアクセストークン**
   - gRPC/ConnectRPCの `AuthInterceptor` と同じロジック（`model.ParseAuthTokens`）で検証する。
   - プログラムから短命トークンで叩くケース（テスト等）向けに残している。

`umi_` プレフィックスの有無でどちらの方式かを判別する。`mcp.NewStreamableHTTPHandler` は
`req.Context()` をそのままサーバーに渡すため、ミドルウェアで注入した `context.WithValue` の値
（`middleware.UserIDKey`）はツールハンドラー内でもそのまま参照できる。

MCPクライアント（Claude Desktop等）の設定例:

```json
{
  "mcpServers": {
    "umi-mikan": {
      "type": "http",
      "url": "http://localhost:2014/",
      "headers": {
        "Authorization": "Bearer umi_xxxxxxxx..."
      }
    }
  }
}
```

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
- 新規テーブル: `user_api_keys`（`schema/1300_user_api_keys.sql`）
- 既存のgRPC/ConnectRPCの認可モデルをそのまま踏襲するため、追加のセキュリティレビューコストが小さい
