package test

import (
	"fmt"
)

func RunAppGatewayTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Check_Application_Gateway_Exists", "AppGatewayTests", func() (bool, string) {
			gws := findResourcesByType(tfState, "azurerm_application_gateway")
			if len(gws) == 0 {
				return false, "Expected at least one Application Gateway"
			}
			return true, ""
		}},
		{"2._Check_AppGW_Has_HTTP2_Enabled", "AppGatewayTests", func() (bool, string) {
			gws := findResourcesByType(tfState, "azurerm_application_gateway")
			if len(gws) == 0 {
				return false, "No Application Gateway to check HTTP2"
			}
			if http2, ok := gws[0]["enable_http2"].(bool); !ok || !http2 {
				return false, "Expected HTTP2 to be enabled"
			}
			return true, ""
		}},
		{"3._Check_SKU_Name_And_Tier", "AppGatewayTests", func() (bool, string) {
			gws := findResourcesByType(tfState, "azurerm_application_gateway")
			if len(gws) == 0 {
				return false, "No Application Gateway to check SKU"
			}
			sku, ok := gws[0]["sku"].([]interface{})
			if !ok || len(sku) == 0 {
				return false, "SKU block is missing or empty"
			}
			skuMap := sku[0].(map[string]interface{})
			if skuMap["name"] != "Standard_v2" || skuMap["tier"] != "Standard_v2" {
				return false, fmt.Sprintf("Expected SKU name/tier to be Standard_v2, got: %v / %v", skuMap["name"], skuMap["tier"])
			}
			return true, ""
		}},
		{"4._Check_AppGW_IP_Config_Subnet", "AppGatewayTests", func() (bool, string) {
			gws := findResourcesByType(tfState, "azurerm_application_gateway")
			if len(gws) == 0 {
				return false, "No Application Gateway to check IP configuration"
			}
			ipConfigs, ok := gws[0]["gateway_ip_configuration"].([]interface{})
			if !ok || len(ipConfigs) == 0 {
				return false, "No IP configuration found"
			}
			ipConfig := ipConfigs[0].(map[string]interface{})
			if ipConfig["subnet_id"] == nil {
				return false, "Expected subnet_id in IP configuration"
			}
			return true, ""
		}},
		{"5._Check_Tags_Are_Set", "AppGatewayTests", func() (bool, string) {
			gws := findResourcesByType(tfState, "azurerm_application_gateway")
			if len(gws) == 0 {
				return false, "No Application Gateway found"
			}
			tags, ok := gws[0]["tags"].(map[string]interface{})
			if !ok || len(tags) == 0 {
				return false, "Tags block is missing or empty"
			}
			return true, ""
		}},
	}

	return executeTestCases(tests)
}
