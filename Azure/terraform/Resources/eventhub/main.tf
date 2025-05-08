# Create an Event Hubs Namespace
resource "azurerm_eventhub_namespace" "this" {
  name                = var.eventhub_namespace_name
  location            = var.location
  resource_group_name = var.resource_group_name
  sku                 = var.sku
  capacity            = var.capacity
  tags                = var.tags
}

# Create an Event Hub within the namespace
resource "azurerm_eventhub" "this" {
  name                = var.eventhub_name
  namespace_id        = azurerm_eventhub_namespace.this.id
  partition_count     = var.partition_count
  message_retention   = var.message_retention
  depends_on          = [azurerm_eventhub_namespace.this]
}