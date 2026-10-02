# ADR 0017: LLM呼び出しを共通GCPプロジェクトのVertex AI経由に変更

## ステータス

Accepted

## コンテキスト

これまでAI機能（月次要約・ハイライト・直近トレンド分析・意味的検索の埋め込み生成）は、
ユーザーが設定画面で登録した各自の Gemini API キー（Google AI Studio）を `user_llms.key` に
平文で保存し、そのキーで Gemini Developer API を呼び出していた
（`adr/0007-latest-trend-analysis.md`、`adr/0008-diary-highlight.md` などの「ユーザ自身のAPIキーを使用」）。

この方式には以下の問題があった。

| 問題 | 内容 |
|---|---|
| 利用のハードル | ユーザーが自分でAPIキーを発行・登録しないとAI機能を使えない |
| プライバシー | Google AI Studio の Free Tier は入力が学習に利用されるため、日記を扱うには有料APIの利用をユーザーに求める必要があった |
| 秘匿情報の保管 | 外部サービスのAPIキーをDBに平文で保持していた |

## 決定事項

### 運営者の共通GCPプロジェクトの Vertex AI 経由で Gemini を呼び出す

| 項目 | 決定 |
|---|---|
| 呼び出し先 | Vertex AI（`google.golang.org/genai` の `BackendVertexAI`）。モデルは従来と同じ `gemini-2.5-flash-lite` / `gemini-embedding-001` |
| 認証 | ADC。サービスアカウントキー（`roles/aiplatform.user`）を `./secrets/gcp-service-account.json` に置き、`GOOGLE_APPLICATION_CREDENTIALS` でコンテナに渡す |
| 設定 | `GOOGLE_CLOUD_PROJECT`（必須）、`GOOGLE_CLOUD_LOCATION`（デフォルト `global`）。読み込みは `constants.LoadVertexAIConfig` |
| クライアント | `container.geminiClientFactory` がプロセス内で1つだけ生成して使い回す（`genai.Client` は並行利用可能）。生成失敗はキャッシュせず次回再試行する |
| 未設定時の挙動 | `GOOGLE_CLOUD_PROJECT` が未設定でもサーバーは起動し、AI機能の呼び出し時にのみエラーとなる |

### AI機能は既存の機能ごとのフラグで制御する（全体の有効化フラグは設けない）

`user_llms` には以前から機能ごとのフラグがあり、自動で動く処理はこれらで対象ユーザーを絞り込んでいる。
そのため「AI機能全体を有効にしたか」を別に持つと二重管理になる。

| フラグ | 制御対象 |
|---|---|
| `auto_summary_monthly` | 月次まとめの自動生成（scheduler） |
| `auto_latest_trend_enabled` | 直近トレンド分析の自動生成（scheduler） |
| `semantic_search_enabled` | RAGの埋め込み生成・自然言語検索（scheduler / subscriber / 検索API） |

| 項目 | 決定 |
|---|---|
| 手動実行（月次まとめ生成・ハイライト生成・トレンド分析の手動実行） | ボタン操作そのものを明示的な同意とみなし、フラグは確認しない。ログインユーザーなら誰でも実行できる |
| 設定レコード | `UpdateAutoSummarySettings` で初めて保存したときに作成する（upsert）。レコードが無い場合は全フラグ false として扱う |
| API | `UpdateLLMKey` / `DeleteLLMKey` を削除。`LLMKeyInfo` は `LLMSettingInfo`（`GetUserInfoResponse.llm_settings`）に改名し、旧 `key` フィールドは `reserved 2` |
| 既存のキー | `user_llms.key` カラムを削除する（平文で保存されていたキーを残さない） |
| 埋め込みベクトル | Vertex AI でも同じ `gemini-embedding-001` を使うため互換性があり、再生成はしない |
| データの扱いの明示 | 設定画面の自動まとめ設定に「日記は運営者のVertex AIに送信されるが学習には使われない」旨を表示する |

## 結果

| 観点 | 影響 |
|---|---|
| ユーザー体験 | APIキーの用意が不要になり、手動実行の機能はすぐ使える |
| プライバシー | Vertex AI に送信したデータはモデルの学習に使われない |
| コスト | LLM利用料は運営者負担になる。新規登録は `REGISTER_KEY` で制限されており、自動処理は機能ごとのフラグを有効にしたユーザーのみが対象 |
| 運用 | 本番環境では `./secrets/gcp-service-account.json` の配置と `GOOGLE_CLOUD_PROJECT` の設定が必須 |
| 互換性 | 旧RPC（`UpdateLLMKey` / `DeleteLLMKey`）は削除した。iOSアプリでは使っていない |

## 補足

- GCP側の設定手順（APIの有効化・サービスアカウント作成・キーのローテーション等）は `GCP.md` にまとめている

- `global` エンドポイントで利用できないモデルが出てきた場合は、`GOOGLE_CLOUD_LOCATION` を `asia-northeast1` などのリージョンに変える
- 今後 Vertex AI 以外のプロバイダーを追加する場合でも、実装が複数そろうまではファクトリをインターフェース化して抽象化しない（YAGNI）
