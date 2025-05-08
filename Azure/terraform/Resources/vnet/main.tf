resource "azurerm_virtual_network" "this" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name
  address_space       = [var.address_space]

  tags = merge({
    Name    = var.name
    Project = var.project
  }, var.tags)

  lifecycle {
    ignore_changes = [
      address_space,
      tags
    ]
    prevent_destroy = true
  }
}

output "vnet_name" {
  value = azurerm_virtual_network.this.name
}

output "vnet_id" {
  value = azurerm_virtual_network.this.id
}

# Retrieve hub VNet by name and RG (to avoid hardcoding VNet ID)
data "azurerm_virtual_network" "hub" {
  name                = var.hub_vnet_name
  resource_group_name = var.hub_resource_group
}

# Peering from this (spoke) to hub
resource "azurerm_virtual_network_peering" "to_hub" {
  name                      = "${var.name}-to-hub"
  resource_group_name       = var.resource_group_name
  virtual_network_name      = azurerm_virtual_network.this.name
  remote_virtual_network_id = data.azurerm_virtual_network.hub.id

  allow_virtual_network_access = true
  allow_forwarded_traffic      = true
  allow_gateway_transit        = false
  use_remote_gateways          = false
}

# Peering from Hub to this (Spoke)
resource "azurerm_virtual_network_peering" "from_hub" {
  name                      = "${var.hub_vnet_name}-to-${var.name}"
  resource_group_name       = var.hub_resource_group
  virtual_network_name      = data.azurerm_virtual_network.hub.name
  remote_virtual_network_id = azurerm_virtual_network.this.id

  allow_virtual_network_access = true
  allow_forwarded_traffic      = true
  allow_gateway_transit        = false
  use_remote_gateways          = false
}

# Link DNS Zones to spoke VNet, auto-registration only for azure-api.net
resource "azurerm_private_dns_zone_virtual_network_link" "dns_links" {
  for_each              = toset(var.private_dns_zones)
  name                  = "${replace(each.key, ".", "-")}-spoke-vnet-link"
  resource_group_name   = var.hub_resource_group
  private_dns_zone_name = each.key
  virtual_network_id    = azurerm_virtual_network.this.id


  tags = {
    Project = var.project
  }

  depends_on = [azurerm_virtual_network_peering.to_hub]
}
