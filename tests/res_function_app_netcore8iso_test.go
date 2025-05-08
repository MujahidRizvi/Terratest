package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

type TestCaseFAPP8 struct {
	XMLName   xml.Name      `xml:"testcase"`
	Classname string        `xml:"classname,attr"`
	Name      string        `xml:"name,attr"`
	Failure   *FailureFAPP8 `xml:"failure,omitempty"`
	Status    string        `xml:"status"`
}

type FailureFAPP8 struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteFAPP8 struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Time      float64         `xml:"time,attr"`
	TestCases []TestCaseFAPP8 `xml:"testcase"`
}

func loadTFStateFAPP8(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesFAPP8(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeFAPP8(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestFunctionAppNetCore8ISOValidations(t *testing.T) {
	var suite TestSuiteFAPP8
	suite.Tests = 4

	tfState1 := loadTFStateFAPP8(t, "../terraform.tfstate")
	tfState2 := loadTFStateFAPP8(t, "../terra.tfstate")
	merged := mergeStatesFAPP8(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Azure_Storage_Account_Creation_and_Configuration", func() (bool, string) {
			sa := findResourcesByTypeFAPP8(merged, "azurerm_storage_account")
			if len(sa) == 0 {
				return false, "Expected a storage account resource"
			}
			if sa[0]["account_tier"] == "" || sa[0]["account_replication_type"] == "" {
				return false, "Storage account missing configuration"
			}
			return true, ""
		}},
		{"2._Verify_Azure_App_Service_Plan_Creation_and_Configuration", func() (bool, string) {
			asp := findResourcesByTypeFAPP8(merged, "azurerm_service_plan")
			if len(asp) == 0 {
				return false, "Expected an App Service Plan resource"
			}
			if asp[0]["os_type"] != "Windows" {
				return false, fmt.Sprintf("Expected OS type to be Windows, got %v", asp[0]["os_type"])
			}
			return true, ""
		}},
		{"3._Verify_Azure_Windows_Function_App_Deployment_and_Settings", func() (bool, string) {
			fn := findResourcesByTypeFAPP8(merged, "azurerm_windows_function_app")
			if len(fn) == 0 {
				return false, "Expected a Windows Function App resource"
			}
			config, ok := fn[0]["site_config"].([]interface{})
			if !ok || len(config) == 0 {
				return false, "Missing site_config in function app"
			}
			return true, ""
		}},
		{"4._Verify_Virtual_Network_Integration_and_Public_Access_Setting", func() (bool, string) {
			fn := findResourcesByTypeFAPP8(merged, "azurerm_windows_function_app")
			if len(fn) == 0 {
				return false, "No function app found"
			}
			if fn[0]["virtual_network_subnet_id"] == nil {
				return false, "Function App not integrated with VNet"
			}
			if fn[0]["public_network_access_enabled"] != false {
				return false, "Public network access should be disabled"
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
			testCase := TestCaseFAPP8{
				Classname: "FunctionAppNetCore8ISOTests",
				Name:      tc.Name,
				Status:    status,
			}
			if !passed {
				suite.Failures++
				testCase.Failure = &FailureFAPP8{
					Message: reason,
					Type:    "failure",
				}
			}
			suite.TestCases = append(suite.TestCases, testCase)
		})
	}

	suite.Time = 1.23

	reportFile := "reports\\res_function_app_netcore8iso.xml"
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
