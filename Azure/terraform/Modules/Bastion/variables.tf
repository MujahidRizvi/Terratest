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

variable "bastion_subnet_id" {
  description = "Subnet ID where the Bastion VM will be deployed"
  type        = string
}

variable "vm_size" {
  description = "Size of the Bastion VM"
  type        = string
  default     = "Standard_B2s"
}

variable "admin_username" {
  description = "Admin username for the Windows VM"
  type        = string
}

variable "admin_password" {
  description = "Admin password for the Windows VM"
  type        = string
  sensitive   = true
}

variable "create_public_ip" {
  description = "Should a public IP be attached to the VM"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags for the VM and resource group"
  type        = map(string)
  default     = {}
}

variable "project" {
  description = "Project name for tagging"
  type        = string
  default     = "API Ecosystem"
}
