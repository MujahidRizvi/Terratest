package test

func RunVNetValidationTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{
			"1._Verify_VNet_Creation_and_Address_Space",
			"VirtualNetworkTests",
			func() (bool, string) {
				vnets := findResourcesByType(tfState, "azurerm_virtual_network")
				if len(vnets) == 0 {
					return false, "No Virtual Network found"
				}
				for _, v := range vnets {
					if v["name"] == "" {
						return false, "VNet missing name"
					}
					if addrSpaces, ok := v["address_space"].([]interface{}); !ok || len(addrSpaces) == 0 {
						return false, "VNet missing address space"
					}
				}
				return true, ""
			},
		},
		{
			"2._Verify_DNS_Servers_and_Tags_If_Set",
			"VirtualNetworkTests",
			func() (bool, string) {
				vnets := findResourcesByType(tfState, "azurerm_virtual_network")
				for _, v := range vnets {
					if tags, ok := v["tags"].(map[string]interface{}); ok && len(tags) > 0 {
						return true, ""
					}
				}
				return false, "No tags found on Virtual Network"
			},
		},
	}

	return executeTestCases(tests)
}
