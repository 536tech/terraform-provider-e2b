output "container_registry_login_server" {
  description = "Azure Container Registry login server for template images."
  value       = azurerm_container_registry.images.login_server
}

output "key_vault_uri" {
  description = "Key Vault URI for operator-managed E2B deployment secrets."
  value       = azurerm_key_vault.main.vault_uri
}

output "log_analytics_workspace_id" {
  description = "Log Analytics workspace ID for deployment telemetry."
  value       = azurerm_log_analytics_workspace.main.id
}

output "managed_identity_client_id" {
  description = "Client ID for the orchestrator managed identity."
  value       = azurerm_user_assigned_identity.orchestrator.client_id
}

output "resource_group_name" {
  description = "Azure resource group name."
  value       = azurerm_resource_group.main.name
}

output "storage_account_name" {
  description = "Storage account name for E2B templates, snapshots, and logs."
  value       = azurerm_storage_account.artifacts.name
}

output "subnet_ids" {
  description = "Azure subnet IDs keyed by workload role."
  value       = { for name, subnet in azurerm_subnet.workload : name => subnet.id }
}

output "virtual_network_id" {
  description = "Azure virtual network ID."
  value       = azurerm_virtual_network.main.id
}
