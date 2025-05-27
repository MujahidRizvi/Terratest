package test

import (
	"strings"
)

func RunExpTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Verify_Resource_Group_Existence_and_Properties", "EXPInfraTests", func() (bool, string) {
			rgs := findResourcesByType(tfState, "azurerm_resource_group")
			for _, rg := range rgs {
				name, _ := rg["name"].(string)
				if strings.Contains(name, "exp-rg") {
					loc, _ := rg["location"].(string)
					tags, _ := rg["tags"].(map[string]interface{})
					if loc != "" && tags["Project"] == "API Ecosystem" {
						return true, ""
					}
				}
			}
			return false, "Resource Group 'exp-rg' not found or properties invalid"
		}},
		{"2._Verify_APIM_Instance", "EXPInfraTests", func() (bool, string) {
			apims := findResourcesByType(tfState, "azurerm_api_management")
			for _, apim := range apims {
				name, _ := apim["name"].(string)
				if strings.Contains(name, "exp-apim") {
					return true, ""
				}
			}
			return false, "APIM instance 'exp-apim' not found"
		}},
		{"3._Verify_APIM_Network_and_Access_Configuration", "EXPInfraTests", func() (bool, string) {
			apims := findResourcesByType(tfState, "azurerm_api_management")
			for _, apim := range apims {
				publicAccess, _ := apim["public_network_access_enabled"].(bool)
				if publicAccess {
					return true, ""
				}
			}
			return false, "APIM public network access not enabled"
		}},
		{"4._Verify_Tag_Consistency", "EXPInfraTests", func() (bool, string) {
			rgs := findResourcesByType(tfState, "azurerm_resource_group")
			for _, rg := range rgs {
				tags, _ := rg["tags"].(map[string]interface{})
				if tags["Project"] == "API Ecosystem" {
					return true, ""
				}
			}
			return false, "Tags not consistent"
		}},
		{"5._Verify_Output_Values", "EXPInfraTests", func() (bool, string) {
			outputs, ok := tfState["outputs"].(map[string]interface{})
			if !ok {
				return false, "No outputs in state"
			}
			apimID, exists := outputs["apim_id"]
			if !exists {
				return false, "Output 'apim_id' not found"
			}
			val, ok := apimID.(map[string]interface{})
			if !ok || val["value"] == nil {
				return false, "Output 'apim_id' missing value"
			}
			return true, ""
		}},
	}

	return executeTestCases(tests)
}
