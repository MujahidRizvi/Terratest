package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

type TestCasePROC struct {
	XMLName   xml.Name     `xml:"testcase"`
	Classname string       `xml:"classname,attr"`
	Name      string       `xml:"name,attr"`
	Failure   *FailurePROC `xml:"failure,omitempty"`
	Status    string       `xml:"status"`
}

type FailurePROC struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuitePROC struct {
	XMLName   xml.Name       `xml:"testsuite"`
	Tests     int            `xml:"tests,attr"`
	Failures  int            `xml:"failures,attr"`
	Errors    int            `xml:"errors,attr"`
	Time      float64        `xml:"time,attr"`
	TestCases []TestCasePROC `xml:"testcase"`
}

func loadTFStatePROC(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesPROC(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypePROC(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestAzureFunctionProcInfraStateBased(t *testing.T) {
	var suite TestSuitePROC
	suite.Tests = 4

	tfState1 := loadTFStatePROC(t, "../terraform.tfstate")
	tfState2 := loadTFStatePROC(t, "../terra.tfstate")
	merged := mergeStatesPROC(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Resource_Group_Existence_and_Properties", func() (bool, string) {
			resourceGroups := findResourcesByTypePROC(merged, "azurerm_resource_group")
			for _, rg := range resourceGroups {
				name, _ := rg["name"].(string)
				if strings.Contains(name, "proc-rg") {
					if location, _ := rg["location"].(string); location != "" {
						return true, ""
					}
				}
			}
			return false, "Resource Group 'proc-rg' not found or invalid"
		}},
		{"2._Verify_Function_App_Instance", func() (bool, string) {
			functions := findResourcesByTypePROC(merged, "azurerm_windows_function_app")
			for _, fn := range functions {
				name, _ := fn["name"].(string)
				if strings.Contains(name, "procReady-fapp") {
					return true, ""
				}
			}
			return false, "Azure Function App 'procReady-fapp' not found"
		}},
		{"3._Verify_Storage_Account_Existence", func() (bool, string) {
			storages := findResourcesByTypePROC(merged, "azurerm_storage_account")
			for _, st := range storages {
				name, _ := st["name"].(string)
				if strings.Contains(name, "prfpreadystg") {
					return true, ""
				}
			}
			return false, "Storage Account 'prfpreadystg' not found"
		}},

		{"4._Verify_Private_Endpoint", func() (bool, string) {
			endpoints := findResourcesByTypePROC(merged, "azurerm_private_endpoint")
			for _, pep := range endpoints {
				name, _ := pep["name"].(string)
				if strings.Contains(name, "prfpReady-pep") {
					return true, ""
				}
			}
			return false, "Private Endpoint 'prfpReady-pep' not found"
		}},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			passed, reason := tc.TestFunc()

			// Terminal output
			if passed {
				fmt.Printf("✅ %s passed\n", tc.Name)
			} else {
				fmt.Printf("❌ %s failed: %s\n", tc.Name, reason)
			}

			// XML report entry
			status := "PASS"
			if !passed {
				status = "FAIL"
			}
			testCase := TestCasePROC{
				Classname: "AzureFunctionProcInfraStateBasedTests",
				Name:      tc.Name,
				Status:    status,
			}
			if !passed {
				suite.Failures++
				testCase.Failure = &FailurePROC{
					Message: reason,
					Type:    "failure",
				}
			}
			suite.TestCases = append(suite.TestCases, testCase)
		})
	}

	suite.Time = 1.87
	file, err := os.Create("reports\\proc_report.xml")
	if err != nil {
		t.Fatalf("❌ Unable to create test report: %v", err)
	}
	defer file.Close()

	xmlWriter := xml.NewEncoder(file)
	xmlWriter.Indent("", "  ")
	if err := xmlWriter.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML test report: %v", err)
	}

	fmt.Println("📄 Test report written to test_report_proc.xml")
}
