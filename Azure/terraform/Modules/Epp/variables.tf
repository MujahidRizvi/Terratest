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
variable "epp_subnet_id" {
  description = "Subnet ID to be used for EPP"
  type        = string
}

variable "project" {
  description = "Project tag value"
  type        = string
  default     = "API Ecosystem"
}

variable "sku" {
  description = "sku"
  type        = string
}

variable "partition_count" {
  description = "EventHub Partition count =1,2....32"
  type        = string
}

variable "message_retention" {
  description = "No of days the messages can be retained - 1--7"
  type        = string
}


variable "capacity" {
  description = "Pricing Tier SKU for EventHub Name Space (Premium, standard)"
  type        = string
}
  
variable "private_ehdns_zones" {
  description = "IDs of the Private DNS Zones where the PEP should register (e.g., privatelink.blob.core.windows.net)"
  type        = list(string)
  default = [
    "privatelink.servicebus.windows.net",
  ]
}

variable "const_subresource_name_eh" {
  description = "Resource name for 'eventhib'"
  default     = "namespace"
  type        = string
}

variable "tags" {
  description = "Tags for resources"
  type        = map(string)
  default     = {}
}