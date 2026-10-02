terraform {
  required_version = ">= 1.9.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 6.0"
    }
  }

  # stateはGCSに置き、誰がどのマシンから実行しても同じstateを参照する
  # バケット名は環境ごとに異なるため backend.hcl で渡す（make tf-init が -backend-config=backend.hcl を指定する）
  backend "gcs" {
    prefix = "umi-mikan/vertex-ai"
  }
}

provider "google" {
  project = var.project_id

  # 予算（Billing Budgets API）はリクエスト元プロジェクトの指定が必要なため有効化する
  user_project_override = true
  billing_project       = var.project_id
}
