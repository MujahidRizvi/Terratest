resource "azurerm_resource_group" "bastion" {
  name     = "${var.project_name}-${var.environment}-${var.region_short}-bst-rg"
  location = var.region

  tags = {
    Project = var.project
  }
}

module "bastion_vm" {
  source              = "../../Resources/windows_vm"
  name                = "${var.project_name}${var.environment}${var.region_short}bst" # Name Can be at max 15 characters only
  location            = var.region
  resource_group_name = azurerm_resource_group.bastion.name
  subnet_id           = var.bastion_subnet_id
  vm_size             = var.vm_size
  admin_username      = var.admin_username
  admin_password      = var.admin_password
  create_public_ip    = var.create_public_ip
  tags                = var.tags

  depends_on = [azurerm_resource_group.bastion]
}

output "bastion_vm_id" {
  value = module.bastion_vm.vm_id
}

output "bastion_vm_name" {
  value = module.bastion_vm.vm_name
}

output "bastion_public_ip" {
  value = module.bastion_vm.public_ip
}
