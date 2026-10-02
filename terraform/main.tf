# 手順は ../GCP.md

locals {
  create_budget = var.billing_account_id != ""
}

resource "google_project_service" "aiplatform" {
  service = "aiplatform.googleapis.com"

  # destroy時にAPIを無効化すると同プロジェクトの他用途に影響するため無効化しない
  disable_on_destroy = false
}

# サービスアカウント作成に必要なIAM API
resource "google_project_service" "iam" {
  service            = "iam.googleapis.com"
  disable_on_destroy = false
}

# 予算アラートに必要なBilling Budgets API
resource "google_project_service" "billingbudgets" {
  count              = local.create_budget ? 1 : 0
  service            = "billingbudgets.googleapis.com"
  disable_on_destroy = false
}

# キーはstateに秘密鍵を残さないためTerraformでは作らず、gcloudで発行する
resource "google_service_account" "vertex" {
  account_id   = var.service_account_id
  display_name = "umi.mikan Vertex AI"
  description  = "umi.mikanのbackend/subscriberからVertex AI（Gemini）を呼び出す"

  depends_on = [google_project_service.iam]
}

# 最小権限としてVertex AIユーザーのロールのみ付与する
resource "google_project_iam_member" "vertex_user" {
  project = var.project_id
  role    = "roles/aiplatform.user"
  member  = "serviceAccount:${google_service_account.vertex.email}"
}

# 月額予算アラート（50% / 90% / 100%で請求先アカウントの管理者にメール通知。利用は止まらない）
resource "google_billing_budget" "monthly" {
  count           = local.create_budget ? 1 : 0
  billing_account = var.billing_account_id
  display_name    = "umi.mikan Vertex AI monthly budget"

  budget_filter {
    projects = ["projects/${data.google_project.current.number}"]
  }

  amount {
    specified_amount {
      currency_code = var.budget_currency
      units         = tostring(var.monthly_budget_amount)
    }
  }

  threshold_rules {
    threshold_percent = 0.5
  }
  threshold_rules {
    threshold_percent = 0.9
  }
  threshold_rules {
    threshold_percent = 1.0
  }

  depends_on = [google_project_service.billingbudgets]
}

# 予算フィルタ用。plan時に読むため Cloud Resource Manager API は事前にgcloudで有効化しておく
data "google_project" "current" {
  project_id = var.project_id
}
