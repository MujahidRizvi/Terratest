variable "environment" {
  description = "Deployment environment (dev, uat, prod)"
  type        = string
}

variable "region" {
  description = "Azure region to deploy resources"
  type        = string
  default     = "uaenorth"
}

variable "project_name" {
  description = "Project name prefix used in resource naming"
  type        = string
  default     = "agida"
}

variable "resource_group_name" {
  description = "Existing Resource Group name to deploy the VNet and Subnet"
  type        = string
}

variable "vnet_name" {
  description = "Name of the Virtual Network (max 64 characters)"
  type        = string
}

variable "vnet_cidr" {
  description = "CIDR block for the Virtual Network (e.g. 10.0.0.0/16)"
  type        = string
}

variable "subnet1_name" {
  description = "Name of the Subnet (max 80 characters)"
  type        = string
}

variable "subnet1_cidr" {
  description = "CIDR block for the Subnet (e.g. 10.0.1.0/24)"
  type        = string
}

variable "subnet4_name" {
  description = "Name of the Subnet (max 80 characters)"
  type        = string
}

variable "subnet4_cidr" {
  description = "CIDR block for the Subnet (e.g. 10.0.4.0/24)"
  type        = string
}

variable "appgw_subnet_cidr" {
  description = "CIDR block for Application Gateway subnet"
  type        = string
}

variable "vm_name" {
  description = "Name"
  type        = string
}

variable "vm_admin_username" {
  description = "Username"
  type        = string
}

variable "vm_admin_password" {
  description = "Password"
  type        = string
}

variable "vm_resource_group_name" {
  description = "Resource Group for VM"
  type        = string
  default     = "agida-main-uaen-adotf-rg"
}

# App Gateway Variables
variable "appgw_sku_name" {
  type    = string
  default = "Standard_v2"
}

variable "appgw_sku_tier" {
  type    = string
  default = "Standard_v2"
}

variable "appgw_capacity" {
  type    = number
  default = 2
}

variable "appgw_frontend_ip_name" {
  type    = string
  default = "appgw-frontend-ip"
}

variable "appgw_frontend_port_name" {
  type    = string
  default = "appgw-frontend-port"
}

variable "appgw_frontend_port" {
  type    = number
  default = 80
}

variable "appgw_listener_name" {
  type    = string
  default = "appgw-http-listener"
}

variable "appgw_listener_protocol" {
  type    = string
  default = "Http"
}

variable "appgw_rule_name" {
  type    = string
  default = "appgw-routing-rule"
}

variable "appgw_rule_type" {
  type    = string
  default = "Basic"
}

variable "appgw_backend_pool_name" {
  type    = string
  default = "appgw-backend-pool"
}

variable "appgw_http_settings_name" {
  type    = string
  default = "appgw-http-settings"
}

variable "appgw_cookie_affinity" {
  type    = string
  default = "Disabled"
}

variable "appgw_path" {
  type    = string
  default = "/"
}

variable "appgw_backend_port" {
  type    = number
  default = 80
}

variable "appgw_backend_protocol" {
  type    = string
  default = "Http"
}

variable "appgw_request_timeout" {
  type    = number
  default = 20
}

variable "appgw_rule_priority" {
  type        = number
  description = "Priority value for the Application Gateway routing rule"
}

variable "private_dns_zones" {
  description = "List of private DNS zones to be created and linked to the hub VNet"
  type        = list(string)
}
