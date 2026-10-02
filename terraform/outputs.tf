output "service_account_email" {
  description = "Vertex AI呼び出し用のサービスアカウント"
  value       = google_service_account.vertex.email
}
