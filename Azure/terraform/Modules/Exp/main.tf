resource "azurerm_resource_group" "exp" {
  name     = "${var.project_name}-${var.environment}-${var.region_short}-exp-rg"
  location = var.region

  tags = {
    Project = var.project
  }
}

module "apim" {
  source                        = "../../Resources/apim"

  name                          = "${var.project_name}-${var.environment}-${var.region_short}-exp-apim"
  location                      = var.region
  resource_group_name           = azurerm_resource_group.exp.name
  sku_name                      = var.apim_sku
  publisher_name                = var.publisher_name
  publisher_email               = var.publisher_email
  outbound_subnet_id            = var.exp_subnet_id
  public_network_access_enabled = var.public_network_access_enabled
  virtual_network_type          = var.virtual_network_type
  tags                          = var.tags
  project                       = var.project

  depends_on = [azurerm_resource_group.exp]
}