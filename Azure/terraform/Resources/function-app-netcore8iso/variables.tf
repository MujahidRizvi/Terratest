
variable "function_resource_group_name" {
  description = "The name of Azure function resource group"
  type        = string
}

variable "storage_resource_group_name" {
  description = "The name of Storage resource group"
  type        = string
}

variable "location" {
  description = "Azure location for the resources"
  type        = string
  default     = "East US"
}

variable "storage_account_name" {
  description = "Storage account name for the Function App"
  type        = string
}

variable "app_service_plan_name" {
  description = "Name of the App Service Plan"
  type        = string
}

variable "function_app_name" {
  description = "Name of the Function App"
  type        = string
}

variable "storage_sku" {
  description = "SKU of the straoge account"
  type        = string
}

variable "storage_replication_type" {
  description = "Replication type of the storage account"
  type        = string
}

variable "app_service_plan_os_type" {
  description = "OS type of the App Service Plan"
  type        = string
}

variable "app_service_plan_sku_name" {
  description = "SKU name of the App Service Plan"
  type        = string
}

variable "function_runtime_version" {
  description = "Runtime version of the Function App"
  type        = string
}

variable "function_dotnet_version" {
  description = "Dotnet runtime version of the Function App"
  type        = string
}

variable "function_subnet_id" {
  description = "Subnet ID for the Function App"
  type        = string
}

variable "public_network_access_enabled" {
  description = "Enable or Disable Public Access to FApp"
  type = bool
  default = false
}