terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.25.0"
    }
  }
}

provider "azurerm" {
  features {}
  subscription_id = "7de87233-96d5-4712-b750-3c65ad0afbe8"
}

# Spoke: Create VNet and all subnets + NSGs
module "spoke" {
  source = "../../Modules/Spoke"

  project_name  = "agida"
  environment   = "dev"
  region        = "UAE North"
  region_short  = "uaen"

  spoke_vnet_cidr         = "10.110.0.0/16"
  bastion_subnet_cidr     = "10.110.10.0/24"
  exp_subnet_cidr         = "10.110.20.0/24"
  proc_subnet_cidr        = "10.110.30.0/24"
  procfapp_subnet_cidr    = "10.110.31.0/24"
  sys_subnet_cidr         = "10.110.40.0/24"
  sysfapp_subnet_cidr     = "10.110.41.0/24"
  srcs_subnet_cidr        = "10.110.50.0/24"
  epp_subnet_cidr         = "10.110.60.0/24"
  

  default_nsg_rules   = var.default_nsg_rules
  exp_apim_nsg_rules  = var.exp_apim_nsg_rules
  bst_vm_nsg_rules    = var.bst_vm_nsg_rules

  tags = {
    Environment = "dev"
    Owner       = "CloudTeam"
    project     = "API Ecosystem"
  }
}

# Bastion: Windows VM in bastion subnet
module "bastion" {
  source = "../../Modules/Bastion"

  project_name   = "agida"
  environment    = "dev"
  region         = "UAE North"
  region_short   = "uaen"

  bastion_subnet_id = module.spoke.subnet_bastion_id

  admin_username   = "bastionadmin"
  admin_password   = var.bastion_password
  create_public_ip = true

  tags = {
    Environment = "dev"
    Owner       = "CloudTeam"
    project     = "API Ecosystem"
  }

  depends_on = [module.spoke]
}

# Exp: APIM with internal network and DNS linking
module "exp" {
  source = "../../Modules/Exp"

  project_name  = "agida"
  environment   = "dev"
  region        = "UAE North"
  region_short  = "uaen"

  apim_sku        = "Developer_1"
  publisher_name  = var.publisher_name
  publisher_email = var.publisher_email
  exp_subnet_id   = module.spoke.subnet_exp_id

  tags = {
    Environment = "dev"
    Owner       = "CloudTeam"
    project     = "API Ecosystem"
  }

  depends_on = [module.spoke]
}

module "proc" {
  source = "../../Modules/Proc"

  project_name  = "agida"
  environment   = "dev"
  region        = "UAE North"
  region_short  = "uaen"
  procfapp_subnet_cidr = module.spoke.subnet_procfapp_id
  procfapp_pep_subnet_cidr = module.spoke.subnet_proc_id
  depends_on = [module.spoke]
}


module "sys" {
  source = "../../Modules/Sys"

  project_name  = "agida"
  environment   = "dev"
  region        = "UAE North"
  region_short  = "uaen"
  sysfapp_subnet_cidr = module.spoke.subnet_sysfapp_id
  sysfapp_pep_subnet_cidr = module.spoke.subnet_sys_id

  depends_on = [module.spoke]
}

module "srcs" {
  source = "../../Modules/Srcs"

  project_name  = "agida"
  environment   = "dev"
  region        = "UAE North"
  region_short  = "uaen"

  depends_on = [module.spoke]
}

module "epp" {
  source = "../../Modules/Epp"

  project_name      = "agida"
  environment       = "dev"
  region            = "UAE North"
  region_short      = "uaen"
  epp_subnet_id     = module.spoke.subnet_epp_id
  sku               = "Standard"
  partition_count   = "8"
  message_retention = "7"
  capacity          = "2"
  
  tags = {
    Environment     = "dev"
    Owner           = "CloudTeam"
    Project         = "API Ecosystem"
  }

  depends_on = [module.spoke]
}