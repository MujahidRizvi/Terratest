variable "name" {}
variable "resource_group_name" {}
variable "virtual_network_name" {}
variable "address_prefix" {}
variable "location" {}
variable "tags" {
  type    = map(string)
  default = {}
}

variable "service_endpoints" {
  type    = list(string)
  default = [
    "Microsoft.Storage",
    "Microsoft.Sql",
    "Microsoft.AzureActiveDirectory",
    "Microsoft.AzureCosmosDB",
    "Microsoft.Web",
    "Microsoft.KeyVault",
    "Microsoft.EventHub",
    "Microsoft.ServiceBus",
    "Microsoft.ContainerRegistry",
    "Microsoft.CognitiveServices"
  ]
}

variable "network_security_group_id" {
  description = "ID of the NSG to associate with this subnet"
  type        = string
}
