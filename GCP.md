# GCP（Vertex AI）設定手順

umi.mikan のAI機能（月次まとめ・ハイライト・直近トレンド分析・自然言語検索の埋め込み生成）は、運営者の共通GCPプロジェクトの Vertex AI 経由で Gemini を呼び出す。
ユーザーごとのAPIキーは不要である。設計の背景は `adr/0017-vertex-ai.md` を参照。

## 全体像

| 項目 | 内容 |
|---|---|
| 呼び出し先 | Vertex AI（`aiplatform.googleapis.com`） |
| 使用モデル | `gemini-2.5-flash-lite`（テキスト生成）、`gemini-embedding-001`（埋め込み） |
| 認証方式 | サービスアカウントキー（JSON）による ADC |
| 必要なロール | `roles/aiplatform.user`（Vertex AI ユーザー） |
| 呼び出すコンテナ | `backend`（意味的検索のクエリ埋め込み）、`subscriber`（要約・ハイライト・トレンド・埋め込み生成） |
| 課金 | GCPプロジェクトの請求先アカウントに計上される |
| データの扱い | Vertex AI に送信した内容はモデルの学習に使われない |

## 必要な環境変数

| 変数 | 必須 | 既定値 | 説明 |
|---|---|---|---|
| `GOOGLE_CLOUD_PROJECT` | ○ | なし | GCPプロジェクトID。未設定でもサーバーは起動するが、AI機能の呼び出し時にエラーになる |
| `GOOGLE_CLOUD_LOCATION` | - | `global` | Vertex AI のロケーション |
| `GOOGLE_APPLICATION_CREDENTIALS` | ○ | なし | サービスアカウントキーのパス。compose では `/secrets/gcp-service-account.json` に固定している |

## 1. GCP側の準備（初回のみ）

Terraform（`terraform/`）で必要なリソースをまとめて作成する。stateはGCSバケットに置くため、誰がどのマシンから実行しても同じstateを参照できる。手作業で作りたい場合は「1-B. gcloudで手動作成」を参照。

### Terraformで作成されるリソース

| リソース | 内容 |
|---|---|
| `google_project_service` | Vertex AI API・IAM API を有効化する（予算アラートを作る場合は Billing Budgets API も有効化する）。destroy してもAPIは無効化しない |
| `google_service_account` | `umi-mikan-vertex`（backend/subscriber 用） |
| `google_project_iam_member` | 上記サービスアカウントに `roles/aiplatform.user` のみ付与する |
| `google_billing_budget` | 月額予算アラート（50% / 90% / 100% で請求先アカウントの管理者にメール通知）。`billing_account_id` を指定した場合のみ作成する |

以下は Terraform の管理外とする。

| 対象 | 理由 |
|---|---|
| GCPプロジェクト・請求先アカウントの紐付け | 既存プロジェクトを使う前提のため |
| state保存用のGCSバケット | stateの置き場所そのものなので、Terraformより先に存在している必要がある |
| Cloud Resource Manager API・Service Usage API | Terraformがプロジェクト情報の読み取り・IAM付与・APIの有効化に使うため、Terraformより先に有効になっている必要がある |
| サービスアカウントキー | Terraformで発行すると秘密鍵がstate（共有バケット）に平文で残るため、gcloudで発行する |

### プロジェクトIDの環境変数

zsh などで `UMI_MIKAN_PROJECT_ID` をexportしておくと、`make tf-*`・開発環境の `docker compose`・本書のgcloudコマンドがそのまま使える。`~/.zshrc` に追記して反映する（`echo 'export UMI_MIKAN_PROJECT_ID=your-gcp-project-id' >> ~/.zshrc && source ~/.zshrc`）。

| 使われる場所 | 内容 |
|---|---|
| `make tf-init` | stateのバケット名を `${UMI_MIKAN_PROJECT_ID}-tfstate` として渡す（未設定なら `terraform/backend.hcl` を使う） |
| `make tf-plan` / `make tf-apply` | `TF_VAR_project_id` として `project_id` に渡す（未設定なら `terraform/terraform.tfvars` を使う） |
| `docker compose`（開発環境） | backend / subscriber の `GOOGLE_CLOUD_PROJECT` に渡す |
| 本書のgcloudコマンド | `${UMI_MIKAN_PROJECT_ID}` を展開してそのまま実行できる |

`terraform.tfvars` に `project_id` を書いた場合は、環境変数よりそちらが優先される（Terraformの仕様）。どちらも無い場合は `project_id` が空だというエラーで止まる。

### 1-A. Terraformで作成（推奨）

| # | 作業 | コマンド |
|---|---|---|
| 1 | gcloudにログイン | `gcloud auth login` |
| 2 | Terraform用の認証情報（ADC）を作成 | `gcloud auth application-default login` |
| 3 | プロジェクトが無ければ作成 | `gcloud projects create ${UMI_MIKAN_PROJECT_ID}` |
| 4 | 請求先アカウントを紐付ける（未設定の場合） | `gcloud billing projects link ${UMI_MIKAN_PROJECT_ID} --billing-account=XXXXXX-XXXXXX-XXXXXX` |
| 5 | state保存用のバケットを作成（初回の1人だけ。us-central1 は Cloud Storage の無料枠（月5GB）の対象のため、stateの保管は0円になる） | `gcloud storage buckets create gs://${UMI_MIKAN_PROJECT_ID}-tfstate --project=${UMI_MIKAN_PROJECT_ID} --location=us-central1 --uniform-bucket-level-access --public-access-prevention` |
| 6 | バケットのバージョニングを有効化（stateを壊したときに戻せるようにする） | `gcloud storage buckets update gs://${UMI_MIKAN_PROJECT_ID}-tfstate --versioning` |
| 6-2 | Terraformが使うAPIを有効化（初回の1人だけ。反映に1〜2分かかる） | `gcloud services enable cloudresourcemanager.googleapis.com serviceusage.googleapis.com --project=${UMI_MIKAN_PROJECT_ID}` |
| 7 | `UMI_MIKAN_PROJECT_ID` を使わない場合のみ、backend設定ファイルを作成 | `cp terraform/backend.hcl.example terraform/backend.hcl` を実行し、`bucket` に手順5のバケット名を書く |
| 8 | 予算アラートを作る場合（または `UMI_MIKAN_PROJECT_ID` を使わない場合）のみ、変数ファイルを作成 | `cp terraform/terraform.tfvars.example terraform/terraform.tfvars` を実行し、`billing_account_id`（環境変数を使わないなら `project_id` も）を書く |
| 9 | 初期化 | `make tf-init` |
| 10 | 差分を確認 | `make tf-plan` |
| 11 | 適用 | `make tf-apply` |
| 12 | キーを発行して `./secrets` に保存（リポジトリ直下で実行） | `gcloud iam service-accounts keys create ./secrets/gcp-service-account.json --iam-account=umi-mikan-vertex@${UMI_MIKAN_PROJECT_ID}.iam.gserviceaccount.com` |

2人目以降・別のマシンからは、手順1〜2と7〜10を行えばよい（バケットとリソースは作成済みなので、planで差分が出ないことを確認する）。キーは既存のものを安全な経路で受け取るか、手順12で別のキーを発行する。

請求先アカウントのIDは `gcloud billing accounts list` で確認できる。

予算アラートを作る場合、Terraform を実行するアカウントには請求先アカウントの「請求先アカウント管理者」（または予算の編集権限）が必要である。

### Terraformの変数

| 変数 | 必須 | 既定値 | 説明 |
|---|---|---|---|
| `project_id` | ○ | なし | GCPプロジェクトID |
| `service_account_id` | - | `umi-mikan-vertex` | サービスアカウントID |
| `billing_account_id` | - | 空 | 空なら予算アラートを作らない |
| `monthly_budget_amount` | - | `1000` | 月額予算（超過しても止まらず、通知のみ） |
| `budget_currency` | - | `JPY` | 請求先アカウントの通貨と一致させる |

### tfstateの取り扱い

| 項目 | 内容 |
|---|---|
| 保存場所 | GCSバケット `gs://<backend.hclのbucket>/umi-mikan/vertex-ai/default.tfstate` |
| 排他制御 | GCSバックエンドはロックに対応しているため、同時に apply しても壊れない |
| 秘密情報 | キーをTerraformで発行しないため、stateに秘密鍵は含まれない |
| アクセス権 | バケットはプロジェクトのオーナー／編集者のみが読み書きできる（均一なバケットレベルのアクセス、公開アクセス防止を有効化している） |
| ローカルのファイル | `backend.hcl` / `terraform.tfvars` / `.terraform/` / `*.tfstate` は `terraform/.gitignore` で除外している。`.terraform.lock.hcl` はproviderのバージョン固定のためコミットする |

### 1-B. gcloudで手動作成

Terraform を使わない場合の手順である。以下のコマンドは、環境変数 `UMI_MIKAN_PROJECT_ID` にプロジェクトIDが入っている前提である（「プロジェクトIDの環境変数」を参照）。

| # | 作業 | コマンド |
|---|---|---|
| 1 | ログイン | `gcloud auth login` |
| 2 | プロジェクト作成（既存プロジェクトを使う場合は不要） | `gcloud projects create ${UMI_MIKAN_PROJECT_ID}` |
| 3 | 操作対象のプロジェクトを設定 | `gcloud config set project ${UMI_MIKAN_PROJECT_ID}` |
| 4 | 請求先アカウントを紐付ける（未設定の場合） | `gcloud billing projects link ${UMI_MIKAN_PROJECT_ID} --billing-account=XXXXXX-XXXXXX-XXXXXX` |
| 5 | Vertex AI API を有効化 | `gcloud services enable aiplatform.googleapis.com` |
| 6 | サービスアカウント作成 | `gcloud iam service-accounts create umi-mikan-vertex --display-name="umi.mikan Vertex AI"` |
| 7 | Vertex AI ユーザーのロールを付与 | `gcloud projects add-iam-policy-binding ${UMI_MIKAN_PROJECT_ID} --member="serviceAccount:umi-mikan-vertex@${UMI_MIKAN_PROJECT_ID}.iam.gserviceaccount.com" --role="roles/aiplatform.user"` |
| 8 | キーを発行して `./secrets` に保存（リポジトリ直下で実行） | `gcloud iam service-accounts keys create ./secrets/gcp-service-account.json --iam-account=umi-mikan-vertex@${UMI_MIKAN_PROJECT_ID}.iam.gserviceaccount.com` |

サービスアカウントに付与するロールは `roles/aiplatform.user` だけにする。オーナーや編集者のロールは付与しない。

## 2. 開発環境の設定

| # | 作業 | 内容 |
|---|---|---|
| 1 | キーを配置 | `./secrets/gcp-service-account.json` に置く（手順1-A-12または1-B-8で発行済みなら不要）。`secrets/` は git 管理外である |
| 2 | プロジェクトIDを設定 | `UMI_MIKAN_PROJECT_ID` をexportしておく（「プロジェクトIDの環境変数」を参照）。`docker compose` を実行するシェルで設定されている必要がある |
| 3 | コンテナを再作成 | `docker compose up -d backend subscriber` |
| 4 | 環境変数の反映を確認 | `docker compose exec subscriber env \| grep GOOGLE` |

`compose.yml` は `GOOGLE_CLOUD_PROJECT: ${UMI_MIKAN_PROJECT_ID:-}` の形で渡している。ロケーションは `global` に固定しており、変える場合は `compose.yml` の `GOOGLE_CLOUD_LOCATION` を書き換える。

## 3. 本番環境の設定

| # | 作業 | 内容 |
|---|---|---|
| 1 | キーを配置 | 本番サーバーの `compose-prod.yml` と同じディレクトリに `secrets/gcp-service-account.json` を置く（`chmod 600` 推奨） |
| 2 | `compose-prod.yml` を更新 | `compose-prod.example.yml` と同じように、`backend` と `subscriber` に `GOOGLE_CLOUD_PROJECT` / `GOOGLE_CLOUD_LOCATION` / `GOOGLE_APPLICATION_CREDENTIALS` を設定し、`./secrets:/secrets:ro` をマウントする |
| 3 | スキーマを適用 | `user_llms` の `key`・`llm_provider` カラムの削除を含む。`make db-apply` が pgvector のバージョン差で失敗する場合は `ALTER TABLE user_llms DROP COLUMN key; ALTER TABLE user_llms DROP CONSTRAINT unique_user_llm; ALTER TABLE user_llms DROP COLUMN llm_provider;` を psql で実行する |
| 4 | 再起動 | `docker compose -f compose-prod.yml up -d backend subscriber` |

開発用と本番用でサービスアカウント（またはプロジェクト）を分けておくと、キーが漏洩したときの影響範囲と、コストの切り分けがしやすい。

## 4. 動作確認

| # | 確認内容 | 方法 |
|---|---|---|
| 1 | ハイライト生成 | 500文字以上の日記を開いてハイライト生成ボタンを押し、生成されることを確認する |
| 2 | subscriber のログ | `docker compose logs -f subscriber` でエラーが出ていないことを確認する |
| 3 | 意味的検索 | 設定画面で自然言語検索を有効にし、`/search` で意味的検索ができることを確認する |
| 4 | API呼び出しの記録 | GCPコンソールの「APIとサービス」→「Vertex AI API」でリクエスト数が増えていることを確認する |
| 5 | 精度評価（任意） | `make b-test-semantic-eval`（backend コンテナの Vertex AI 設定で実際の API を呼ぶ） |

## 5. コスト管理

### 見積もり

単価は Vertex AI の料金表（2026年10月時点、オンライン推論）による。

| モデル | 入力 | 出力 |
|---|---|---|
| `gemini-2.5-flash-lite` | $0.10 / 100万トークン | $0.40 / 100万トークン |
| `gemini-embedding-001` | $0.15 / 100万トークン | 無料 |

前提は、1ユーザーが毎日1,000文字の日記を書き、全機能のフラグを有効にした場合とする（日本語は1文字≒1トークンで多めに見積もる）。

| 処理 | 頻度（月） | 入力 | 出力 | 費用 |
|---|---|---|---|---|
| 月次まとめ | 1回（1か月分の日記＋プロンプト） | 3.1万 | 0.1万 | 約$0.004 |
| 直近トレンド | 30回（3日分の日記＋プロンプト） | 12万 | 3万 | 約$0.024 |
| ハイライト | 60回（1日2回再生成する想定） | 9万 | 3万 | 約$0.021 |
| RAGのチャンク分割 | 30回 | 4.5万 | 3.6万 | 約$0.019 |
| 埋め込み（日記＋検索クエリ） | 30回＋検索100回 | 3.8万 | - | 約$0.006 |
| **合計** | | | | **約$0.07（約11円）/ユーザー/月** |

| 規模 | 月額の目安 |
|---|---|
| 1ユーザー | 約11円 |
| 5ユーザー | 約55円 |
| 20ユーザー | 約220円 |
| 埋め込みの全再生成（3年分・約1,000件、1回だけ） | 約120円/ユーザー |
| tfstate用のGCSバケット | us-central1 に置くため無料枠内で0円 |

予算アラートの既定値は **1,000円/月** とした。数人規模なら通常は100円未満に収まるため、50%（500円）の通知が来た時点で、ループや不正利用などの異常を疑える。利用者が増えたら `monthly_budget_amount` を引き上げる。

### 管理方法

| 項目 | 内容 |
|---|---|
| 予算アラート | Terraform で `billing_account_id` を指定すると自動で作成される。手動で作る場合は、GCPコンソールの「お支払い」→「予算とアラート」で設定する |
| 利用状況 | `/llm` ページと Grafana（`umi-mikan-pubsub` / `umi-mikan-rag` ダッシュボード）で処理件数を確認する |
| 対象ユーザーの制限 | 新規登録は `REGISTER_KEY` で制限している。自動処理は、機能ごとのフラグを有効にしたユーザーだけが対象である |
| 割り当て（クォータ） | 「IAMと管理」→「割り当て」で Vertex AI のリクエスト上限を確認し、必要に応じて引き上げを申請する |

## 6. キーのローテーション

キーは Terraform の管理外なので、gcloud で以下の手順で行う。

| # | 作業 | コマンド |
|---|---|---|
| 1 | 新しいキーを発行 | `gcloud iam service-accounts keys create ./secrets/gcp-service-account.json --iam-account=umi-mikan-vertex@${UMI_MIKAN_PROJECT_ID}.iam.gserviceaccount.com` |
| 2 | コンテナを再起動 | `docker compose up -d --force-recreate backend subscriber` |
| 3 | 古いキーのIDを確認 | `gcloud iam service-accounts keys list --iam-account=umi-mikan-vertex@${UMI_MIKAN_PROJECT_ID}.iam.gserviceaccount.com` |
| 4 | 古いキーを削除 | `gcloud iam service-accounts keys delete KEY_ID --iam-account=umi-mikan-vertex@${UMI_MIKAN_PROJECT_ID}.iam.gserviceaccount.com` |

クライアントは初回の呼び出し時に1度だけ生成して使い回すため、キーを差し替えたらコンテナの再起動が必要である。

## 7. トラブルシューティング

| 症状（ログに出るエラー） | 原因 | 対処 |
|---|---|---|
| `vertex AI project and location are required` | `GOOGLE_CLOUD_PROJECT` が未設定 | 開発は `echo $UMI_MIKAN_PROJECT_ID` で確認し、そのシェルで `docker compose up -d backend subscriber` する（本番は `compose-prod.yml` を確認する） |
| `could not find default credentials` / キーファイルが見つからない | キーが `./secrets/gcp-service-account.json` に無い | ファイル名と配置場所を確認する。コンテナ内では `docker compose exec subscriber ls /secrets` で確認できる |
| `PERMISSION_DENIED` / `403` | ロール未付与、または API が無効 | `make tf-apply` を再実行する（gcloud の場合は手順1-B-5と1-B-7） |
| `SERVICE_DISABLED` | Vertex AI API が無効 | `gcloud services enable aiplatform.googleapis.com` |
| `NOT_FOUND`（モデルが見つからない） | 指定したロケーションでそのモデルが使えない | `GOOGLE_CLOUD_LOCATION` を `us-central1` や `asia-northeast1` に変える |
| `RESOURCE_EXHAUSTED` / `429` | クォータ超過 | 時間をおいて再試行するか、クォータの引き上げを申請する。`SUBSCRIBER_MAX_CONCURRENT_JOBS` を下げるのも有効 |
| キー発行時に `constraints/iam.disableServiceAccountKeyCreation` | 組織ポリシーでキー作成が禁止されている | 組織の管理者にポリシーの例外を依頼する |
| `billing` 関連のエラー | 請求先アカウントが未設定 | `gcloud billing projects link` で請求先を紐付ける |
| `make tf-plan` で `project_idが空です` | `UMI_MIKAN_PROJECT_ID` が未exportで、`terraform.tfvars` も無い | `echo $UMI_MIKAN_PROJECT_ID` で確認し、`~/.zshrc` にexportを書いて `source ~/.zshrc` する |
| `make tf-plan` で `Cloud Resource Manager API has not been used in project` | Terraformが使うAPIが無効 | 手順1-A-6-2を実行し、1〜2分待ってから再実行する |
| `terraform plan` で `could not find default credentials` | Terraform用の ADC が無い | `gcloud auth application-default login` |
| `make tf-init` で `bucket doesn't exist` / `backend.hcl` が無い | stateのバケット、または `terraform/backend.hcl` が未作成 | 手順1-A-5〜7を行う |
| `make tf-init` / `make tf-plan` で `bucket doesn't exist` / `404` が出たり出なかったりする | stateのバケットを同じ名前で削除・再作成した直後で、GCS側の反映が終わっていない | 数分〜十数分待つ。その間に残ったロックは `terraform force-unlock` で外す |
| `make tf-apply` で `Error acquiring the state lock` | 他の人の apply が実行中、または中断してロックが残っている | 実行中でないことを確認してから `cd terraform && terraform force-unlock LOCK_ID` |
| 予算作成時に `PERMISSION_DENIED` | 請求先アカウントの権限不足 | 請求先アカウント管理者に依頼するか、`billing_account_id` を空にして予算を作らない |
