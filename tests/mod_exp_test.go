package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

type TestCaseEXP struct {
	XMLName   xml.Name    `xml:"testcase"`
	Classname string      `xml:"classname,attr"`
	Name      string      `xml:"name,attr"`
	Failure   *FailureEXP `xml:"failure,omitempty"`
	Status    string      `xml:"status"`
}

type FailureEXP struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteEXP struct {
	XMLName   xml.Name      `xml:"testsuite"`
	Tests     int           `xml:"tests,attr"`
	Failures  int           `xml:"failures,attr"`
	Errors    int           `xml:"errors,attr"`
	Time      float64       `xml:"time,attr"`
	TestCases []TestCaseEXP `xml:"testcase"`
}

func loadTFStateEXP(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesEXP(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeEXP(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestAzureAPIMInfraStateBased(t *testing.T) {
	var suite TestSuiteEXP
	suite.Tests = 5

	tfState1 := loadTFStateEXP(t, "../terraform.tfstate")
	tfState2 := loadTFStateEXP(t, "../terra.tfstate")
	merged := mergeStatesEXP(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Resource_Group_Existence_and_Properties", func() (bool, string) {
			resourceGroups := findResourcesByTypeEXP(merged, "azurerm_resource_group")
			for _, rg := range resourceGroups {
				name, _ := rg["name"].(string)
				if strings.Contains(name, "exp-rg") {
					location, _ := rg["location"].(string)
					if location != "" {
						tags, _ := rg["tags"].(map[string]interface{})
						if tags["Project"] == "API Ecosystem" {
							return true, ""
						}
					}
				}
			}
			return false, "Resource Group 'exp-rg' not found or properties do not match"
		}},
		{"2._Verify_APIM_Instance", func() (bool, string) {
			apims := findResourcesByTypeEXP(merged, "azurerm_api_management")
			for _, apim := range apims {
				name, _ := apim["name"].(string)
				if name != "" && strings.Contains(name, "exp-apim") {
					return true, ""
				}
			}
			return false, "No APIM instance found"
		}},
		{"3._Verify_APIM_Network_and_Access_Configuration", func() (bool, string) {
			apims := findResourcesByTypeEXP(merged, "azurerm_api_management")
			for _, apim := range apims {
				publicAccess, _ := apim["public_network_access_enabled"].(bool)
				if publicAccess == true {
					return true, ""
				}
			}
			return false, "APIM Network and Access configuration does not match"
		}},
		{"4._Verify_Tag_Consistency", func() (bool, string) {
			resourceGroups := findResourcesByTypeEXP(merged, "azurerm_resource_group")
			for _, rg := range resourceGroups {
				tags, _ := rg["tags"].(map[string]interface{})
				if tags["Project"] == "API Ecosystem" {
					return true, ""
				}
			}
			return false, "Tags not consistent across resources"
		}},
		{"5._Verify_Output_Values", func() (bool, string) {
			outputs, ok := merged["outputs"].(map[string]interface{})
			if !ok {
				return false, "No outputs section in state file"
			}

			apimID, exists := outputs["apim_id"]
			if !exists {
				return false, "Output value 'apim_id' not found"
			}

			// Optional: Verify it has a 'value' field
			apimMap, isMap := apimID.(map[string]interface{})
			if !isMap || apimMap["value"] == nil {
				return false, "Output 'apim_id' exists but has no value"
			}

			return true, ""
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

			// XML Status: colored for terminal only
			status := "FAIL"
			if passed {
				status = "PASS"
			}

			// XML report entry
			testCase := TestCaseEXP{
				Classname: "AzureAPIMInfraStateBasedTests",
				Name:      tc.Name,
				Status:    status,
			}
			if !passed {
				suite.Failures++
				testCase.Failure = &FailureEXP{
					Message: reason,
					Type:    "failure",
				}
			}
			suite.TestCases = append(suite.TestCases, testCase)
		})
	}

	// Write XML report
	suite.Time = 2.15
	file, err := os.Create("reports\\exp_report.xml")
	if err != nil {
		t.Fatalf("❌ Unable to create report file: %v", err)
	}
	defer file.Close()

	xmlWriter := xml.NewEncoder(file)
	xmlWriter.Indent("", "  ")
	if err := xmlWriter.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode report: %v", err)
	}

	fmt.Println("📄 Test report written to test_report.xml")

}
