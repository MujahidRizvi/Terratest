data "azurerm_private_dns_zone" "dns_zones" {
  for_each            = toset(var.private_dns_zones)
  name                = each.key
  resource_group_name = "agida-main-uaen-dahub-rg"
}

resource "azurerm_private_endpoint" "this" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.subnet_id
  tags                = merge({
    Name    = var.name
    Project = var.project
  }, var.tags)

  private_service_connection {
    name                           = "${var.name}-psc"
    private_connection_resource_id = var.target_resource_id
    subresource_names              = var.subresource_names
    is_manual_connection           = var.is_manual_connection
  }
  custom_network_interface_name = var.custom_nic_name

  private_dns_zone_group {
    name = "default"

    private_dns_zone_ids = [for zone in data.azurerm_private_dns_zone.dns_zones : zone.id]
  }
}
