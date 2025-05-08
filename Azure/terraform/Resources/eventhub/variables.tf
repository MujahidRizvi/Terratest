variable "resource_group_name" {
  description = "Resource group where EPP will be deployed"
  type        = string
}

variable "eventhub_name" {
  description = "Name of the Event Hub"
  type        = string
}

variable "eventhub_namespace_name" {
  description = "Name of the Event Hub name space"
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

variable "location" {
  description = "Azure location for EventHub Namespace region"
  type        = string
}

variable "sku" {
  description = "Pricing Tier SKU for EventHub Name Space (Premium, standard)"
  type        = string
}

variable "capacity" {
  description = "Pricing Tier SKU for EventHub Name Space (Premium, standard)"
  type        = string
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}