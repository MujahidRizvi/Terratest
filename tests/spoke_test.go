package test

import "strings"

func RunSpokeTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Verify_Resource_Group_Existence_and_Properties", "SpokeInfraTests", func() (bool, string) {
			for _, rg := range findResourcesByType(tfState, "azurerm_resource_group") {
				if name, _ := rg["name"].(string); strings.Contains(name, "spk-rg") {
					if location, _ := rg["location"].(string); location != "" {
						return true, ""
					}
				}
			}
			return false, "spk-rg not found"
		}},
		{"2._Verify_Virtual_Network_Existence", "SpokeInfraTests", func() (bool, string) {
			for _, vn := range findResourcesByType(tfState, "azurerm_virtual_network") {
				if name, _ := vn["name"].(string); strings.Contains(name, "spoke-vnet") {
					return true, ""
				}
			}
			return false, "spoke-vnet not found"
		}},
		{"3._Verify_Bastion_NSG", "SpokeInfraTests", func() (bool, string) {
			for _, nsg := range findResourcesByType(tfState, "azurerm_network_security_group") {
				if name, _ := nsg["name"].(string); strings.Contains(name, "bst-nsg") {
					return true, ""
				}
			}
			return false, "bst-nsg not found"
		}},
	}

	return executeTestCases(tests)
}
