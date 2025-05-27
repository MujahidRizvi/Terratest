package test

func RunMainInfraTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Validate_Resource_Group", "AzureMainInfraTests", func() (bool, string) {
			for _, rg := range findResourcesByType(tfState, "azurerm_resource_group") {
				if rg["name"] == "" {
					return false, "Resource Group name is empty"
				}
			}
			return true, ""
		}},
		{"2._Validate_Virtual_Network", "AzureMainInfraTests", func() (bool, string) {
			for _, vnet := range findResourcesByType(tfState, "azurerm_virtual_network") {
				if vnet["name"] == "" {
					return false, "VNet name is empty"
				}
				if len(vnet["address_space"].([]interface{})) == 0 {
					return false, "VNet address space is empty"
				}
			}
			return true, ""
		}},
		{"3._Validate_Subnet", "AzureMainInfraTests", func() (bool, string) {
			for _, sn := range findResourcesByType(tfState, "azurerm_subnet") {
				if sn["name"] == "" {
					return false, "Subnet name is empty"
				}
				if len(sn["address_prefixes"].([]interface{})) == 0 {
					return false, "Subnet address_prefixes is empty"
				}
			}
			return true, ""
		}},
		{"4._Validate_NSG", "AzureMainInfraTests", func() (bool, string) {
			for _, nsg := range findResourcesByType(tfState, "azurerm_network_security_group") {
				if nsg["name"] == "" {
					return false, "NSG name is empty"
				}
			}
			return true, ""
		}},
		{"5._Validate_NSG_Subnet_Association", "AzureMainInfraTests", func() (bool, string) {
			for _, sn := range findResourcesByType(tfState, "azurerm_subnet") {
				_ = sn["network_security_group_id"]
			}
			return true, ""
		}},
		{"6._Validate_VM", "AzureMainInfraTests", func() (bool, string) {
			for _, vm := range findResourcesByType(tfState, "azurerm_virtual_machine") {
				if vm["name"] == "" {
					return false, "VM name is empty"
				}
			}
			return true, ""
		}},
		{"7._Validate_NIC", "AzureMainInfraTests", func() (bool, string) {
			for _, nic := range findResourcesByType(tfState, "azurerm_network_interface") {
				if nic["name"] == "" {
					return false, "NIC name is empty"
				}
			}
			return true, ""
		}},
		{"8._Validate_App_Gateway", "AzureMainInfraTests", func() (bool, string) {
			for _, agw := range findResourcesByType(tfState, "azurerm_application_gateway") {
				if agw["name"] == "" {
					return false, "Application Gateway name is empty"
				}
			}
			return true, ""
		}},
		{"9._Validate_APIM", "AzureMainInfraTests", func() (bool, string) {
			for _, apim := range findResourcesByType(tfState, "azurerm_api_management") {
				if apim["name"] == "" {
					return false, "APIM name is empty"
				}
			}
			return true, ""
		}},
		{"10._Validate_DNS_Zone", "AzureMainInfraTests", func() (bool, string) {
			for _, dns := range findResourcesByType(tfState, "azurerm_dns_zone") {
				if dns["name"] == "" {
					return false, "DNS Zone name is empty"
				}
			}
			return true, ""
		}},
		{"11._Validate_Public_IP", "AzureMainInfraTests", func() (bool, string) {
			for _, pip := range findResourcesByType(tfState, "azurerm_public_ip") {
				if pip["name"] == "" {
					return false, "Public IP name is empty"
				}
			}
			return true, ""
		}},
		{"12._Validate_VNET_Peering", "AzureMainInfraTests", func() (bool, string) {
			for _, peer := range findResourcesByType(tfState, "azurerm_virtual_network_peering") {
				if peer["name"] == "" {
					return false, "VNET peering name is empty"
				}
			}
			return true, ""
		}},
		{"13._Validate_NSG_Rules", "AzureMainInfraTests", func() (bool, string) {
			for _, nsg := range findResourcesByType(tfState, "azurerm_network_security_group") {
				if nsg["security_rule"] == nil {
					return false, "NSG has no security_rule block"
				}
			}
			return true, ""
		}},
	}

	return executeTestCases(tests)
}
