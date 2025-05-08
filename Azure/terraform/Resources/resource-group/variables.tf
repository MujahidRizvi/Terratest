variable "name" {
  description = "Name of the resource group"
  type        = string
}

variable "location" {
  description = "Azure region where the resource group will be created"
  type        = string
}

variable "project" {
  description = "Project name for tagging"
  type        = string
  default     = "API Ecosystem"
}

variable "tags" {
  description = "Additional tags to apply to the resource group"
  type        = map(string)
  default     = {}
}
