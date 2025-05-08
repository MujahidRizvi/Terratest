resource "azurerm_api_management" "this" {
  name                          = var.name
  location                      = var.location
  resource_group_name           = var.resource_group_name
  publisher_name                = var.publisher_name
  publisher_email               = var.publisher_email
  sku_name                      = var.sku_name
  public_network_access_enabled = var.public_network_access_enabled
  virtual_network_type          = var.virtual_network_type

  identity {
    type = "SystemAssigned"
  }

  virtual_network_configuration {
    subnet_id = var.outbound_subnet_id
  }

  tags = merge({
    Name    = var.name,
    Project = var.project
  }, var.tags)
}

# Log Analytics Workspace
module "log_analytics_workspace" {
  source              = "../log-analytics-workspace"
  name                = "${var.name}-law"
  location            = var.location
  resource_group_name = var.resource_group_name

  depends_on = [azurerm_api_management.this]
}

resource "azurerm_application_insights" "this" {
  name                                = "${var.name}-appi"
  location                            = var.location
  resource_group_name                 = var.resource_group_name
  application_type                    = "other"
  workspace_id                        = module.log_analytics_workspace.log_analytics_workspace_id
  retention_in_days                   = 30
  daily_data_cap_in_gb                = 100
  disable_ip_masking                  = false
  force_customer_storage_for_profiler = false
  internet_ingestion_enabled          = true
  internet_query_enabled              = true
  local_authentication_disabled       = false

  tags = {
    Name        = "${var.name}-appi"
    Project     = "API Ecosystem"
    Owner       = "CloudTeam"
  }

  depends_on = [module.log_analytics_workspace]
}

resource "azurerm_api_management_logger" "appi_logger" {
  name                = "${var.name}-appi-logger"
  api_management_name = azurerm_api_management.this.name
  resource_group_name = var.resource_group_name

  application_insights {
    instrumentation_key = azurerm_application_insights.this.instrumentation_key
  }

  depends_on = [azurerm_application_insights.this]
}

resource "azurerm_private_dns_a_record" "apim_a_record" {
  name                = var.name
  zone_name           = "azure-api.net"
  resource_group_name = "agida-main-uaen-dahub-rg"
  ttl                 = 60
  records             = [azurerm_api_management.this.private_ip_addresses[0]]

  depends_on = [azurerm_api_management.this]
}

resource "azurerm_private_dns_a_record" "apim_dns_records" {
  for_each            = toset(var.dns_record_prefixes)
  name                = each.value == "" ? var.name : "${var.name}.${each.value}"
  zone_name           = "azure-api.net"
  resource_group_name = "agida-main-uaen-dahub-rg"
  ttl                 = 60
  records             = [azurerm_api_management.this.private_ip_addresses[0]]
  
  depends_on = [azurerm_api_management.this]
}

output "apim_id" {
  value = azurerm_api_management.this.id
}

output "apim_name" {
  value = azurerm_api_management.this.name
}

output "apim_private_ip" {
  value = length(azurerm_api_management.this.private_ip_addresses) > 0 ? azurerm_api_management.this.private_ip_addresses[0] : null
}

output "apim_fqdn" {
  value = "${var.name}.azure-api.net"
}