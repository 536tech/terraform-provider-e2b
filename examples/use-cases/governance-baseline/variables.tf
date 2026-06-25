variable "team_id" {
  description = "E2B team ID used for read-only team metric evidence."
  type        = string
}

variable "template_name" {
  description = "Approved E2B template name."
  type        = string
}

variable "template_image" {
  description = "Container image used to build the approved E2B template."
  type        = string
  default     = "e2bdev/base:latest"
}

variable "template_cpu_count" {
  description = "CPU cores for the approved template."
  type        = number
  default     = 2
}

variable "template_memory_mb" {
  description = "Memory in MB for the approved template."
  type        = number
  default     = 512
}

variable "sandbox_timeout_seconds" {
  description = "Controlled sandbox time to live in seconds."
  type        = number
  default     = 300
}

variable "allowed_egress_domains" {
  description = "Domains the controlled sandbox may reach."
  type        = set(string)
  default     = ["api.example.com"]
}

variable "egress_proxy_address" {
  description = "SOCKS5 proxy address in host:port format."
  type        = string
}

variable "egress_proxy_username" {
  description = "SOCKS5 proxy username."
  type        = string
  default     = null
}

variable "egress_proxy_password" {
  description = "SOCKS5 proxy password."
  type        = string
  sensitive   = true
  default     = null
}

variable "policy_header_value" {
  description = "Value injected into outbound requests as X-E2B-Policy."
  type        = string
  default     = "terraform-managed"
}

variable "lifecycle_webhook_url" {
  description = "HTTPS endpoint that receives E2B lifecycle audit events."
  type        = string
}

variable "lifecycle_webhook_signature_secret" {
  description = "Secret E2B uses to sign lifecycle webhook payloads."
  type        = string
  sensitive   = true
}

