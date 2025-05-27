package test

func RunSubnetTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{
			"1._Verify_Subnet_Created_with_Name_and_Prefix",
			"SubnetTests",
			func() (bool, string) {
				subnets := findResourcesByType(tfState, "azurerm_subnet")
				if len(subnets) == 0 {
					return false, "No subnets found"
				}
				for _, s := range subnets {
					if name, ok := s["name"].(string); !ok || name == "" {
						return false, "Subnet missing name"
					}
					prefixes, ok := s["address_prefixes"].([]interface{})
					if !ok || len(prefixes) == 0 {
						return false, "Missing address_prefixes"
					}
				}
				return true, ""
			},
		},
		{
			"2._Verify_Subnet_NSG_Association_Exists",
			"SubnetTests",
			func() (bool, string) {
				subnets := findResourcesByType(tfState, "azurerm_subnet")
				for _, s := range subnets {
					if s["network_security_group_id"] == nil {
						name, _ := s["name"].(string)
						return false, "Subnet " + name + " has no NSG associated"
					}
				}
				return true, ""
			},
		},
	}

	return executeTestCases(tests)
}
