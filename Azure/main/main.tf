terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "rg" {
  name     = var.resource_group_name
  location = var.region

  tags = {
    Name    = var.resource_group_name
    Project = "API Ecosystem"
  }
}

resource "azurerm_virtual_network" "vnet" {
  name                = "${var.project_name}-${var.environment}-${var.region}-${var.vnet_name}-vnet"
  location            = var.region
  resource_group_name = azurerm_resource_group.rg.name
  address_space       = [var.vnet_cidr]

  tags = {
    Name    = "${var.project_name}-${var.environment}-${var.region}-${var.vnet_name}-vnet"
    Project = "API Ecosystem"
  }
}

resource "azurerm_subnet" "subnet1" {
  name                 = "${var.project_name}-${var.environment}-${var.region}-${var.subnet1_name}-snet"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = [var.subnet1_cidr]

  service_endpoints = [
    "Microsoft.Storage",
    "Microsoft.Sql",
    "Microsoft.AzureActiveDirectory",
    "Microsoft.AzureCosmosDB",
    "Microsoft.Web",
    "Microsoft.KeyVault",
    "Microsoft.EventHub",
    "Microsoft.ServiceBus",
    "Microsoft.ContainerRegistry",
    "Microsoft.CognitiveServices"
  ]

  depends_on = [azurerm_virtual_network.vnet]
}

resource "azurerm_subnet" "subnet4" {
  name                 = "${var.project_name}-${var.environment}-${var.region}-${var.subnet4_name}-snet"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = [var.subnet4_cidr]

  service_endpoints = [
    "Microsoft.Storage",
    "Microsoft.Sql",
    "Microsoft.AzureActiveDirectory",
    "Microsoft.AzureCosmosDB",
    "Microsoft.Web",
    "Microsoft.KeyVault",
    "Microsoft.EventHub",
    "Microsoft.ServiceBus",
    "Microsoft.ContainerRegistry",
    "Microsoft.CognitiveServices"
  ]

  depends_on = [azurerm_virtual_network.vnet]
}

resource "azurerm_network_security_group" "subnet1_nsg" {
  name                = "${var.project_name}-${var.environment}-${var.region}-${var.subnet1_name}-nsg"
  location            = var.region
  resource_group_name = var.resource_group_name

  depends_on = [azurerm_subnet.subnet1]
}

resource "azurerm_network_security_group" "subnet4_nsg" {
  name                = "${var.project_name}-${var.environment}-${var.region}-${var.subnet4_name}-nsg"
  location            = var.region
  resource_group_name = var.resource_group_name

  depends_on = [azurerm_subnet.subnet4]
}

resource "azurerm_subnet_network_security_group_association" "subnet1_nsg_assoc" {
  subnet_id                 = azurerm_subnet.subnet1.id
  network_security_group_id = azurerm_network_security_group.subnet1_nsg.id
}

resource "azurerm_subnet_network_security_group_association" "subnet4_nsg_assoc" {
  subnet_id                 = azurerm_subnet.subnet4.id
  network_security_group_id = azurerm_network_security_group.subnet4_nsg.id
}

resource "azurerm_network_security_rule" "subnet1_outbound" {
  name                        = "AllowAllOutbound"
  priority                    = 100
  direction                   = "Outbound"
  access                      = "Allow"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "0.0.0.0/0"
  network_security_group_name = azurerm_network_security_group.subnet1_nsg.name
  resource_group_name         = var.resource_group_name
}

resource "azurerm_network_security_rule" "subnet1_intra_vnet" {
  name                        = "AllowIntraVNet"
  priority                    = 120
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "VirtualNetwork"
  destination_address_prefix  = "VirtualNetwork"
  network_security_group_name = azurerm_network_security_group.subnet1_nsg.name
  resource_group_name         = var.resource_group_name
}

resource "azurerm_network_security_rule" "subnet1_appgw_inbound_probe_ports" {
  name                        = "AllowAppGwInbound65200-65535"
  priority                    = 130
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_ranges     = ["65200-65535"]
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  network_security_group_name = azurerm_network_security_group.subnet1_nsg.name
  resource_group_name         = var.resource_group_name
}

resource "azurerm_network_security_rule" "subnet4_ssh_inbound" {
  name                        = "AllowSSH"
  priority                    = 100
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "Tcp"
  source_port_range           = "*"
  destination_port_range      = "22"
  source_address_prefix       = "*"
  destination_address_prefix  = "*"
  network_security_group_name = azurerm_network_security_group.subnet4_nsg.name
  resource_group_name         = var.resource_group_name
}

resource "azurerm_network_security_rule" "subnet4_intra_vnet_inbound" {
  name                        = "AllowIntraVNetInbound"
  priority                    = 110
  direction                   = "Inbound"
  access                      = "Allow"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "VirtualNetwork"
  destination_address_prefix  = "VirtualNetwork"
  network_security_group_name = azurerm_network_security_group.subnet4_nsg.name
  resource_group_name         = var.resource_group_name
}

resource "azurerm_network_security_rule" "subnet4_intra_vnet_outbound" {
  name                        = "AllowIntraVNetOutbound"
  priority                    = 120
  direction                   = "Outbound"
  access                      = "Allow"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "VirtualNetwork"
  destination_address_prefix  = "VirtualNetwork"
  network_security_group_name = azurerm_network_security_group.subnet4_nsg.name
  resource_group_name         = var.resource_group_name
}

resource "azurerm_network_security_rule" "subnet4_allow_all_outbound" {
  name                        = "AllowAllOutbound"
  priority                    = 130
  direction                   = "Outbound"
  access                      = "Allow"
  protocol                    = "*"
  source_port_range           = "*"
  destination_port_range      = "*"
  source_address_prefix       = "*"
  destination_address_prefix  = "0.0.0.0/0"
  network_security_group_name = azurerm_network_security_group.subnet4_nsg.name
  resource_group_name         = var.resource_group_name
}

resource "azurerm_public_ip" "vm_public_ip" {
  name                = "${var.project_name}-${var.environment}-${var.region}-vm-pip"
  location            = var.region
  resource_group_name = var.vm_resource_group_name
  allocation_method   = "Dynamic"

  tags = {
    Name    = "${var.project_name}-${var.environment}-${var.region}-vm-pip"
    Project = "API Ecosystem"
  }
}

resource "azurerm_network_interface" "vm_nic" {
  name                = "${var.project_name}-${var.environment}-${var.region}-vm-nic"
  location            = var.region
  resource_group_name = var.vm_resource_group_name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.subnet4.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.vm_public_ip.id
  }

  tags = {
    Name    = "${var.project_name}-${var.environment}-${var.region}-vm-nic"
    Project = "API Ecosystem"
  }
}

resource "azurerm_network_interface_security_group_association" "vm_nic_nsg" {
  network_interface_id      = azurerm_network_interface.vm_nic.id
  network_security_group_id = azurerm_network_security_group.subnet4_nsg.id
}

resource "azurerm_linux_virtual_machine" "ubuntu_vm" {
  name                  = "${var.project_name}-${var.environment}-${var.region}-${var.vm_name}-vm"
  location              = var.region
  resource_group_name   = var.vm_resource_group_name
  network_interface_ids = [azurerm_network_interface.vm_nic.id]
  size                  = "Standard_B2s"
  admin_username        = var.vm_admin_username
  admin_password        = var.vm_admin_password
  disable_password_authentication = false

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
    name                 = "${var.project_name}-${var.environment}-${var.region}-osdisk"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-focal"
    sku       = "20_04-lts-gen2"
    version   = "latest"
  }

  tags = {
    Name    = "${var.project_name}-${var.environment}-${var.region}-${var.vm_name}-vm"
    Project = "API Ecosystem"
  }
}

resource "azurerm_public_ip" "main_appgw_pip" {
  name                = "${var.project_name}-${var.environment}-${var.region}-dev-agwpip"
  location            = var.region
  resource_group_name = azurerm_resource_group.rg.name
  allocation_method   = "Static"
  sku                 = "Standard"

  depends_on = [
    azurerm_virtual_network.vnet,
    azurerm_subnet.subnet1,
    azurerm_network_security_group.subnet1_nsg
  ]

  tags = {
    Name    = "${var.project_name}-${var.environment}-${var.region}-dev-agwpip"
    Project = "API Ecosystem"
  }
}

resource "azurerm_application_gateway" "main_appgw" {
  name                = "${var.project_name}-${var.environment}-${var.region}-dev-appgw"
  location            = var.region
  resource_group_name = azurerm_resource_group.rg.name

  sku {
    name     = var.appgw_sku_name
    tier     = var.appgw_sku_tier
    capacity = var.appgw_capacity
  }

  enable_http2 = true

  gateway_ip_configuration {
    name      = "appgw-ip-configuration"
    subnet_id = azurerm_subnet.subnet1.id
  }

  frontend_ip_configuration {
    name                 = var.appgw_frontend_ip_name
    public_ip_address_id = azurerm_public_ip.main_appgw_pip.id
  }

  frontend_port {
    name = var.appgw_frontend_port_name
    port = var.appgw_frontend_port
  }

  backend_address_pool {
    name = var.appgw_backend_pool_name
  }

  backend_http_settings {
    name                  = var.appgw_http_settings_name
    cookie_based_affinity = var.appgw_cookie_affinity
    path                  = var.appgw_path
    port                  = var.appgw_backend_port
    protocol              = var.appgw_backend_protocol
    request_timeout       = var.appgw_request_timeout
  }

  http_listener {
    name                           = var.appgw_listener_name
    frontend_ip_configuration_name = var.appgw_frontend_ip_name
    frontend_port_name             = var.appgw_frontend_port_name
    protocol                       = var.appgw_listener_protocol
  }

  request_routing_rule {
    name                       = var.appgw_rule_name
    rule_type                  = var.appgw_rule_type
    http_listener_name         = var.appgw_listener_name
    backend_address_pool_name  = var.appgw_backend_pool_name
    backend_http_settings_name = var.appgw_http_settings_name
    priority                   = var.appgw_rule_priority
  }

  depends_on = [
    azurerm_public_ip.main_appgw_pip,
    azurerm_subnet_network_security_group_association.subnet1_nsg_assoc
  ]

  tags = {
    Name    = "${var.project_name}-${var.environment}-${var.region}-dev-appgw"
    Project = "API Ecosystem"
  }
}

# Create private DNS zones in hub
resource "azurerm_private_dns_zone" "core_zones" {
  for_each            = toset(var.private_dns_zones)
  name                = each.value
  resource_group_name = azurerm_resource_group.rg.name

  tags = {
    Name    = each.value
    Project = "API Ecosystem"
  }
}

# Link each private DNS zone to the hub VNet
resource "azurerm_private_dns_zone_virtual_network_link" "core_links" {
  for_each              = toset(var.private_dns_zones)
  name                  = "${replace(each.value, ".", "-")}-hub-vnet-link"
  resource_group_name   = azurerm_resource_group.rg.name
  private_dns_zone_name = azurerm_private_dns_zone.core_zones[each.value].name
  virtual_network_id    = azurerm_virtual_network.vnet.id
  registration_enabled  = false

  tags = {
    Project = "API Ecosystem"
  }
}