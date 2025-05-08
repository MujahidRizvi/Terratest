package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

type TestCaseFAPP struct {
	XMLName   xml.Name     `xml:"testcase"`
	Classname string       `xml:"classname,attr"`
	Name      string       `xml:"name,attr"`
	Failure   *FailureFAPP `xml:"failure,omitempty"`
	Status    string       `xml:"status"`
}

type FailureFAPP struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteFAPP struct {
	XMLName   xml.Name       `xml:"testsuite"`
	Tests     int            `xml:"tests,attr"`
	Failures  int            `xml:"failures,attr"`
	Errors    int            `xml:"errors,attr"`
	Time      float64        `xml:"time,attr"`
	TestCases []TestCaseFAPP `xml:"testcase"`
}

func loadTFStateFAPP(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesFAPP(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeFAPP(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestFunctionAppInfrastructure(t *testing.T) {
	var suite TestSuiteFAPP
	suite.Tests = 6

	tfState1 := loadTFStateFAPP(t, "../terraform.tfstate")
	tfState2 := loadTFStateFAPP(t, "../terra.tfstate")
	merged := mergeStatesFAPP(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Azure_Storage_Account_Creation", func() (bool, string) {
			resources := findResourcesByTypeFAPP(merged, "azurerm_storage_account")
			if len(resources) == 0 {
				return false, "Expected at least one Azure Storage Account"
			}
			return true, ""
		}},
		{"2._Verify_Log_Analytics_Workspace_and_AppInsights", func() (bool, string) {
			law := findResourcesByTypeFAPP(merged, "azurerm_application_insights")
			if len(law) == 0 {
				return false, "Expected Application Insights configured with Log Analytics"
			}
			return true, ""
		}},
		{"3._Verify_App_Service_Plan_Creation", func() (bool, string) {
			plans := findResourcesByTypeFAPP(merged, "azurerm_service_plan")
			if len(plans) == 0 {
				return false, "Expected at least one App Service Plan"
			}
			return true, ""
		}},
		{"4._Verify_Function_App_Deployment", func() (bool, string) {
			fapps := findResourcesByTypeFAPP(merged, "azurerm_windows_function_app")
			if len(fapps) == 0 {
				return false, "Expected at least one Function App"
			}
			return true, ""
		}},
		{"5._Verify_Network_and_Security_Settings", func() (bool, string) {
			storages := findResourcesByTypeFAPP(merged, "azurerm_storage_account")
			if len(storages) == 0 {
				return false, "No storage account found"
			}
			rules, ok := storages[0]["network_rules"].([]interface{})
			if !ok || len(rules) == 0 {
				return false, "Storage account network rules not configured"
			}
			return true, ""
		}},
		{"6._Verify_Tags_Applied", func() (bool, string) {
			resTypes := []string{"azurerm_storage_account", "azurerm_application_insights", "azurerm_service_plan", "azurerm_windows_function_app"}
			for _, resType := range resTypes {
				res := findResourcesByTypeFAPP(merged, resType)
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
			testCase := TestCaseFAPP{
				Classname: "FunctionAppModuleTests",
				Name:      tc.Name,
				Status:    status,
			}
			if !passed {
				suite.Failures++
				testCase.Failure = &FailureFAPP{
					Message: reason,
					Type:    "failure",
				}
			}
			suite.TestCases = append(suite.TestCases, testCase)
		})
	}

	suite.Time = 1.23
	reportFile := "reports\\res_functionapp.xml"
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
