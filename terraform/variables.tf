variable "project_id" {
  description = "Vertex AIを利用するGCPプロジェクトID（請求先アカウントが紐付いた既存プロジェクト）。make経由では環境変数 UMI_MIKAN_PROJECT_ID から渡される"
  type        = string

  validation {
    condition     = length(var.project_id) > 0
    error_message = "project_idが空です。UMI_MIKAN_PROJECT_IDをexportするか、terraform.tfvarsにproject_idを書いてください。"
  }
}

variable "service_account_id" {
  description = "backend/subscriberがVertex AIを呼び出すためのサービスアカウントID"
  type        = string
  default     = "umi-mikan-vertex"
}

variable "billing_account_id" {
  description = "予算アラートを作成する請求先アカウントID（XXXXXX-XXXXXX-XXXXXX）。空なら予算アラートを作成しない"
  type        = string
  default     = ""
}

variable "monthly_budget_amount" {
  description = "月額予算（budget_currency単位の整数）。超過しても止まらず通知のみ"
  type        = number
  default     = 1000
}

variable "budget_currency" {
  description = "予算の通貨コード。請求先アカウントの通貨と一致させる"
  type        = string
  default     = "JPY"
}
