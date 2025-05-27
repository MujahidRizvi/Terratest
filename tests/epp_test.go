package test

import (
	"fmt"
	"strings"
)

func RunEppTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Verify_EPP_Resource_Group", "EppInfraTests", func() (bool, string) {
			for _, rg := range findResourcesByType(tfState, "azurerm_resource_group") {
				if strings.Contains(rg["name"].(string), "epp-rg") {
					return true, ""
				}
			}
			return false, "EPP Resource Group 'epp-rg' not found"
		}},
		{"2._Verify_EPP_EventHub_Namespace_And_EventHub", "EppInfraTests", func() (bool, string) {
			foundNS, foundEH := false, false
			for _, ns := range findResourcesByType(tfState, "azurerm_eventhub_namespace") {
				if strings.Contains(ns["name"].(string), "epphubspace-ns") {
					foundNS = true
				}
			}
			for _, eh := range findResourcesByType(tfState, "azurerm_eventhub") {
				if strings.Contains(eh["name"].(string), "epphub-eh") {
					foundEH = true
				}
			}
			if !foundNS {
				return false, "Event Hub Namespace 'epphubspace-ns' not found"
			}
			if !foundEH {
				return false, "Event Hub 'epphub-eh' not found"
			}
			return true, ""
		}},
		{"3._Verify_EPP_Private_Endpoint", "EppInfraTests", func() (bool, string) {
			for _, pep := range findResourcesByType(tfState, "azurerm_private_endpoint") {
				if strings.Contains(pep["name"].(string), "epphub-pep") {
					return true, ""
				}
			}
			return false, "Private Endpoint 'epphub-pep' not found"
		}},
		{"4._Verify_Tag_Consistency", "EPPInfraTests", func() (bool, string) {
			rgs := findResourcesByType(tfState, "azurerm_resource_group")
			for _, rg := range rgs {
				tags, _ := rg["tags"].(map[string]interface{})
				if tags["Project"] == "API Ecosystem" {
					return true, ""
				}
			}
			return false, "Tags not consistent"
		}},
		{"5._Verify_EPP_EventHub_Namespace_Public_Network_Disabled", "EppInfraTests", func() (bool, string) {
			for _, ns := range findResourcesByType(tfState, "azurerm_eventhub_namespace") {
				if enabled, ok := ns["public_network_access_enabled"].(bool); ok && enabled {
					return false, fmt.Sprintf("Public network access ENABLED on Event Hub Namespace '%v'", ns["name"])
				}
			}
			return true, ""
		}},
	}

	return executeTestCases(tests)
}
