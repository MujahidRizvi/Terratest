package test

import (
	"strings"
)

func RunNSGTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{
			"1._Verify_NSG_Creation_and_Name_Tagging",
			"NSGTests",
			func() (bool, string) {
				nsgs := findResourcesByType(tfState, "azurerm_network_security_group")
				if len(nsgs) == 0 {
					return false, "No NSG resources found"
				}
				for _, nsg := range nsgs {
					if name, ok := nsg["name"].(string); ok && strings.TrimSpace(name) != "" {
						tags, ok := nsg["tags"].(map[string]interface{})
						if !ok || len(tags) == 0 {
							return false, "NSG has no tags"
						}
						return true, ""
					}
				}
				return false, "No NSG with valid name found"
			},
		},
		{
			"2._Verify_NSG_Security_Rules_Configured",
			"NSGTests",
			func() (bool, string) {
				nsgs := findResourcesByType(tfState, "azurerm_network_security_group")
				for _, nsg := range nsgs {
					rules, ok := nsg["security_rule"].([]interface{})
					if !ok || len(rules) == 0 {
						return false, "NSG has no security rules configured"
					}
				}
				return true, ""
			},
		},
	}

	return executeTestCases(tests)
}
