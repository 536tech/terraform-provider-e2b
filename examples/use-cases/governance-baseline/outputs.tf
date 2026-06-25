output "sandbox_id" {
  description = "Controlled sandbox ID."
  value       = e2b_sandbox.controlled.id
}

output "template_id" {
  description = "Approved private template ID."
  value       = e2b_template.approved.id
}

output "template_public" {
  description = "Whether the approved template is public."
  value       = e2b_template.approved.public
}

output "sandbox_governance_evidence" {
  description = "Selected sandbox controls that can be consumed by external policy checks."
  value = {
    allow_internet_access        = data.e2b_sandbox.evidence.allow_internet_access
    network_allow_public_traffic = data.e2b_sandbox.evidence.network_allow_public_traffic
    network_allow_out            = data.e2b_sandbox.evidence.network_allow_out
    network_deny_out             = data.e2b_sandbox.evidence.network_deny_out
    metadata                     = data.e2b_sandbox.evidence.metadata
  }
}

output "max_concurrent_sandboxes" {
  description = "Maximum observed concurrent sandboxes for the team over E2B's default metric interval."
  value       = data.e2b_team_metric_max.concurrent_sandboxes.value
}
