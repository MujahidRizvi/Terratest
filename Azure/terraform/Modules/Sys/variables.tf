variable "project_name" {
    description = "Project name prefix"
    type        = string
}

variable "environment" {
    description = "Environment name (dev, qa, prod)"
    type        = string
}

variable "region" {
    description = "Azure region"
    type        = string
}

variable "region_short" {
    description = "Short version of the Azure region"
    type        = string
}

/*variable "resource_type" {
    description = "Resource type (e.g., 'function', 'apim')"
    type        = string
}*/

variable "const_app_service_plan_sku_name_b1" {
    description = "App Service Plan SKU Tier for Basic 1"
    default     = "B1"
    type        = string
}

variable "const_storage_sku_Standard" {
    description = "Storage SKU 'Standard'"
    default     = "Standard"
    type        = string
}

variable "const_storage_replication_type_LRS" {
    description = "Storage Replication Type LRS"
    default     = "LRS"
    type        = string
}

variable "const_function_runtime_version_4" {
    description = "Function Runtime Version 4"
    default     = "~4"
    type        = string
}

variable "const_function_dotnet_version_v8" {
    description = "Function .Net Runtime Version 8"
    default     = "v8.0"
    type        = string
}

variable "const_app_service_plan_os_type_windows" {
    description = "App Service Plan OS Type - Windows"
    default     = "Windows"
    type        = string
}

variable "const_resource_name_sites" {
    description = "Resource name for 'sites'"
    default     = "sites"
    type        = string
}

variable "sysfapp_subnet_cidr" {
    description = "Subnet ID for the Function App"
    type        = string
}

variable "sysfapp_pep_subnet_cidr" {
    description = "Subnet ID for the Function App Pep"
    type        = string
}

variable "private_webdns_zones" {
  description = "IDs of the Private DNS Zones where the PEP should register (e.g., privatelink.blob.core.windows.net)"
  type        = list(string)
  default = [
    "privatelink.azurewebsites.net",
  ]
}