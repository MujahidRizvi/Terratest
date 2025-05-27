package test

import (
	"strings"
)

func RunSysTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Verify_System_Resource_Group_Existence_and_Properties", "SystemInfraTests", func() (bool, string) {
			rgs := findResourcesByType(tfState, "azurerm_resource_group")
			for _, rg := range rgs {
				if strings.Contains(rg["name"].(string), "sys-rg") {
					if loc, _ := rg["location"].(string); loc != "" {
						return true, ""
					}
				}
			}
			return false, "sys-rg not found or invalid"
		}},
		{"2._Verify_Azure_Function_App_Existence_and_Configuration", "SystemInfraTests", func() (bool, string) {
			fapps := findResourcesByType(tfState, "azurerm_windows_function_app")
			for _, fa := range fapps {
				if strings.Contains(fa["name"].(string), "sysReady-fapp") {
					if loc, _ := fa["location"].(string); loc != "" {
						return true, ""
					}
				}
			}
			return false, "sysReady-fapp not found or invalid"
		}},
		{"3._Verify_Private_Endpoint_Existence_and_Configuration", "SystemInfraTests", func() (bool, string) {
			peps := findResourcesByType(tfState, "azurerm_private_endpoint")
			for _, pep := range peps {
				if strings.Contains(strings.ToLower(pep["name"].(string)), "sysready-fapp-pep") {
					if conn, ok := pep["private_service_connection"].([]interface{}); ok && len(conn) > 0 {
						return true, ""
					}
				}
			}
			return false, "sysfappReady-pep not found or misconfigured"
		}},
	}

	return executeTestCases(tests)
}
