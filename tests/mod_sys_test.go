package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

type TestCaseSYS struct {
	XMLName   xml.Name    `xml:"testcase"`
	Classname string      `xml:"classname,attr"`
	Name      string      `xml:"name,attr"`
	Failure   *FailureSYS `xml:"failure,omitempty"`
	Status    string      `xml:"status"`
}

type FailureSYS struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteSYS struct {
	XMLName   xml.Name      `xml:"testsuite"`
	Tests     int           `xml:"tests,attr"`
	Failures  int           `xml:"failures,attr"`
	Errors    int           `xml:"errors,attr"`
	Time      float64       `xml:"time,attr"`
	TestCases []TestCaseSYS `xml:"testcase"`
}

func loadTFStateSYS(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesSYS(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeSYS(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestSysStateBased(t *testing.T) {
	var suite TestSuiteSYS
	suite.Tests = 3

	// Load and merge state files
	tfState1 := loadTFStateSYS(t, "../terraform.tfstate")
	tfState2 := loadTFStateSYS(t, "../terra.tfstate")
	merged := mergeStatesSYS(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		// System Resource Group Verification
		{"1._Verify_System_Resource_Group_Existence_and_Properties", func() (bool, string) {
			resourceGroups := findResourcesByTypeSYS(merged, "azurerm_resource_group")
			for _, rg := range resourceGroups {
				name, _ := rg["name"].(string)
				if strings.Contains(name, "sys-rg") {
					if location, _ := rg["location"].(string); location != "" {
						return true, ""
					}
				}
			}
			return false, "System Resource Group 'sys-rg' not found or invalid"
		}},

		// Azure Function App Existence and Configuration
		{"2._Verify_Azure_Function_App_Existence_and_Configuration", func() (bool, string) {
			functionApps := findResourcesByTypeSYS(merged, "azurerm_windows_function_app")
			for _, fa := range functionApps {
				name, _ := fa["name"].(string)
				if strings.Contains(name, "sysReady-fapp") {
					if location, _ := fa["location"].(string); location != "" {
						return true, ""
					}
				}
			}
			return false, "Azure Function App 'sysReady-fapp' not found or invalid"
		}},

		// Storage Account Verification
		/*{"3._Verify_Storage_Account_Existence_and_Properties", func() (bool, string) {
			storageAccounts := findResourcesByTypeSYS(merged, "azurerm_storage_account")
			for _, sa := range storageAccounts {
				name, _ := sa["name"].(string)
				if strings.Contains(name, "sysfardystg") {
					sku, _ := sa["sku"].(map[string]interface{})
					replication, _ := sa["account_replication_type"].(string)
					if skuName, ok := sku["name"].(string); ok && skuName == "Standard" && replication == "LRS" {
						return true, ""
					}
				}
			}
			return false, "Storage Account 'sysfardystg' with correct SKU and replication type not found"
		}},*/

		// Private Endpoint Verification
		{"3._Verify_Private_Endpoint_Existence_and_Configuration", func() (bool, string) {
			privateEndpoints := findResourcesByTypeSYS(merged, "azurerm_private_endpoint")
			for _, pep := range privateEndpoints {
				name, _ := pep["name"].(string)
				if strings.Contains(strings.ToLower(name), "sysfappready-pep") {
					if targetID, _ := pep["private_service_connection"].([]interface{}); len(targetID) > 0 {
						return true, ""
					}
				}
			}
			return false, "Private Endpoint 'sysfappReady-pep' not found or invalid"
		}},
	}

	// Execute the tests
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result, failureMsg := test.TestFunc()
			status := "PASS"
			if !result {
				status = "FAIL"
				suite.Failures++
				suite.TestCases = append(suite.TestCases, TestCaseSYS{
					Classname: "Terraform Test",
					Name:      test.Name,
					Failure: &FailureSYS{
						Message: failureMsg,
						Type:    "error",
					},
					Status: status,
				})
			} else {
				suite.TestCases = append(suite.TestCases, TestCaseSYS{
					Classname: "Terraform Test",
					Name:      test.Name,
					Status:    status,
				})
			}
		})
	}

	// Marshal the XML output
	output, err := xml.MarshalIndent(suite, "", "  ")
	if err != nil {
		t.Fatalf("❌ Failed to generate XML output: %v", err)
	}

	// Save the report to a file
	reportFile := "reports\\sys_report.xml"
	err = os.WriteFile(reportFile, output, 0644)
	if err != nil {
		t.Fatalf("❌ Failed to write JUnit XML report: %v", err)
	}

	// Optionally print the output
	fmt.Println("SYS report saved to", reportFile)
}
