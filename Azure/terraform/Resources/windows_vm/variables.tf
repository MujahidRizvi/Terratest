variable "name" {
  description = "Name of the VM"
  type        = string
}

variable "location" {
  description = "Region to deploy VM in"
  type        = string
}

variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
}

variable "subnet_id" {
  description = "Subnet ID to place the NIC in"
  type        = string
}

variable "vm_size" {
  description = "Azure VM size (e.g., Standard_B2s)"
  type        = string
  default     = "Standard_B2s"
}

variable "admin_username" {
  description = "Admin username for the VM"
  type        = string
}

variable "admin_password" {
  description = "Admin password for the VM"
  type        = string
  sensitive   = true
}

variable "os_disk_type" {
  description = "Storage type for OS disk"
  type        = string
  default     = "Standard_LRS"
}

variable "create_public_ip" {
  description = "Whether to create and attach a public IP"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default     = {}
}
