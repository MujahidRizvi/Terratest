package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

type TestCaseVM struct {
	XMLName   xml.Name   `xml:"testcase"`
	Classname string     `xml:"classname,attr"`
	Name      string     `xml:"name,attr"`
	Failure   *FailureVM `xml:"failure,omitempty"`
	Status    string     `xml:"status"`
}

type FailureVM struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteVM struct {
	XMLName   xml.Name     `xml:"testsuite"`
	Tests     int          `xml:"tests,attr"`
	Failures  int          `xml:"failures,attr"`
	Errors    int          `xml:"errors,attr"`
	Time      float64      `xml:"time,attr"`
	TestCases []TestCaseVM `xml:"testcase"`
}

func loadTFStateVM(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesVM(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeVM(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestWindowsVMStateBased(t *testing.T) {
	var suite TestSuiteVM
	suite.Tests = 3

	tfState1 := loadTFStateVM(t, "../terraform.tfstate")
	tfState2 := loadTFStateVM(t, "../terra.tfstate")
	merged := mergeStatesVM(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			Name: "1._Verify_Windows_VM_Exists_with_Correct_Configuration",
			TestFunc: func() (bool, string) {
				vms := findResourcesByTypeVM(merged, "azurerm_windows_virtual_machine")
				if len(vms) == 0 {
					return false, "Windows VM not found"
				}
				for _, vm := range vms {
					if vm["name"] != nil && vm["network_interface_ids"] != nil && vm["os_disk"] != nil {
						return true, ""
					}
				}
				return false, "Windows VM exists but missing required configuration"
			},
		},
		{
			Name: "2._Verify_Network_Interface_Attached_to_VM",
			TestFunc: func() (bool, string) {
				nics := findResourcesByTypeVM(merged, "azurerm_network_interface")
				vms := findResourcesByTypeVM(merged, "azurerm_windows_virtual_machine")
				if len(nics) == 0 || len(vms) == 0 {
					return false, "NIC or VM not found"
				}
				nicID := nics[0]["id"]
				for _, vm := range vms {
					if ids, ok := vm["network_interface_ids"].([]interface{}); ok {
						for _, id := range ids {
							if id == nicID {
								return true, ""
							}
						}
					}
				}
				return false, "NIC not attached to VM"
			},
		},
		{
			Name: "3._Verify_Public_IP_Creation_and_Association",
			TestFunc: func() (bool, string) {
				publicIPs := findResourcesByTypeVM(merged, "azurerm_public_ip")
				nics := findResourcesByTypeVM(merged, "azurerm_network_interface")
				if len(publicIPs) == 0 {
					return false, "Public IP not created"
				}
				if len(nics) == 0 {
					return false, "NIC not found for validation"
				}
				for _, nic := range nics {
					if ipConfigs, ok := nic["ip_configuration"].([]interface{}); ok {
						for _, cfg := range ipConfigs {
							cfgMap, ok := cfg.(map[string]interface{})
							if ok && cfgMap["public_ip_address_id"] != nil {
								return true, ""
							}
						}
					}
				}
				return false, "Public IP not associated with NIC"
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
				suite.TestCases = append(suite.TestCases, TestCaseVM{
					Classname: "Terraform Test",
					Name:      test.Name,
					Failure: &FailureVM{
						Message: failureMsg,
						Type:    "error",
					},
					Status: status,
				})
			} else {
				suite.TestCases = append(suite.TestCases, TestCaseVM{
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
	reportFile := "reports\\windows_vm_report.xml"
	if err := os.WriteFile(reportFile, output, 0644); err != nil {
		t.Fatalf("❌ Failed to write JUnit XML report: %v", err)
	}
	fmt.Println("✅ Windows VM report saved to", reportFile)
}
