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

variable "apim_sku" {
  description = "SKU name for APIM"
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

variable "exp_subnet_id" {
  description = "Subnet ID to be used for APIM"
  type        = string
}

variable "tags" {
  description = "Tags for resources"
  type        = map(string)
  default     = {}
}

variable "project" {
  description = "Project tag value"
  type        = string
  default     = "API Ecosystem"
}

variable "public_network_access_enabled" {
  description = "True for Public, False for Private"
  type        = bool
  default     = true
}

variable "virtual_network_type" {
  description = "Internal for Vnet Injection"
  type        = string
  default     = "Internal"
}