variable "name" {
  description = "Name of the virtual network"
  type        = string
}

variable "location" {
  description = "Azure region where the virtual network will be deployed"
  type        = string
}

variable "resource_group_name" {
  description = "Resource group name for the virtual network"
  type        = string
}

variable "address_space" {
  description = "CIDR block for the virtual network"
  type        = string
}

variable "project" {
  description = "Project name for tagging"
  type        = string
  default     = "API Ecosystem"
}

variable "tags" {
  description = "Additional tags to apply"
  type        = map(string)
  default     = {}
}

variable "hub_vnet_name" {
  description = "Name of the hub virtual network to peer with"
  type        = string
  default     = "agida-main-uaenorth-dahub-vnet"
}

variable "hub_resource_group" {
  description = "Resource group of the hub virtual network"
  type        = string
  default     = "agida-main-uaen-dahub-rg"
}

variable "private_dns_zones" {
  description = "List of private DNS zones to link to this VNet"
  type        = list(string)
  default = [
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
}
