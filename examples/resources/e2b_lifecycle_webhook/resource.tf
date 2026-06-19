variable "e2b_webhook_signature_secret" {
  description = "Secret E2B uses to sign lifecycle webhook payloads."
  type        = string
  sensitive   = true
}

resource "e2b_lifecycle_webhook" "audit" {
  name             = "terraform-audit"
  url              = "https://example.com/e2b/lifecycle"
  enabled          = true
  signature_secret = var.e2b_webhook_signature_secret
  events           = ["sandbox.started", "sandbox.paused", "sandbox.killed"]
}
