variable "name" {
  description = "Name of the private endpoint"
  type        = string
}

variable "location" {
  description = "Azure region for the private endpoint"
  type        = string
}

variable "resource_group_name" {
  description = "Resource group in which to create the private endpoint"
  type        = string
}

variable "subnet_id" {
  description = "Subnet ID where the private endpoint will be deployed"
  type        = string
}

variable "target_resource_id" {
  description = "ID of the resource to link via private endpoint (e.g., APIM, Storage Account)"
  type        = string
}

variable "subresource_names" {
  description = "List of subresources to connect to (e.g., ['gateway'], ['blob'], etc.)"
  type        = list(string)
}

variable "is_manual_connection" {
  description = "Whether the connection requires manual approval"
  type        = bool
  default     = false
}

variable "custom_nic_name" {
  description = "Custom name for the network interface"
  type        = string
  default     = null
}

variable "project" {
  description = "Project name for tagging"
  type        = string
  default     = "API Ecosystem"
}

variable "tags" {
  description = "Additional tags to apply to the resource"
  type        = map(string)
  default     = {}
}

variable "private_dns_zones" {
  description = "IDs of the Private DNS Zones where the PEP should register (e.g., privatelink.blob.core.windows.net)"
  type        = list(string)
}