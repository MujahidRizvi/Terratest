# Azure Storage Account
resource "azurerm_storage_account" "storage" {
  name                     = var.storage_account_name
  resource_group_name      = var.storage_resource_group_name
  location                 = var.location
  account_tier             = var.storage_sku
  account_replication_type = var.storage_replication_type
  public_network_access_enabled = true
  shared_access_key_enabled     = true
  allow_nested_items_to_be_public = false
  min_tls_version = "TLS1_2"
  network_rules {
  default_action = "Deny"
  bypass         = ["AzureServices"]
  virtual_network_subnet_ids = [var.function_subnet_id]
  }

  tags = {
    Name        = var.storage_account_name
    Project     = "API Ecosystem"
    Owner       = "CloudTeam"
  }

}

# Log Analytics Workspace
module "log_analytics_workspace" {
  source              = "../log-analytics-workspace"
  name                = "${var.function_app_name}-law"
  location            = var.location
  resource_group_name = var.function_resource_group_name
}

resource "azurerm_application_insights" "this" {
  name                                = "${var.function_app_name}-appi"
  location                            = var.location
  resource_group_name                 = var.function_resource_group_name
  application_type                    = "web"
  workspace_id                        = module.log_analytics_workspace.log_analytics_workspace_id
  retention_in_days                   = 30
  daily_data_cap_in_gb                = 100
  disable_ip_masking                  = false
  force_customer_storage_for_profiler = false
  internet_ingestion_enabled          = true
  internet_query_enabled              = true
  local_authentication_disabled       = false

  tags = {
    Name        = "${var.function_app_name}-appi"
    Project     = "API Ecosystem"
    Owner       = "CloudTeam"
  }

  depends_on = [azurerm_storage_account.storage]
}

# Azure App Service Plan
resource "azurerm_service_plan" "asp" {
  
  name                = var.app_service_plan_name
  location            = var.location
  resource_group_name = var.function_resource_group_name
  sku_name            = var.app_service_plan_sku_name
  os_type             = var.app_service_plan_os_type

  tags = {
    Name        = var.app_service_plan_name
    Project     = "API Ecosystem"
    Owner       = "CloudTeam"
  }


  depends_on = [ azurerm_storage_account.storage ]
}

# Azure Function App
resource "azurerm_windows_function_app" "function" {
  name                          = var.function_app_name
  location                      = var.location
  resource_group_name           = var.function_resource_group_name
  service_plan_id               = azurerm_service_plan.asp.id
  storage_account_name          = azurerm_storage_account.storage.name
  storage_account_access_key    = azurerm_storage_account.storage.primary_access_key
  functions_extension_version   = var.function_runtime_version
  virtual_network_subnet_id     = var.function_subnet_id
  public_network_access_enabled = var.public_network_access_enabled

  identity {
    type = "SystemAssigned"
  }

  site_config {
    always_on             = true
    ftps_state            = "FtpsOnly"
    http2_enabled         = true
    minimum_tls_version   = "1.2"
    use_32_bit_worker     = false
    scm_minimum_tls_version = "1.2"

    application_stack {
      dotnet_version               = var.function_dotnet_version
      use_dotnet_isolated_runtime = true
    }
  }

  app_settings = {
    "APPINSIGHTS_INSTRUMENTATIONKEY"          = azurerm_application_insights.this.instrumentation_key
    "APPLICATIONINSIGHTS_CONNECTION_STRING"   = azurerm_application_insights.this.connection_string
    "APPLICATIONINSIGHTS_ROLE_NAME"           = var.function_app_name
  }

  tags = {
    Name        = var.function_app_name
    Project     = "API Ecosystem"
    Owner       = "CloudTeam"
  }
}