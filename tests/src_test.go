package test

import (
	"strings"
)

func RunSrcTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Verify_System_Resource_Group_Existence_and_Properties", "SRCInfraTests", func() (bool, string) {
			rgs := findResourcesByType(tfState, "azurerm_resource_group")
			for _, rg := range rgs {
				name, _ := rg["name"].(string)
				if strings.Contains(name, "srcs-rg") {
					loc, _ := rg["location"].(string)
					if loc != "" {
						return true, ""
					}
					return false, "Location missing for srcs-rg"
				}
			}
			return false, "srcs-rg not found"
		}},
	}

	return executeTestCases(tests)
}
