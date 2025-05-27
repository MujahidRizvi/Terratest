package test

import "strings"

func RunProcTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Verify_Resource_Group_Existence_and_Properties", "ProcInfraTests", func() (bool, string) {
			for _, rg := range findResourcesByType(tfState, "azurerm_resource_group") {
				if name, _ := rg["name"].(string); strings.Contains(name, "proc-rg") {
					if location, _ := rg["location"].(string); location != "" {
						return true, ""
					}
				}
			}
			return false, "proc-rg not found or invalid"
		}},
		{"2._Verify_Function_App_Instance", "ProcInfraTests", func() (bool, string) {
			for _, fn := range findResourcesByType(tfState, "azurerm_windows_function_app") {
				if name, _ := fn["name"].(string); strings.Contains(name, "procReady-fapp") {
					return true, ""
				}
			}
			return false, "procReady-fapp not found"
		}},
		{"3._Verify_Storage_Account_Existence", "ProcInfraTests", func() (bool, string) {
			for _, sa := range findResourcesByType(tfState, "azurerm_storage_account") {
				instances, ok := sa["instances"].([]interface{})
				if !ok {
					continue
				}
				for _, inst := range instances {
					attributes := inst.(map[string]interface{})["attributes"].(map[string]interface{})
					name := attributes["name"].(string)
					if strings.Contains(strings.ToLower(name), "prfpreadystg") {
						return true, ""
					}
				}
			}
			return false, "prfpreadystg not found"
		}},
		{"4._Verify_Private_Endpoint", "ProcInfraTests", func() (bool, string) {
			for _, ep := range findResourcesByType(tfState, "azurerm_private_endpoint") {
				if name, _ := ep["name"].(string); strings.Contains(name, "prfpReady-pep") {
					return true, ""
				}
			}
			return false, "prfpReady-pep not found"
		}},
	}

	return executeTestCases(tests)
}
