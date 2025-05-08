package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

// Structs renamed for APIM module context
type TestCaseAPIM struct {
	XMLName   xml.Name     `xml:"testcase"`
	Classname string       `xml:"classname,attr"`
	Name      string       `xml:"name,attr"`
	Failure   *FailureAPIM `xml:"failure,omitempty"`
	Status    string       `xml:"status"`
}

type FailureAPIM struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteAPIM struct {
	XMLName   xml.Name       `xml:"testsuite"`
	Tests     int            `xml:"tests,attr"`
	Failures  int            `xml:"failures,attr"`
	Errors    int            `xml:"errors,attr"`
	Time      float64        `xml:"time,attr"`
	TestCases []TestCaseAPIM `xml:"testcase"`
}

// Functions renamed and made read-only
func loadTFStateAPIM(t *testing.T, path string) map[string]interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("❌ Failed to read terraform state from %s: %v", path, err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(data, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse terraform state from %s: %v", path, err)
	}
	return tfState
}

func mergeStatesAPIM(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeAPIM(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var foundResources []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return foundResources
	}
	for _, res := range resources {
		resourceMap, ok := res.(map[string]interface{})
		if !ok {
			continue
		}
		if resourceMap["type"] == resourceType {
			instances, ok := resourceMap["instances"].([]interface{})
			if !ok || len(instances) == 0 {
				continue
			}
			for _, inst := range instances {
				instanceMap, ok := inst.(map[string]interface{})
				if !ok {
					continue
				}
				attributes, ok := instanceMap["attributes"].(map[string]interface{})
				if ok {
					foundResources = append(foundResources, attributes)
				}
			}
		}
	}
	return foundResources
}

func TestApimStateBased(t *testing.T) {
	var suite TestSuiteAPIM
	suite.Tests = 5

	tfState1 := loadTFStateAPIM(t, "../terraform.tfstate")
	tfState2 := loadTFStateAPIM(t, "../terra.tfstate")
	merged := mergeStatesAPIM(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_APIM_Resource_Deployment_and_Configuration",
			func() (bool, string) {
				apims := findResourcesByTypeAPIM(merged, "azurerm_api_management")
				for _, apim := range apims {
					name, _ := apim["name"].(string)
					if name != "" {
						return true, ""
					}
				}
				return false, "APIM resource not found or incorrectly configured"
			},
		},
		{
			"2._Verify_App_Insights_and_Log_Analytics_Integration",
			func() (bool, string) {
				appInsights := findResourcesByTypeAPIM(merged, "azurerm_application_insights")
				logAnalytics := findResourcesByTypeAPIM(merged, "azurerm_log_analytics_workspace")
				if len(appInsights) > 0 && len(logAnalytics) > 0 {
					return true, ""
				}
				return false, "Application Insights or Log Analytics not found"
			},
		},
		{
			Name: "3._Verify_Private_DNS_A_Records_for_APIM_and_Prefixes",
			TestFunc: func() (bool, string) {
				dnsRecords := findResourcesByTypeAPIM(merged, "azurerm_private_dns_a_record")
				expectedPrefixes := []string{"management", "developer", "portal"}

				foundPrefixes := make(map[string]bool)
				for _, rec := range dnsRecords {
					name, _ := rec["name"].(string)
					for _, prefix := range expectedPrefixes {
						if name == prefix {
							foundPrefixes[prefix] = true
						}
					}
				}

				for _, prefix := range expectedPrefixes {
					if !foundPrefixes[prefix] {
						return false, fmt.Sprintf("Missing expected DNS A record for prefix: %s", prefix)
					}
				}
				return true, ""
			},
		},

		{
			Name: "4._Verify_AppInsights_Logger_and_Log_Retention",
			TestFunc: func() (bool, string) {
				loggers := findResourcesByTypeAPIM(merged, "azurerm_api_management_logger")
				appInsights := findResourcesByTypeAPIM(merged, "azurerm_application_insights")

				if len(loggers) == 0 || len(appInsights) == 0 {
					return false, "Missing either Application Insights or APIM Logger resources"
				}

				// Check logger is using Application Insights
				for _, logger := range loggers {
					resourceID, _ := logger["resource_id"].(string)
					if strings.Contains(resourceID, "applicationInsights") {
						return true, ""
					}
				}

				return false, "Logger does not appear to be linked with Application Insights"
			},
		},
		{
			Name: "5._Verify_Terraform_Outputs_for_APIM_ID_Name_PrivateIP_FQDN",
			TestFunc: func() (bool, string) {
				outputsRaw, ok := tfState1["outputs"].(map[string]interface{})
				if !ok {
					return false, "Terraform outputs not found in state file"
				}

				requiredKeys := []string{"apim_id", "apim_name", "apim_private_ip", "apim_fqdn"}
				for _, key := range requiredKeys {
					output, exists := outputsRaw[key]
					if !exists {
						return false, fmt.Sprintf("Output key '%s' not found", key)
					}
					outputMap, ok := output.(map[string]interface{})
					if !ok || outputMap["value"] == nil || outputMap["value"] == "" {
						return false, fmt.Sprintf("Output key '%s' is empty or invalid", key)
					}
				}
				return true, ""
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result, failureMsg := test.TestFunc()
			status := "PASS"
			if !result {
				status = "FAIL"
				suite.Failures++
				suite.TestCases = append(suite.TestCases, TestCaseAPIM{
					Classname: "Terraform Test",
					Name:      test.Name,
					Failure: &FailureAPIM{
						Message: failureMsg,
						Type:    "error",
					},
					Status: status,
				})
			} else {
				suite.TestCases = append(suite.TestCases, TestCaseAPIM{
					Classname: "Terraform Test",
					Name:      test.Name,
					Status:    status,
				})
			}
		})
	}

	output, err := xml.MarshalIndent(suite, "", "  ")
	if err != nil {
		t.Fatalf("❌ Failed to generate XML report: %v", err)
	}
	reportFile := "reports\\res_apim_report.xml"
	if err := os.WriteFile(reportFile, output, 0644); err != nil {
		t.Fatalf("❌ Failed to write JUnit XML report: %v", err)
	}
	fmt.Println("✅ APIM report saved to", reportFile)
}
