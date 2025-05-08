package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

type TestCaseRG struct {
	XMLName   xml.Name   `xml:"testcase"`
	Classname string     `xml:"classname,attr"`
	Name      string     `xml:"name,attr"`
	Failure   *FailureRG `xml:"failure,omitempty"`
	Status    string     `xml:"status"`
}

type FailureRG struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteRG struct {
	XMLName   xml.Name     `xml:"testsuite"`
	Tests     int          `xml:"tests,attr"`
	Failures  int          `xml:"failures,attr"`
	Errors    int          `xml:"errors,attr"`
	Time      float64      `xml:"time,attr"`
	TestCases []TestCaseRG `xml:"testcase"`
}

func loadTFStateRG(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesRG(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeRG(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var foundResources []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return foundResources
	}
	for _, res := range resources {
		resourceMap, ok := res.(map[string]interface{})
		if !ok || resourceMap["type"] != resourceType {
			continue
		}
		instances, ok := resourceMap["instances"].([]interface{})
		if !ok {
			continue
		}
		for _, inst := range instances {
			if instanceMap, ok := inst.(map[string]interface{}); ok {
				if attributes, ok := instanceMap["attributes"].(map[string]interface{}); ok {
					foundResources = append(foundResources, attributes)
				}
			}
		}
	}
	return foundResources
}

func TestResourceGroupStateBased(t *testing.T) {
	var suite TestSuiteRG
	suite.Tests = 3

	tfState1 := loadTFStateRG(t, "../terraform.tfstate")
	tfState2 := loadTFStateRG(t, "../terra.tfstate")
	merged := mergeStatesRG(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			Name: "1._Verify_Resource_Group_Exists_with_Correct_Name_and_Location",
			TestFunc: func() (bool, string) {
				rgs := findResourcesByTypeRG(merged, "azurerm_resource_group")
				if len(rgs) == 0 {
					return false, "Resource group not found"
				}
				for _, rg := range rgs {
					if rg["name"] != nil && rg["location"] != nil {
						return true, ""
					}
				}
				return false, "Resource group exists but name or location is missing"
			},
		},
		{
			Name: "2._Verify_Tags_Applied_on_Resource_Group",
			TestFunc: func() (bool, string) {
				rgs := findResourcesByTypeRG(merged, "azurerm_resource_group")
				if len(rgs) == 0 {
					return false, "Resource group not found for tag check"
				}
				tags, ok := rgs[0]["tags"].(map[string]interface{})
				if !ok || len(tags) == 0 {
					return false, "No tags found on the resource group"
				}
				if _, hasName := tags["Name"]; !hasName {
					return false, "Expected tag 'Name' not found"
				}
				if _, hasProject := tags["Project"]; !hasProject {
					return false, "Expected tag 'Project' not found"
				}
				return true, ""
			},
		},
		{
			Name: "3._Verify_Terraform_Outputs_for_Resource_Group_Name_and_ID",
			TestFunc: func() (bool, string) {
				outputs, ok := tfState1["outputs"].(map[string]interface{})
				if !ok {
					return false, "Terraform outputs not found"
				}
				required := []string{"name", "id"}
				for _, key := range required {
					out, ok := outputs[key]
					if !ok {
						return false, fmt.Sprintf("Output '%s' not found", key)
					}
					outMap, ok := out.(map[string]interface{})
					if !ok || outMap["value"] == nil || outMap["value"] == "" {
						return false, fmt.Sprintf("Output '%s' is invalid or empty", key)
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
				suite.TestCases = append(suite.TestCases, TestCaseRG{
					Classname: "Terraform Test",
					Name:      test.Name,
					Failure: &FailureRG{
						Message: failureMsg,
						Type:    "error",
					},
					Status: status,
				})
			} else {
				suite.TestCases = append(suite.TestCases, TestCaseRG{
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
	reportFile := "reports\\res_resource_group_report.xml"
	if err := os.WriteFile(reportFile, output, 0644); err != nil {
		t.Fatalf("❌ Failed to write JUnit XML report: %v", err)
	}
	fmt.Println("✅ Resource Group report saved to", reportFile)
}
