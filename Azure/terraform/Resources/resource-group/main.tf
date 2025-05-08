resource "azurerm_resource_group" "this" {
  name     = var.name
  location = var.location

  tags = merge({
    Name    = var.name
    Project = var.project
  }, var.tags)
}

output "name" {
  value = azurerm_resource_group.this.name
}

output "id" {
  value = azurerm_resource_group.this.id
}
