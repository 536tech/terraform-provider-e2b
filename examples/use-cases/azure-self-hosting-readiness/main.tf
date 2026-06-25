data "azurerm_client_config" "current" {}

resource "random_string" "suffix" {
  length  = 6
  lower   = true
  numeric = true
  special = false
  upper   = false
}

resource "azurerm_resource_group" "main" {
  location = var.location
  name     = "rg-${var.project_name}-${var.environment}"
  tags     = local.tags
}

resource "azurerm_virtual_network" "main" {
  address_space       = var.address_space
  location            = azurerm_resource_group.main.location
  name                = "vnet-${var.project_name}-${var.environment}"
  resource_group_name = azurerm_resource_group.main.name
  tags                = local.tags
}

resource "azurerm_subnet" "workload" {
  for_each = var.subnet_prefixes

  address_prefixes     = [each.value]
  name                 = "snet-${each.key}"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
}

resource "azurerm_network_security_group" "workload" {
  location            = azurerm_resource_group.main.location
  name                = "nsg-${var.project_name}-${var.environment}"
  resource_group_name = azurerm_resource_group.main.name
  tags                = local.tags
}

resource "azurerm_subnet_network_security_group_association" "workload" {
  for_each = azurerm_subnet.workload

  network_security_group_id = azurerm_network_security_group.workload.id
  subnet_id                 = each.value.id
}

resource "azurerm_storage_account" "artifacts" {
  account_replication_type        = "LRS"
  account_tier                    = "Standard"
  allow_nested_items_to_be_public = false
  location                        = azurerm_resource_group.main.location
  min_tls_version                 = "TLS1_2"
  name                            = substr("${local.name_prefix}${random_string.suffix.result}st", 0, 24)
  resource_group_name             = azurerm_resource_group.main.name
  shared_access_key_enabled       = false
  tags                            = local.tags
}

resource "azurerm_container_registry" "images" {
  admin_enabled       = false
  location            = azurerm_resource_group.main.location
  name                = substr("${local.name_prefix}${random_string.suffix.result}acr", 0, 50)
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "Basic"
  tags                = local.tags
}

resource "azurerm_log_analytics_workspace" "main" {
  location            = azurerm_resource_group.main.location
  name                = "log-${var.project_name}-${var.environment}"
  resource_group_name = azurerm_resource_group.main.name
  retention_in_days   = 30
  sku                 = "PerGB2018"
  tags                = local.tags
}

resource "azurerm_key_vault" "main" {
  rbac_authorization_enabled = true
  location                   = azurerm_resource_group.main.location
  name                       = substr("${local.name_prefix}${random_string.suffix.result}kv", 0, 24)
  purge_protection_enabled   = true
  resource_group_name        = azurerm_resource_group.main.name
  sku_name                   = "standard"
  soft_delete_retention_days = 7
  tags                       = local.tags
  tenant_id                  = data.azurerm_client_config.current.tenant_id
}

resource "azurerm_user_assigned_identity" "orchestrator" {
  location            = azurerm_resource_group.main.location
  name                = "id-${var.project_name}-${var.environment}-orchestrator"
  resource_group_name = azurerm_resource_group.main.name
  tags                = local.tags
}

resource "azurerm_role_assignment" "storage_blob_data_contributor" {
  principal_id         = azurerm_user_assigned_identity.orchestrator.principal_id
  scope                = azurerm_storage_account.artifacts.id
  role_definition_name = "Storage Blob Data Contributor"
}

resource "azurerm_role_assignment" "acr_pull" {
  principal_id         = azurerm_user_assigned_identity.orchestrator.principal_id
  scope                = azurerm_container_registry.images.id
  role_definition_name = "AcrPull"
}

resource "azurerm_role_assignment" "key_vault_secrets_user" {
  principal_id         = azurerm_user_assigned_identity.orchestrator.principal_id
  scope                = azurerm_key_vault.main.id
  role_definition_name = "Key Vault Secrets User"
}
