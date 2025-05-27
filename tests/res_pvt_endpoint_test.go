package test

import (
	"strings"
)

func RunPrivateEndpointTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{
			"1._Verify_Private_Endpoint_Exists_and_Has_Correct_Connection",
			"PrivateEndpointTests",
			func() (bool, string) {
				peps := findResourcesByType(tfState, "azurerm_private_endpoint")
				if len(peps) == 0 {
					return false, "No private endpoint found"
				}
				for _, pep := range peps {
					if conns, ok := pep["private_service_connection"].([]interface{}); ok && len(conns) > 0 {
						return true, ""
					}
				}
				return false, "No private service connection found in any private endpoint"
			},
		},
		{
			"2._Verify_Private_Endpoint_Has_Correct_Subnet_Reference",
			"PrivateEndpointTests",
			func() (bool, string) {
				peps := findResourcesByType(tfState, "azurerm_private_endpoint")
				for _, pep := range peps {
					if subnetID, ok := pep["subnet_id"].(string); ok && strings.Contains(subnetID, "subnet") {
						return true, ""
					}
				}
				return false, "No valid subnet_id found for private endpoint"
			},
		},
	}

	return executeTestCases(tests)
}
