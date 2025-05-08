project_name             = "agida"
environment              = "main"
region                   = "uaenorth"
resource_group_name      = "agida-main-uaen-dahub-rg"
vnet_name                = "dahub"
vnet_cidr                = "10.40.0.0/16"

subnet1_name             = "dadev"
subnet1_cidr             = "10.40.10.0/24"

subnet4_name             = "adopa"
subnet4_cidr             = "10.40.40.0/24"
appgw_subnet_cidr        = "10.40.50.0/24"

vm_name                  = "adopa"
vm_admin_username        = "adopa"
vm_admin_password        = "AgiDaP00l@ge98"
vm_resource_group_name   = "agida-main-uaen-adotf-rg"

# App Gateway overrides
appgw_sku_name           = "Standard_v2"
appgw_sku_tier           = "Standard_v2"
appgw_capacity           = 1
appgw_frontend_ip_name   = "fepip"
appgw_frontend_port_name = "feport"
appgw_frontend_port      = 80
appgw_listener_name      = "lstnr"
appgw_listener_protocol  = "Http"
appgw_rule_name          = "rtrule"
appgw_rule_type          = "Basic"
appgw_backend_pool_name  = "bepool"
appgw_http_settings_name = "http-settings"
appgw_cookie_affinity    = "Disabled"
appgw_path               = "/"
appgw_backend_port       = 80
appgw_backend_protocol   = "Http"
appgw_request_timeout    = 20
appgw_rule_priority      = 100

private_dns_zones = [
  "privatelink.blob.core.windows.net",
  "privatelink.file.core.windows.net",
  "privatelink.queue.core.windows.net",
  "privatelink.table.core.windows.net",
  "privatelink.web.core.windows.net",
  "privatelink.azurewebsites.net",
  "privatelink.database.windows.net",
  "privatelink.documents.azure.com",
  "privatelink.vaultcore.azure.net",
  "privatelink.servicebus.windows.net",
  "azure-api.net"
]
