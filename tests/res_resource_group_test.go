package test

func RunResourceGroupTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{
			"1._Verify_Resource_Group_Exists_with_Name_and_Location",
			"ResourceGroupTests",
			func() (bool, string) {
				rgs := findResourcesByType(tfState, "azurerm_resource_group")
				if len(rgs) == 0 {
					return false, "No resource group found"
				}
				for _, rg := range rgs {
					if name, ok := rg["name"].(string); ok && name != "" {
						if location, ok := rg["location"].(string); ok && location != "" {
							return true, ""
						}
					}
				}
				return false, "Resource group missing name or location"
			},
		},
		{
			"2._Verify_Tags_on_Resource_Group",
			"ResourceGroupTests",
			func() (bool, string) {
				rgs := findResourcesByType(tfState, "azurerm_resource_group")
				for _, rg := range rgs {
					if tags, ok := rg["tags"].(map[string]interface{}); ok {
						if tags["Project"] != nil && tags["Environment"] != nil {
							return true, ""
						}
					}
				}
				return false, "Resource group missing expected tags"
			},
		},
	}

	return executeTestCases(tests)
}
