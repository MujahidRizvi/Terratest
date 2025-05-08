resource "azurerm_resource_group" "epp_resource_group" {
  name      = "${var.project_name}-${var.environment}-${var.region_short}-epp-rg"
  location  = var.region

  tags = var.tags
}

module "epp_eventhub"{
 
 source = "../../Resources/eventhub"
  
  eventhub_namespace_name      = "${var.project_name}-${var.environment}-${var.region_short}-epphubspace-ns"
  location                     = var.region
  resource_group_name          = azurerm_resource_group.epp_resource_group.name
  sku                          = var.sku
  capacity                     = var.capacity
  eventhub_name                = "${var.project_name}-${var.environment}-${var.region_short}-epphub-eh"
  partition_count              = var.partition_count
  message_retention            = var.message_retention
  tags                         = var.tags

  depends_on                   = [ azurerm_resource_group.epp_resource_group ]
}

module "epp_eventhub_pep" {
  source                          = "../../Resources/private-endpoint"
  name                            = "${var.project_name}-${var.environment}-${var.region_short}-epphub-pep"
  location                        = var.region
  resource_group_name             = azurerm_resource_group.epp_resource_group.name
  subnet_id                       = var.epp_subnet_id
  target_resource_id              = module.epp_eventhub.eventhub_resource_id
  subresource_names               = [var.const_subresource_name_eh]
  private_dns_zones               = var.private_ehdns_zones
  tags                            = var.tags

  depends_on                      = [ module.epp_eventhub ]
}