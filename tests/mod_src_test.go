package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

type TestCaseSRC struct {
	XMLName   xml.Name    `xml:"testcase"`
	Classname string      `xml:"classname,attr"`
	Name      string      `xml:"name,attr"`
	Failure   *FailureSRC `xml:"failure,omitempty"`
	Status    string      `xml:"status"`
}

type FailureSRC struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteSRC struct {
	XMLName   xml.Name      `xml:"testsuite"`
	Tests     int           `xml:"tests,attr"`
	Failures  int           `xml:"failures,attr"`
	Errors    int           `xml:"errors,attr"`
	Time      float64       `xml:"time,attr"`
	TestCases []TestCaseSRC `xml:"testcase"`
}

func loadTFStateSRC(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesSRC(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeSRC(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestSysResourceGroupStateBased(t *testing.T) {
	var suite TestSuiteSRC
	suite.Tests = 1

	tfState1 := loadTFStateSRC(t, "../terraform.tfstate")
	tfState2 := loadTFStateSRC(t, "../terra.tfstate")
	merged := mergeStatesSRC(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_System_Resource_Group_Existence_and_Properties", func() (bool, string) {
			resourceGroups := findResourcesByTypeSRC(merged, "azurerm_resource_group")
			for _, rg := range resourceGroups {
				name, _ := rg["name"].(string)
				if strings.Contains(name, "srcs-rg") {
					if location, _ := rg["location"].(string); location != "" {
						return true, ""
					}
				}
			}
			return false, "System Resource Group 'srcs-rg' not found or invalid"
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
				suite.TestCases = append(suite.TestCases, TestCaseSRC{
					Classname: "Terraform Test",
					Name:      test.Name,
					Failure: &FailureSRC{
						Message: failureMsg,
						Type:    "error",
					},
					Status: status,
				})
			} else {
				suite.TestCases = append(suite.TestCases, TestCaseSRC{
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
	reportFile := "reports\\src_report.xml"
	err = os.WriteFile(reportFile, output, 0644)
	if err != nil {
		t.Fatalf("❌ Failed to write JUnit XML report: %v", err)
	}

	// Optionally print the output
	fmt.Println("SRC report saved to", reportFile)
}
