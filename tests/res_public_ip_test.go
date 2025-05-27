package test

import (
	"fmt"
)

func RunPublicIPTests(tfState map[string]interface{}) []TestCase {
	publicIPs := findResourcesByType(tfState, "azurerm_public_ip")
	tests := []GenericTest{
		{
			"1._Validate_Public_IP_Exists",
			"PublicIPTests",
			func() (bool, string) {
				if len(publicIPs) == 0 {
					return false, "No Public IP resource found"
				}
				return true, ""
			},
		},
		{
			"2._Validate_Public_IP_Tags",
			"PublicIPTests",
			func() (bool, string) {
				if len(publicIPs) == 0 {
					return false, "No Public IP resource found"
				}
				tags, ok := publicIPs[0]["tags"].(map[string]interface{})
				if !ok || len(tags) == 0 {
					return false, "Expected tags on Public IP"
				}
				return true, ""
			},
		},
		{
			"3._Validate_Public_IP_Allocation_Method",
			"PublicIPTests",
			func() (bool, string) {
				if len(publicIPs) == 0 {
					return false, "No Public IP resource found"
				}
				if method, ok := publicIPs[0]["allocation_method"].(string); !ok || method != "Static" {
					return false, fmt.Sprintf("Expected allocation_method 'Static', got: %v", publicIPs[0]["allocation_method"])
				}
				return true, ""
			},
		},
	}

	return executeTestCases(tests)
}
