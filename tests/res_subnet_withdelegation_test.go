package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

// XML Structs for JUnit report
type TestCaseSubnet struct {
	XMLName   xml.Name       `xml:"testcase"`
	Classname string         `xml:"classname,attr"`
	Name      string         `xml:"name,attr"`
	Failure   *FailureSubnet `xml:"failure,omitempty"`
	Status    string         `xml:"status"`
}

type FailureSubnet struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteSubnet struct {
	XMLName   xml.Name         `xml:"testsuite"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Errors    int              `xml:"errors,attr"`
	Time      float64          `xml:"time,attr"`
	TestCases []TestCaseSubnet `xml:"testcase"`
}

// Reuse functions from previous script
func loadTFStateSubnet(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesSubnet(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeSubnet(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestSubnetWithDelegation(t *testing.T) {
	var suite TestSuiteSubnet
	suite.Tests = 4

	tfState1 := loadTFStateSubnet(t, "../terraform.tfstate")
	tfState2 := loadTFStateSubnet(t, "../terra.tfstate")
	merged := mergeStatesSubnet(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_Subnet_Creation_with_Correct_Address_Prefix_and_Service_Endpoints",
			func() (bool, string) {
				subnets := findResourcesByTypeSubnet(merged, "azurerm_subnet")
				for _, subnet := range subnets {
					addrPrefixes, _ := subnet["address_prefixes"].([]interface{})
					serviceEndpoints, _ := subnet["service_endpoints"].([]interface{})
					if len(addrPrefixes) > 0 && len(serviceEndpoints) >= 3 {
						return true, ""
					}
				}
				return false, "Subnet with correct address prefix and service endpoints not found"
			},
		},
		{
			"2._Verify_Subnet_Delegation_Configuration",
			func() (bool, string) {
				subnets := findResourcesByTypeSubnet(merged, "azurerm_subnet")
				for _, subnet := range subnets {
					delegations, ok := subnet["delegation"].([]interface{})
					if ok && len(delegations) > 0 {
						for _, d := range delegations {
							dMap, ok := d.(map[string]interface{})
							if ok && dMap["name"] != nil {
								return true, ""
							}
						}
					}
				}
				return false, "Delegation block not found or incorrectly configured in subnet"
			},
		},
		{
			"3._Verify_NSG_Association_with_Subnet",
			func() (bool, string) {
				assocs := findResourcesByTypeSubnet(merged, "azurerm_subnet_network_security_group_association")
				for _, assoc := range assocs {
					if assoc["network_security_group_id"] != nil && assoc["subnet_id"] != nil {
						return true, ""
					}
				}
				return false, "NSG association with subnet not found"
			},
		},
		{
			"4._Verify_Terraform_Outputs_for_Subnet_ID_and_Name",
			func() (bool, string) {
				outputs, ok := tfState1["outputs"].(map[string]interface{})
				if !ok {
					return false, "Terraform outputs not found"
				}
				for _, key := range []string{"subnet_id", "subnet_name"} {
					out, exists := outputs[key]
					if !exists {
						return false, fmt.Sprintf("Output '%s' missing", key)
					}
					valMap, ok := out.(map[string]interface{})
					if !ok || valMap["value"] == nil || valMap["value"] == "" {
						return false, fmt.Sprintf("Output '%s' is empty or invalid", key)
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
				suite.TestCases = append(suite.TestCases, TestCaseSubnet{
					Classname: "Terraform Test",
					Name:      test.Name,
					Failure: &FailureSubnet{
						Message: failureMsg,
						Type:    "error",
					},
					Status: status,
				})
			} else {
				suite.TestCases = append(suite.TestCases, TestCaseSubnet{
					Classname: "Terraform Test",
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
	reportFile := "reports\\res_subnet_with_delegation_report.xml"
	if err := os.WriteFile(reportFile, xmlOut, 0644); err != nil {
		t.Fatalf("❌ Failed to write XML report: %v", err)
	}
	fmt.Println("✅ Subnet with Delegation report saved to", reportFile)
}
