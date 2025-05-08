resource "azurerm_resource_group" "spoke" {
  name     = "${var.project_name}-${var.environment}-${var.region_short}-spk-rg"
  location = var.region

  tags = {
    Project = var.tags["project"]
  }
}

module "spoke_vnet" {
  source              = "../../Resources/vnet"
  name                = "${var.project_name}-${var.environment}-${var.region_short}-spoke-vnet"
  location            = var.region
  resource_group_name = azurerm_resource_group.spoke.name
  address_space       = var.spoke_vnet_cidr
  tags                = var.tags
}

# BASTION
module "nsg_bastion" {
  source              = "../../Resources/nsg"
  name                = "${var.project_name}-${var.environment}-${var.region_short}-bst-nsg"
  location            = var.region
  resource_group_name = azurerm_resource_group.spoke.name
  tags                = var.tags
  security_rules      = var.bst_vm_nsg_rules
}

module "subnet_bastion" {
  source                    = "../../Resources/subnets"
  name                      = "${var.project_name}-${var.environment}-${var.region_short}-bst-snet"
  resource_group_name       = azurerm_resource_group.spoke.name
  virtual_network_name      = module.spoke_vnet.vnet_name
  address_prefix            = var.bastion_subnet_cidr
  location                  = var.region
  network_security_group_id = module.nsg_bastion.id
  tags                      = var.tags
}

# EXP
module "nsg_exp" {
  source              = "../../Resources/nsg"
  name                = "${var.project_name}-${var.environment}-${var.region_short}-exp-nsg"
  location            = var.region
  resource_group_name = azurerm_resource_group.spoke.name
  tags                = var.tags
  security_rules      = var.exp_apim_nsg_rules
}

module "subnet_exp" {
  source                    = "../../Resources/subnets"
  name                      = "${var.project_name}-${var.environment}-${var.region_short}-exp-snet"
  resource_group_name       = azurerm_resource_group.spoke.name
  virtual_network_name      = module.spoke_vnet.vnet_name
  address_prefix            = var.exp_subnet_cidr
  location                  = var.region
  network_security_group_id = module.nsg_exp.id
  tags                      = var.tags
}

# PROC
module "nsg_proc" {
  source              = "../../Resources/nsg"
  name                = "${var.project_name}-${var.environment}-${var.region_short}-proc-nsg"
  location            = var.region
  resource_group_name = azurerm_resource_group.spoke.name
  tags                = var.tags
  security_rules      = var.default_nsg_rules
}

module "subnet_proc" {
  source                    = "../../Resources/subnets"
  name                      = "${var.project_name}-${var.environment}-${var.region_short}-proc-snet"
  resource_group_name       = azurerm_resource_group.spoke.name
  virtual_network_name      = module.spoke_vnet.vnet_name
  address_prefix            = var.proc_subnet_cidr
  location                  = var.region
  network_security_group_id = module.nsg_proc.id
  tags                      = var.tags
}

module "nsg_procfapp" {
  source              = "../../Resources/nsg"
  name                = "${var.project_name}-${var.environment}-${var.region_short}-procfapp-nsg"
  location            = var.region
  resource_group_name = azurerm_resource_group.spoke.name
  tags                = var.tags
  security_rules      = var.default_nsg_rules
}

module "subnet_procfapp" {
  source                        = "../../Resources/subnet-withdelegation"
  name                          = "${var.project_name}-${var.environment}-${var.region_short}-procfapp-snet"
  resource_group_name           = azurerm_resource_group.spoke.name
  virtual_network_name          = module.spoke_vnet.vnet_name
  address_prefix                = var.procfapp_subnet_cidr
  location                      = var.region
  network_security_group_id     = module.nsg_procfapp.id
  snet_delegation_name          = "${var.project_name}-${var.environment}-${var.region_short}-procfapp-snetdelg"
  snet_delegation_service_name  = var.const_snet_delegation_service_name_serverfarms
  snet_delegation_action        = var.const_snet_delegation_action
  tags                          = var.tags
}

# SYS
module "nsg_sys" {
  source              = "../../Resources/nsg"
  name                = "${var.project_name}-${var.environment}-${var.region_short}-sys-nsg"
  location            = var.region
  resource_group_name = azurerm_resource_group.spoke.name
  tags                = var.tags
  security_rules      = var.default_nsg_rules
}

module "subnet_sys" {
  source                    = "../../Resources/subnets"
  name                      = "${var.project_name}-${var.environment}-${var.region_short}-sys-snet"
  resource_group_name       = azurerm_resource_group.spoke.name
  virtual_network_name      = module.spoke_vnet.vnet_name
  address_prefix            = var.sys_subnet_cidr
  location                  = var.region
  network_security_group_id = module.nsg_sys.id
  tags                      = var.tags
}

module "nsg_sysfapp" {
  source              = "../../Resources/nsg"
  name                = "${var.project_name}-${var.environment}-${var.region_short}-sysfapp-nsg"
  location            = var.region
  resource_group_name = azurerm_resource_group.spoke.name
  tags                = var.tags
  security_rules      = var.default_nsg_rules
}

module "subnet_sysfapp" {
  source                        = "../../Resources/subnet-withdelegation"
  name                          = "${var.project_name}-${var.environment}-${var.region_short}-sysfapp-snet"
  resource_group_name           = azurerm_resource_group.spoke.name
  virtual_network_name          = module.spoke_vnet.vnet_name
  address_prefix                = var.sysfapp_subnet_cidr
  location                      = var.region
  network_security_group_id     = module.nsg_sysfapp.id
  snet_delegation_name          = "${var.project_name}-${var.environment}-${var.region_short}-sysfapp-snetdelg"
  snet_delegation_service_name  = var.const_snet_delegation_service_name_serverfarms
  snet_delegation_action        = var.const_snet_delegation_action
  tags                          = var.tags
}

# Shared Resources
module "nsg_srcs" {
  source              = "../../Resources/nsg"
  name                = "${var.project_name}-${var.environment}-${var.region_short}-srcs-nsg"
  location            = var.region
  resource_group_name = azurerm_resource_group.spoke.name
  tags                = var.tags
  security_rules      = var.default_nsg_rules
}

module "subnet_srcs" {
  source                    = "../../Resources/subnets"
  name                      = "${var.project_name}-${var.environment}-${var.region_short}-srcs-snet"
  resource_group_name       = azurerm_resource_group.spoke.name
  virtual_network_name      = module.spoke_vnet.vnet_name
  address_prefix            = var.srcs_subnet_cidr
  location                  = var.region
  network_security_group_id = module.nsg_srcs.id
  tags                      = var.tags
}

# EPP
module "nsg_epp" {
  source              = "../../Resources/nsg"
  name                = "${var.project_name}-${var.environment}-${var.region_short}-epp-nsg"
  location            = var.region
  resource_group_name = azurerm_resource_group.spoke.name
  tags                = var.tags
  security_rules      = var.default_nsg_rules
}

module "subnet_epp" {
  source                    = "../../Resources/subnets"
  name                      = "${var.project_name}-${var.environment}-${var.region_short}-epp-snet"
  resource_group_name       = azurerm_resource_group.spoke.name
  virtual_network_name      = module.spoke_vnet.vnet_name
  address_prefix            = var.epp_subnet_cidr
  location                  = var.region
  network_security_group_id = module.nsg_epp.id
  tags                      = var.tags
}