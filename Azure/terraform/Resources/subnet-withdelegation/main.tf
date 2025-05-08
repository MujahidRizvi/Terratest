resource "azurerm_subnet" "subnet_with_delegation" {
  name                 = var.name
  resource_group_name  = var.resource_group_name
  virtual_network_name = var.virtual_network_name
  address_prefixes     = [var.address_prefix]
  service_endpoints    = var.service_endpoints

  delegation {
    name = var.snet_delegation_name

    service_delegation {
      name = var.snet_delegation_service_name

      actions = [
        var.snet_delegation_action
      ]
    }
  }
}

resource "azurerm_subnet_network_security_group_association" "subnet_nsg_assoc" {
  subnet_id                 = azurerm_subnet.subnet_with_delegation.id
  network_security_group_id = var.network_security_group_id
}

output "subnet_id" {
  value = azurerm_subnet.subnet_with_delegation.id
}

output "subnet_name" {
  value = azurerm_subnet.subnet_with_delegation.name
}