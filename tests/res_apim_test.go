package test

import (
	"fmt"
	"strings"
)

func RunAPIMTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{
			"1._Verify_APIM_Resource_Deployment_and_Configuration", "APIMInfraTests", func() (bool, string) {
				apims := findResourcesByType(tfState, "azurerm_api_management")
				for _, apim := range apims {
					if name, _ := apim["name"].(string); name != "" {
						return true, ""
					}
				}
				return false, "APIM resource not found or incorrectly configured"
			},
		},
		{
			"2._Verify_App_Insights_and_Log_Analytics_Integration", "APIMInfraTests", func() (bool, string) {
				appInsights := findResourcesByType(tfState, "azurerm_application_insights")
				logAnalytics := findResourcesByType(tfState, "azurerm_log_analytics_workspace")
				if len(appInsights) > 0 && len(logAnalytics) > 0 {
					return true, ""
				}
				return false, "Application Insights or Log Analytics not found"
			},
		},
		{
			"3._Verify_Private_DNS_A_Records_for_APIM_and_Prefixes", "APIMInfraTests", func() (bool, string) {
				dnsRecords := findResourcesByType(tfState, "azurerm_private_dns_a_record")
				expectedPrefixes := []string{"management", "developer", "portal"}
				found := map[string]bool{}
				for _, rec := range dnsRecords {
					name := strings.ToLower(rec["name"].(string))
					for _, prefix := range expectedPrefixes {
						if strings.Contains(name, prefix) {
							found[prefix] = true
						}
					}
				}
				for _, prefix := range expectedPrefixes {
					if !found[prefix] {
						return false, fmt.Sprintf("Missing DNS A record containing prefix: %s", prefix)
					}
				}
				return true, ""
			},
		},
		{
			"4._Verify_AppInsights_Logger_and_Log_Retention", "APIMInfraTests", func() (bool, string) {
				loggers := findResourcesByType(tfState, "azurerm_api_management_logger")
				for _, logger := range loggers {
					appInsights, ok := logger["application_insights"].([]interface{})
					if ok && len(appInsights) > 0 {
						ai, ok := appInsights[0].(map[string]interface{})
						if ok {
							key, keyOk := ai["instrumentation_key"].(string)
							if keyOk && key != "" {
								return true, ""
							}
						}
					}
				}
				return false, "Logger not linked with Application Insights (missing instrumentation_key)"
			},
		},
		{
			"5._Verify_Terraform_Outputs_for_APIM_ID_Name_PrivateIP_FQDN", "APIMInfraTests", func() (bool, string) {
				outputsRaw, ok := tfState["outputs"].(map[string]interface{})
				if !ok {
					return false, "Outputs missing in state"
				}
				requiredKeys := []string{"apim_id", "apim_name", "apim_private_ip", "apim_fqdn"}

				for _, key := range requiredKeys {
					out, exists := outputsRaw[key]
					if !exists {
						return false, fmt.Sprintf("Missing output key: %s", key)
					}
					m, ok := out.(map[string]interface{})
					if !ok {
						return false, fmt.Sprintf("Invalid output structure for: %s", key)
					}
					val := m["value"]
					if val == nil || (fmt.Sprintf("%v", val) == "") {
						return false, fmt.Sprintf("Invalid or empty output value for: %s", key)
					}
				}
				return true, ""
			},
		},
	}

	return executeTestCases(tests)
}
