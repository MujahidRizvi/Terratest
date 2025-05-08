locals {
  vnet_links = flatten([
    for dns_zone in var.dns_zones : [
      for vnet_id in var.virtual_network_ids : {
        key           = "${dns_zone}-${basename(vnet_id)}"
        dns_zone_name = dns_zone
        vnet_id       = vnet_id
      }
    ]
  ])
}

resource "azurerm_private_dns_zone" "this" {
  for_each            = toset(var.dns_zones)
  name                = each.value
  resource_group_name = var.resource_group_name
  tags                = merge({
    Project = var.project
  }, var.tags)
}

resource "azurerm_private_dns_zone_virtual_network_link" "this" {
  for_each = {
    for link in local.vnet_links : link.key => link
  }

  name                  = "${each.key}-link"
  resource_group_name   = var.resource_group_name
  private_dns_zone_name = each.value.dns_zone_name
  virtual_network_id    = each.value.vnet_id
  registration_enabled  = false
  tags                  = merge({
    Project = var.project
  }, var.tags)
}
