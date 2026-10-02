terraform {
  required_version = ">= 1.9.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = ">= 6.0"
    }
  }

  # バケット名は環境ごとに異なるため make tf-init で渡す
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
