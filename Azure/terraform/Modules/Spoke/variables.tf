variable "project_name" {
  type        = string
  description = "Project name prefix"
}

variable "environment" {
  type        = string
  description = "Environment (dev, qa, prod)"
}

variable "region" {
  type        = string
  description = "Azure region"
}

variable "region_short" {
  type        = string
  description = "Short version of Azure region"
}

variable "spoke_vnet_cidr" {
  type        = string
  description = "CIDR block for the Spoke VNet"
}

variable "bastion_subnet_cidr" {
  type        = string
  description = "CIDR block for the Bastion subnet"
}

variable "exp_subnet_cidr" {
  type        = string
  description = "CIDR block for the Experience (Exp) subnet"
}

variable "proc_subnet_cidr" {
  type        = string
  description = "CIDR block for the Processing subnet"
}

variable "sys_subnet_cidr" {
  type        = string
  description = "CIDR block for the System subnet"
}

variable "procfapp_subnet_cidr" {
  type        = string
  description = "CIDR block for the Processing subnet"
}

variable "sysfapp_subnet_cidr" {
  type        = string
  description = "CIDR block for the Processing subnet"
}

variable "srcs_subnet_cidr" {
  type        = string
  description = "CIDR block for the Processing subnet"
}
variable "epp_subnet_cidr" {
  type        = string
  description = "CIDR block for the Event Processing Platform subnet"
}



variable "default_nsg_rules" {
  description = "NSG rules for all subnets"
  type = map(object({
    priority                   = number
    direction                  = string
    access                     = string
    protocol                   = string
    source_port_range          = string
    destination_port_range     = string
    source_address_prefix      = string
    destination_address_prefix = string
  }))
}

variable "exp_apim_nsg_rules" {
  description = "NSG rules for all subnets"
  type = map(object({
    priority                   = number
    direction                  = string
    access                     = string
    protocol                   = string
    source_port_range          = string
    destination_port_range     = string
    source_address_prefix      = string
    destination_address_prefix = string
  }))
}

variable "bst_vm_nsg_rules" {
  description = "NSG rules for all subnets"
  type = map(object({
    priority                   = number
    direction                  = string
    access                     = string
    protocol                   = string
    source_port_range          = string
    destination_port_range     = string
    source_address_prefix      = string
    destination_address_prefix = string
  }))
}

variable "tags" {
  type        = map(string)
  description = "Common tags applied to resources"
}

variable "const_snet_delegation_service_name_serverfarms" {
  description = "Name of service point for which subnet delegation happens"
  default = "Microsoft.Web/serverFarms"
  type        = string
}

variable "const_snet_delegation_action" {
  description = "Action to be taken on the subnet by the Azure service"
  default = "Microsoft.Network/virtualNetworks/subnets/action"
  type        = string
}