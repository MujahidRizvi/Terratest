variable "bastion_password" {
  description = "Admin password for the Bastion Windows VM"
  type        = string
  sensitive   = true
}

variable "default_nsg_rules" {
  description = "Standard NSG rules to apply to each subnet's NSG"
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
  description = "Standard NSG rules to apply to each subnet's NSG"
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
  description = "Standard NSG rules to apply to each subnet's NSG"
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

variable "apim_sku" {
  description = "APIM SKU tier"
  type        = string
  default     = "Developer_1"
}

variable "publisher_name" {
  description = "Publisher name for APIM"
  type        = string
  default = "AG Investments"
}

variable "publisher_email" {
  description = "Publisher email for APIM"
  type        = string
  default = "abdel.rehman@al-ghurair.com"
}