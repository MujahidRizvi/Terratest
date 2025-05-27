package test

import (
	"fmt"
)

func RunLogAnalyticsTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{
			"1._Verify_Log_Analytics_Workspace_Exists_with_Correct_Properties",
			"LogAnalyticsTests",
			func() (bool, string) {
				ws := findResourcesByType(tfState, "azurerm_log_analytics_workspace")
				if len(ws) == 0 {
					return false, "Expected a Log Analytics Workspace"
				}
				if ws[0]["sku"] != "PerGB2018" || ws[0]["retention_in_days"] != float64(30) {
					return false, fmt.Sprintf(
						"Expected sku=PerGB2018 & retention=30, got sku=%v, retention=%v",
						ws[0]["sku"], ws[0]["retention_in_days"])
				}
				return true, ""
			},
		},
		{
			"2._Verify_Tags_on_Log_Analytics_Workspace",
			"LogAnalyticsTests",
			func() (bool, string) {
				ws := findResourcesByType(tfState, "azurerm_log_analytics_workspace")
				if len(ws) == 0 {
					return false, "Workspace not found"
				}
				tags, ok := ws[0]["tags"].(map[string]interface{})
				if !ok || tags["Project"] != "API Ecosystem" {
					return false, "Missing tag: Project=API Ecosystem"
				}
				return true, ""
			},
		},
		{
			"3._Verify_Terraform_Outputs_for_Workspace_Name_and_ID",
			"LogAnalyticsTests",
			func() (bool, string) {
				outputs, ok := tfState["outputs"].(map[string]interface{})
				if !ok {
					return false, "Outputs block not found"
				}
				idOut, ok1 := outputs["log_analytics_workspace_id"].(map[string]interface{})
				nameOut, ok2 := outputs["log_analytics_workspace_name"].(map[string]interface{})
				if !ok1 || idOut["value"] == "" {
					return false, "Missing or empty log_analytics_workspace_id"
				}
				if !ok2 || nameOut["value"] == "" {
					return false, "Missing or empty log_analytics_workspace_name"
				}
				return true, ""
			},
		},
	}

	return executeTestCases(tests)
}
