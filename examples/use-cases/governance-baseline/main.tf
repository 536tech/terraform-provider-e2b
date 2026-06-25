resource "e2b_template" "approved" {
  name       = var.template_name
  from_image = var.template_image
  start_cmd  = "sh -c \"sleep 3600\""
  ready_cmd  = "true"
  cpu_count  = var.template_cpu_count
  memory_mb  = var.template_memory_mb
  public     = false
}

resource "e2b_lifecycle_webhook" "audit" {
  name             = "terraform-governance-audit"
  url              = var.lifecycle_webhook_url
  enabled          = true
  signature_secret = var.lifecycle_webhook_signature_secret
  events = [
    "sandbox.started",
    "sandbox.paused",
    "sandbox.killed",
  ]
}

resource "e2b_sandbox" "controlled" {
  template_id = e2b_template.approved.id
  timeout     = var.sandbox_timeout_seconds
  auto_resume = true

  allow_internet_access        = true
  network_allow_public_traffic = false
  network_allow_out            = var.allowed_egress_domains
  network_deny_out             = ["ALL_TRAFFIC"]

  network_egress_proxy = {
    address  = var.egress_proxy_address
    username = var.egress_proxy_username
    password = var.egress_proxy_password
  }

  network_rules = {
    for domain in var.allowed_egress_domains : domain => [
      {
        headers = {
          "X-E2B-Policy" = var.policy_header_value
        }
      }
    ]
  }

  metadata = {
    managed_by      = "terraform"
    governance_tier = "restricted"
    audit_webhook   = e2b_lifecycle_webhook.audit.id
  }
}

data "e2b_sandbox" "evidence" {
  id = e2b_sandbox.controlled.id
}

data "e2b_team_metric_max" "concurrent_sandboxes" {
  team_id = var.team_id
  metric  = "concurrent_sandboxes"
}
