package test

func RunSubnetWithDelegationTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{
			"1._Verify_Subnet_With_Delegation_Exists",
			"SubnetDelegationTests",
			func() (bool, string) {
				subnets := findResourcesByType(tfState, "azurerm_subnet")
				for _, s := range subnets {
					if _, ok := s["delegation"]; ok {
						return true, ""
					}
				}
				return false, "No subnet found with delegation"
			},
		},
		{
			"2._Verify_Delegation_Config_Contains_Microsoft_Web_ServerFarms",
			"SubnetDelegationTests",
			func() (bool, string) {
				subnets := findResourcesByType(tfState, "azurerm_subnet")
				for _, s := range subnets {
					if delegs, ok := s["delegation"].([]interface{}); ok {
						for _, d := range delegs {
							if dMap, ok := d.(map[string]interface{}); ok {
								if dMap["name"] != nil {
									if config, ok := dMap["service_delegation"].([]interface{}); ok {
										for _, c := range config {
											if cMap, ok := c.(map[string]interface{}); ok {
												if cMap["name"] == "Microsoft.Web/serverFarms" {
													return true, ""
												}
											}
										}
									}
								}
							}
						}
					}
				}
				return false, "No valid service delegation found for Microsoft.Web/serverFarms"
			},
		},
	}

	return executeTestCases(tests)
}
