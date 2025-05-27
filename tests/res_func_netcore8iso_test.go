package test

func RunFunctionAppNetCore8ISOTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Verify_Storage_Account_Creation_and_Configuration", "FunctionAppNetCore8ISOTests", func() (bool, string) {
			storageAccounts := findResourcesByType(tfState, "azurerm_storage_account")
			if len(storageAccounts) == 0 {
				return false, "Expected a storage account resource"
			}
			sa := storageAccounts[0]
			tier, _ := sa["account_tier"].(string)
			replication, _ := sa["account_replication_type"].(string)
			if tier == "" || replication == "" {
				return false, "Storage account missing tier or replication type"
			}
			return true, ""
		}},
		{"2._Verify_App_Service_Plan_Creation_and_Configuration", "FunctionAppNetCore8ISOTests", func() (bool, string) {
			appPlans := findResourcesByType(tfState, "azurerm_service_plan")
			if len(appPlans) == 0 {
				return false, "Expected an App Service Plan resource"
			}
			plan := appPlans[0]
			osType, _ := plan["os_type"].(string)
			if osType != "Windows" {
				return false, "App Service Plan OS type is not Windows"
			}
			return true, ""
		}},
		{"3._Verify_Windows_Function_App_Deployment_and_Settings", "FunctionAppNetCore8ISOTests", func() (bool, string) {
			funcApps := findResourcesByType(tfState, "azurerm_windows_function_app")
			if len(funcApps) == 0 {
				return false, "Expected a Windows Function App resource"
			}
			fn := funcApps[0]
			configs, ok := fn["site_config"].([]interface{})
			if !ok || len(configs) == 0 {
				return false, "Function App missing site_config"
			}
			return true, ""
		}},
		{"4._Verify_VNet_Integration_and_Public_Access_Setting", "FunctionAppNetCore8ISOTests", func() (bool, string) {
			funcApps := findResourcesByType(tfState, "azurerm_windows_function_app")
			if len(funcApps) == 0 {
				return false, "No Windows Function App found"
			}
			fn := funcApps[0]
			if fn["virtual_network_subnet_id"] == nil {
				return false, "Function App not integrated with Virtual Network"
			}
			publicAccess, ok := fn["public_network_access_enabled"].(bool)
			if !ok || publicAccess {
				return false, "Public network access should be disabled"
			}
			return true, ""
		}},
	}

	return executeTestCases(tests)
}
