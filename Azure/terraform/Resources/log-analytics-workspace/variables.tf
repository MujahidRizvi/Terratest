variable "name" {
  description = "Name of the Log Analytics Workspace"
  type        = string
}

variable "location" {
  description = "Azure region for the Log Analytics Workspace"
  type        = string
}

variable "resource_group_name" {
  description = "Resource Group for the Log Analytics Workspace"
  type        = string
}
