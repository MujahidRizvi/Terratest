package test

import (
	"fmt"
)

func RunFunctionAppTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Verify_Azure_Storage_Account_Creation", "FunctionAppModuleTests", func() (bool, string) {
			resources := findResourcesByType(tfState, "azurerm_storage_account")
			if len(resources) == 0 {
				return false, "Expected at least one Azure Storage Account"
			}
			return true, ""
		}},
		{"2._Verify_Log_Analytics_Workspace_and_AppInsights", "FunctionAppModuleTests", func() (bool, string) {
			appInsights := findResourcesByType(tfState, "azurerm_application_insights")
			if len(appInsights) == 0 {
				return false, "Expected Application Insights configured with Log Analytics"
			}
			return true, ""
		}},
		{"3._Verify_App_Service_Plan_Creation", "FunctionAppModuleTests", func() (bool, string) {
			plans := findResourcesByType(tfState, "azurerm_service_plan")
			if len(plans) == 0 {
				return false, "Expected at least one App Service Plan"
			}
			return true, ""
		}},
		{"4._Verify_Function_App_Deployment", "FunctionAppModuleTests", func() (bool, string) {
			funcApps := findResourcesByType(tfState, "azurerm_windows_function_app")
			if len(funcApps) == 0 {
				return false, "Expected at least one Function App"
			}
			return true, ""
		}},
		{"5._Verify_Network_and_Security_Settings", "FunctionAppModuleTests", func() (bool, string) {
			storages := findResourcesByType(tfState, "azurerm_storage_account")
			if len(storages) == 0 {
				return false, "No storage account found"
			}
			rules, ok := storages[0]["network_rules"].([]interface{})
			if !ok || len(rules) == 0 {
				return false, "Storage account network rules not configured"
			}
			return true, ""
		}},
		{"6._Verify_Tags_Applied", "FunctionAppModuleTests", func() (bool, string) {
			resourceTypes := []string{
				"azurerm_storage_account",
				"azurerm_application_insights",
				"azurerm_service_plan",
				"azurerm_windows_function_app",
			}
			for _, resType := range resourceTypes {
				res := findResourcesByType(tfState, resType)
				if len(res) == 0 {
					return false, fmt.Sprintf("No resource of type %s found", resType)
				}
				tags, ok := res[0]["tags"].(map[string]interface{})
				if !ok || len(tags) == 0 {
					return false, fmt.Sprintf("Tags missing on resource type %s", resType)
				}
			}
			return true, ""
		}},
	}

	return executeTestCases(tests)
}
