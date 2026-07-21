# ADR 0016: MCPサーバーのOAuth 2.0対応

## ステータス

Accepted

## コンテキスト

`adr/0014-mcp-server.md` で追加したMCPサーバーは、`Authorization: Bearer <トークン>` ヘッダーを
直接設定できるMCPクライアント（Claude Desktopの設定ファイルなど）を前提に、APIキー/JWT認証のみを
実装していた。

しかしClaude.ai（Web版）のカスタムコネクタ機能は、接続時に
[MCP仕様のAuthorization](https://spec.modelcontextprotocol.io/specification/basic/authorization/)
に従いOAuth 2.0 Discoveryを試みる。具体的には以下の順でアクセスし、これに応答できないサーバーへの
接続は失敗する（「サインインサービスに登録できませんでした」エラー）。

1. `GET /.well-known/oauth-protected-resource`（RFC9728）
2. `GET /.well-known/oauth-authorization-server`（RFC8414）
3. `POST /register`（RFC7591 Dynamic Client Registration）
4. Authorization Code + PKCEフロー（ブラウザでのログイン・同意画面）

Claude.aiからumi.mikanのMCPサーバーに接続するには、この一連のOAuthフローに対応する必要がある。

## 決定事項

### 最小構成のOAuth 2.0 Authorization Serverを追加する

フルスペックのOAuthサーバーを新規実装するのではなく、既存資産（APIキー発行ロジック・
フロントエンドのログインセッション・Redis）を再利用した最小構成にする。

- **アクセストークンの実体は既存のAPIキー**: `/oauth/token` は新しいトークン管理の仕組みを
  作らず、既存の `user.UserEntry.CreateApiKeyForUser`（`backend/service/user/api_key.go` から
  `CreateApiKey` のコア処理を切り出したもの）を呼び出し、発行された `umi_` プレフィックスの
  APIキーをそのままOAuthの `access_token` として返す。これにより、MCPサーバーの既存の認証
  ミドルウェア（`AuthMiddleware`、`backend/infrastructure/mcpserver/auth.go`）は無改造のまま、
  OAuth経由のトークンも他のAPIキーと同じ90日有効期限・設定ページからの一覧表示/失効の対象になる
  （キー名は固定文字列 `"MCP OAuth (Claude connector)"` で判別できるようにする）。
- **同意画面は既存フロントエンドのログインセッションを再利用**: 新しい認証UIを作らず、
  SvelteKitの `/oauth/authorize` ページ（`frontend/src/routes/oauth/authorize/`）で
  既存のCookieベースのログイン状態（`ensureValidAccessToken`）を確認し、未ログインなら
  `/login` へ誘導、ログイン済みなら同意ボタンを表示する。同意すると
  `POST /oauth/consent`（バックエンド、JWTアクセストークンをBearerで送る）を呼び、
  authorization codeを発行してもらう。
- **Dynamic Client Registrationはredirect_urisをRedisに保存する最小実装**: `POST /register` は
  `redirect_uris`（http/https の絶対URLのみ許可）を受け取り、`client_id` に紐付けてRedisに
  30日TTLで保存する（`oauth_client_store.go`、DB等への永続化は行わない public client、
  `token_endpoint_auth_method: none`）。`/oauth/authorize` と `POST /oauth/consent` は、
  渡された `redirect_uri` が当該 `client_id` の登録済み `redirect_uris` と完全一致することを
  検証してから先に進む。
  - 当初の実装では `redirect_uris` の登録・照合を一切行わず、`isValidRedirectURI`
    （スキームがhttp/https・ホストが存在すること）のみを検証していた。しかしこれでは、
    第三者が任意の `client_id` を取得したうえで自分の `redirect_uri` を指定し、
    ログイン済みの被害者に同意させることで、被害者のauthorization codeを自分のサーバーへ
    誘導できてしまう（Authorization Code Interception）。PKCEは「code_verifierを持たない
    第三者による横取り」を防ぐものであり、攻撃者自身が最初にcode_challenge/code_verifier
    ペアを生成するこのシナリオでは防御にならない。セキュリティレビューでこの点を
    High判定の脆弱性として指摘され、`redirect_uri` の事前登録・一致検証を追加した。
- **PKCE (S256) を必須化**: client_secretを持たないpublic clientの安全性は、PKCEの
  `code_challenge`/`code_verifier` の一致検証で担保する。`plain` 方式は非対応とし、`S256`
  のみ受け付ける（ただし上記の通り、redirect_uri検証と組み合わせて初めて有効な防御になる）。
- **authorization codeはRedisにTTL付きで保存**: 5分の有効期限・単回使用（取得と同時に削除）
  とし、既存のRedis基盤（`rueidis.Client`、DIコンテナから注入）をそのまま使う。新しい永続化層
  は追加しない。
- **同意画面にリダイレクト先ホストを表示**: `frontend/src/routes/oauth/authorize/+page.svelte`
  で `redirect_uri` のホスト名をユーザーに表示し、何に同意しているか確認できるようにする。

### エンドポイント構成

MCPサーバー（`:8014`）に以下を追加する。ルーティングは `http.ServeMux` に切り替えた
（`backend/infrastructure/mcpserver/server.go`）。

| パス | 実装ファイル | 役割 |
|---|---|---|
| `GET /.well-known/oauth-protected-resource` | `oauth_metadata.go` | RFC9728メタデータ |
| `GET /.well-known/oauth-authorization-server` | `oauth_metadata.go` | RFC8414メタデータ |
| `POST /register` | `oauth_register.go` | Dynamic Client Registration |
| `GET /oauth/authorize` | `oauth_authorize.go` | パラメータ検証後、フロントエンドの同意画面へ302リダイレクト |
| `POST /oauth/consent` | `oauth_consent.go` | フロントエンドからの同意通知を受けてauthorization codeを発行（JWT Bearer認証） |
| `POST /oauth/token` | `oauth_token.go` | codeをAPIキー（access_token）に交換 |
| `/`（他すべて） | `server.go` | 既存のMCP Streamable HTTPハンドラー（APIキー/JWT認証） |

フロントエンド側は `frontend/src/routes/oauth/authorize/` にログイン確認・同意画面を追加し、
`frontend/src/lib/server/mcp-oauth-api.ts` からバックエンドの `/oauth/consent` を呼び出す。

### 設定

新規環境変数 `MCP_SERVER_BASE_URL`（OAuth DiscoveryメタデータのURL組み立てに使用）と
`FRONTEND_BASE_URL`（認可リクエストのリダイレクト先）を追加した
（`backend/constants/env.go`、開発環境では `localhost:2014` / `localhost:2000` がデフォルト）。

## 影響

- 新規ファイル: `backend/infrastructure/mcpserver/oauth_{metadata,register,authorize,consent,token,store,client_store}.go`
  とそれぞれのテスト
- 変更: `backend/service/user/api_key.go`（`CreateApiKeyForUser` をコア処理として切り出し）、
  `backend/infrastructure/mcpserver/server.go`（`http.ServeMux` 化、依存追加）、
  `backend/cmd/server/main.go`、`backend/constants/env.go`
- 新規: `frontend/src/routes/oauth/authorize/`、`frontend/src/lib/server/mcp-oauth-api.ts`
- 既存のAPIキー認証・JWT認証・`user_api_keys` テーブルは無変更。OAuthは「トークン発行の別入口」
  として追加されるだけで、認証・認可のコアロジックは増えない。
