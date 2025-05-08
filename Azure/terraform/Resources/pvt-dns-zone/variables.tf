variable "dns_zones" {
  description = "List of private DNS zones to create"
  type        = list(string)
  default = [
    "privatelink.azurewebsites.net",
    "privatelink.azurecr.io",
    "privatelink.azure-api.net",
    "privatelink.servicebus.windows.net",
    "azure-api.net",
    "privatelink.aks.azure.com",
    "privatelink.logic.azure.com",
    "privatelink.vaultcore.azure.net",     
    "privatelink.blob.core.windows.net",   
    "privatelink.eventgrid.azure.net",     
    "privatelink.servicebus.windows.net",  
    "privatelink.monitor.azure.com",       
    "privatelink.grafana.azure.com"
  ]
}

variable "virtual_network_ids" {
  description = "List of VNet IDs to link with private DNS zones"
  type        = list(string)
}

variable "resource_group_name" {
  description = "Resource group where DNS zones and links will be created"
  type        = string
}

variable "project" {
  description = "Project name for tagging"
  type        = string
  default     = "API Ecosystem"
}

variable "tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}
