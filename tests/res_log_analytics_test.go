package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

type TestCaseLogAnalytics struct {
	XMLName   xml.Name             `xml:"testcase"`
	Classname string               `xml:"classname,attr"`
	Name      string               `xml:"name,attr"`
	Failure   *FailureLogAnalytics `xml:"failure,omitempty"`
	Status    string               `xml:"status"`
}

type FailureLogAnalytics struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteLogAnalytics struct {
	XMLName   xml.Name               `xml:"testsuite"`
	Tests     int                    `xml:"tests,attr"`
	Failures  int                    `xml:"failures,attr"`
	Errors    int                    `xml:"errors,attr"`
	Time      float64                `xml:"time,attr"`
	TestCases []TestCaseLogAnalytics `xml:"testcase"`
}

func loadTFStateLogAnalytics(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesLogAnalytics(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeLogAnalytics(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var found []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return found
	}
	for _, res := range resources {
		resource, ok := res.(map[string]interface{})
		if !ok {
			continue
		}
		if resource["type"] == resourceType {
			instances, ok := resource["instances"].([]interface{})
			if !ok || len(instances) == 0 {
				continue
			}
			for _, inst := range instances {
				instance, ok := inst.(map[string]interface{})
				if !ok {
					continue
				}
				attrs, ok := instance["attributes"].(map[string]interface{})
				if ok {
					found = append(found, attrs)
				}
			}
		}
	}
	return found
}

func TestLogAnalyticsWorkspaceValidation(t *testing.T) {
	var suite TestSuiteLogAnalytics
	suite.Tests = 3

	tfState1 := loadTFStateLogAnalytics(t, "../terraform.tfstate")
	tfState2 := loadTFStateLogAnalytics(t, "../terra.tfstate")
	merged := mergeStatesLogAnalytics(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Log_Analytics_Workspace_Exists_with_Correct_Properties", func() (bool, string) {
			ws := findResourcesByTypeLogAnalytics(merged, "azurerm_log_analytics_workspace")
			if len(ws) == 0 {
				return false, "Expected a Log Analytics Workspace resource"
			}
			if ws[0]["sku"] != "PerGB2018" || ws[0]["retention_in_days"] != float64(30) {
				return false, fmt.Sprintf("Expected sku PerGB2018 and retention 30, got sku: %v, retention: %v", ws[0]["sku"], ws[0]["retention_in_days"])
			}
			return true, ""
		}},
		{"2._Verify_Tags_on_Log_Analytics_Workspace", func() (bool, string) {
			ws := findResourcesByTypeLogAnalytics(merged, "azurerm_log_analytics_workspace")
			if len(ws) == 0 {
				return false, "Workspace resource not found"
			}
			tags, ok := ws[0]["tags"].(map[string]interface{})
			if !ok || tags["Project"] != "API Ecosystem" {
				return false, "Expected tag Project=API Ecosystem"
			}
			return true, ""
		}},
		{"3._Verify_Terraform_Outputs_for_Workspace_Name_and_ID", func() (bool, string) {
			outputs, ok := merged["outputs"].(map[string]interface{})
			if !ok {
				return false, "No outputs found in state"
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
		}},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			passed, reason := tc.TestFunc()

			if passed {
				fmt.Printf("✅ %s passed\n", tc.Name)
			} else {
				fmt.Printf("❌ %s failed: %s\n", tc.Name, reason)
			}

			status := "FAIL"
			if passed {
				status = "PASS"
			}
			testCase := TestCaseLogAnalytics{
				Classname: "LogAnalyticsWorkspaceTests",
				Name:      tc.Name,
				Status:    status,
			}
			if !passed {
				suite.Failures++
				testCase.Failure = &FailureLogAnalytics{
					Message: reason,
					Type:    "failure",
				}
			}
			suite.TestCases = append(suite.TestCases, testCase)
		})
	}

	suite.Time = 1.11

	reportFile := "reports\\res_log_analytics.xml"
	file, err := os.Create(reportFile)
	if err != nil {
		t.Fatalf("❌ Failed to create XML report: %v", err)
	}
	defer file.Close()

	xmlWriter := xml.NewEncoder(file)
	xmlWriter.Indent("", "  ")
	if err := xmlWriter.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to write XML report: %v", err)
	}

	fmt.Printf("📄 XML test report written to %s\n", reportFile)
}
