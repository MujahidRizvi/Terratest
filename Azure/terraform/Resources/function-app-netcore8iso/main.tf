# Storage Account
resource "azurerm_storage_account" "storage" {
  name                     = var.storage_account_name
  resource_group_name      = var.storage_resource_group_name
  location                 = var.location
  account_tier             = var.storage_sku
  account_replication_type = var.storage_replication_type
}

# App Service Plan
resource "azurerm_service_plan" "asp" {
  
  name                = var.app_service_plan_name
  location            = var.location
  resource_group_name = var.function_resource_group_name
  sku_name            = var.app_service_plan_sku_name
  os_type             = var.app_service_plan_os_type

  depends_on = [ azurerm_storage_account.storage ]
}

# Function App
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

  site_config {
    always_on = true
    ftps_state              = "FtpsOnly"
    http2_enabled           = true
    minimum_tls_version     = "1.2"
    scm_minimum_tls_version = "1.2"
    use_32_bit_worker       = false
    application_stack{
      dotnet_version = var.function_dotnet_version
      use_dotnet_isolated_runtime = true
    }
  }

  identity {
    type = "SystemAssigned"
  }

  depends_on = [ azurerm_service_plan.asp ]
}