package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

// XML Structs
type TestCaseSNET struct {
	XMLName   xml.Name     `xml:"testcase"`
	Classname string       `xml:"classname,attr"`
	Name      string       `xml:"name,attr"`
	Failure   *FailureSNET `xml:"failure,omitempty"`
	Status    string       `xml:"status"`
}

type FailureSNET struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteSNET struct {
	XMLName   xml.Name       `xml:"testsuite"`
	Tests     int            `xml:"tests,attr"`
	Failures  int            `xml:"failures,attr"`
	Errors    int            `xml:"errors,attr"`
	Time      float64        `xml:"time,attr"`
	TestCases []TestCaseSNET `xml:"testcase"`
}

// Helpers
func loadTFStateSNET(t *testing.T, path string) map[string]interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("❌ Failed to read Terraform state from %s: %v", path, err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(data, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse Terraform state from %s: %v", path, err)
	}
	return tfState
}

func mergeStatesSNET(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeSNET(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var found []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return found
	}
	for _, res := range resources {
		rMap, ok := res.(map[string]interface{})
		if !ok || rMap["type"] != resourceType {
			continue
		}
		instances, ok := rMap["instances"].([]interface{})
		if !ok {
			continue
		}
		for _, inst := range instances {
			instMap, ok := inst.(map[string]interface{})
			if !ok {
				continue
			}
			if attrs, ok := instMap["attributes"].(map[string]interface{}); ok {
				found = append(found, attrs)
			}
		}
	}
	return found
}

// Main Test
func TestSubnet(t *testing.T) {
	var suite TestSuiteSNET
	suite.Tests = 3

	tfState1 := loadTFStateSNET(t, "../terraform.tfstate")
	tfState2 := loadTFStateSNET(t, "../terra.tfstate")
	merged := mergeStatesSNET(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_Subnet_Exists_with_Correct_Address_Prefix_and_Service_Endpoints",
			func() (bool, string) {
				subnets := findResourcesByTypeSNET(merged, "azurerm_subnet")
				for _, subnet := range subnets {
					prefixes, _ := subnet["address_prefixes"].([]interface{})
					endpoints, _ := subnet["service_endpoints"].([]interface{})
					if len(prefixes) == 1 && len(endpoints) >= 3 {
						return true, ""
					}
				}
				return false, "Expected subnet with correct address prefix and service endpoints not found"
			},
		},
		{
			"2._Verify_Network_Security_Group_Association_with_Subnet",
			func() (bool, string) {
				assocs := findResourcesByTypeSNET(merged, "azurerm_subnet_network_security_group_association")
				for _, assoc := range assocs {
					if assoc["network_security_group_id"] != nil && assoc["subnet_id"] != nil {
						return true, ""
					}
				}
				return false, "NSG association with subnet not found"
			},
		},
		{
			"3._Verify_Terraform_Outputs_for_Subnet_ID_and_Name",
			func() (bool, string) {
				stateList := []map[string]interface{}{tfState1, tfState2}
				requiredKeys := []string{"subnet_id", "subnet_name"}

				for _, key := range requiredKeys {
					found := false
					for _, state := range stateList {
						outputs, ok := state["outputs"].(map[string]interface{})
						if !ok {
							continue
						}
						out, exists := outputs[key]
						if !exists {
							continue
						}
						valMap, ok := out.(map[string]interface{})
						if !ok {
							continue
						}
						if valMap["value"] != nil && valMap["value"] != "" {
							found = true
							break
						}
					}
					if !found {
						return false, fmt.Sprintf("Output '%s' missing or empty in all state files", key)
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
				suite.TestCases = append(suite.TestCases, TestCaseSNET{
					Classname: "Terraform Subnet Module",
					Name:      test.Name,
					Failure: &FailureSNET{
						Message: failureMsg,
						Type:    "error",
					},
					Status: status,
				})
			} else {
				suite.TestCases = append(suite.TestCases, TestCaseSNET{
					Classname: "Terraform Subnet Module",
					Name:      test.Name,
					Status:    status,
				})
			}
		})
	}

	xmlOut, err := xml.MarshalIndent(suite, "", "  ")
	if err != nil {
		t.Fatalf("❌ Failed to generate XML report: %v", err)
	}
	reportFile := "reports\\res_subnet_report.xml"
	if err := os.WriteFile(reportFile, xmlOut, 0644); err != nil {
		t.Fatalf("❌ Failed to write XML report: %v", err)
	}
	fmt.Println("✅ Subnet report saved to", reportFile)
}
