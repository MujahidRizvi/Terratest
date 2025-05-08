variable "name" {
  description = "Name of the API Management instance"
  type        = string
}

variable "location" {
  description = "Azure region for APIM"
  type        = string
}

variable "resource_group_name" {
  description = "Resource group where APIM will be deployed"
  type        = string
}

variable "sku_name" {
  description = "SKU for APIM (Developer, Basic, Standard_v2, Premium_v2)"
  type        = string
}

variable "publisher_name" {
  description = "Publisher name for APIM"
  type        = string
}

variable "publisher_email" {
  description = "Publisher email for APIM"
  type        = string
}

variable "outbound_subnet_id" {
  description = "Subnet ID for virtual network integration"
  type        = string
}

variable "public_network_access_enabled" {
  description = "Disable or enable public access"
  type        = bool
}

variable "virtual_network_type" {
  description = "internal when injection vnet"
  type = string
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}

variable "project" {
  description = "Project tag value"
  type        = string
}

variable "dns_record_prefixes" {
  type    = list(string)
  default = ["management", "developer", "portal"]
}