output "spoke_resource_group_name" {
  value = azurerm_resource_group.spoke.name
}

output "vnet_name" {
  value = module.spoke_vnet.vnet_name
}

output "vnet_id" {
  value = module.spoke_vnet.vnet_id
}

output "subnet_bastion_id" {
  value = module.subnet_bastion.subnet_id
}

output "subnet_exp_id" {
  value = module.subnet_exp.subnet_id
}

output "subnet_proc_id" {
  value = module.subnet_proc.subnet_id
}

output "subnet_procfapp_id" {
  value = module.subnet_procfapp.subnet_id
}

output "subnet_sys_id" {
  value = module.subnet_sys.subnet_id
}

output "subnet_sysfapp_id" {
  value = module.subnet_sysfapp.subnet_id
}

output "subnet_srcs_id" {
  value = module.subnet_srcs.subnet_id
}

output "subnet_epp_id" {
  value = module.subnet_epp.subnet_id
}