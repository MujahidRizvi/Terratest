resource "azurerm_subnet" "subnet" {
  name                 = var.name
  resource_group_name  = var.resource_group_name
  virtual_network_name = var.virtual_network_name
  address_prefixes     = [var.address_prefix]
  service_endpoints    = var.service_endpoints
}

resource "azurerm_subnet_network_security_group_association" "subnet_nsg_assoc" {
  subnet_id                 = azurerm_subnet.subnet.id
  network_security_group_id = var.network_security_group_id
}

output "subnet_id" {
  value = azurerm_subnet.subnet.id
}

output "subnet_name" {
  value = azurerm_subnet.subnet.name
}
