package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

type TestCaseVMNew struct {
	XMLName   xml.Name      `xml:"testcase"`
	Classname string        `xml:"classname,attr"`
	Name      string        `xml:"name,attr"`
	Failure   *FailureVMNew `xml:"failure,omitempty"`
	Status    string        `xml:"status"`
}

type FailureVMNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteVMNew struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Time      float64         `xml:"time,attr"`
	TestCases []TestCaseVMNew `xml:"testcase"`
}

func loadRemoteTFStateVMNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote Terraform state: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read remote state: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse remote state JSON: %v", err)
	}
	return tfState
}

func findResourcesByTypeVMNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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
		for _, inst := range resourceMap["instances"].([]interface{}) {
			instMap, ok := inst.(map[string]interface{})
			if !ok {
				continue
			}
			if attrs, ok := instMap["attributes"].(map[string]interface{}); ok {
				foundResources = append(foundResources, attrs)
			}
		}
	}
	return foundResources
}

func TestWindowsVMStateBasedNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"

	tfState := loadRemoteTFStateVMNew(t, remoteStateURL)

	var suite TestSuiteVMNew
	suite.Tests = 3

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			Name: "1._Verify_Windows_VM_Exists_with_Correct_Configuration",
			TestFunc: func() (bool, string) {
				vms := findResourcesByTypeVMNew(tfState, "azurerm_windows_virtual_machine")
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
				nics := findResourcesByTypeVMNew(tfState, "azurerm_network_interface")
				vms := findResourcesByTypeVMNew(tfState, "azurerm_windows_virtual_machine")
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
				publicIPs := findResourcesByTypeVMNew(tfState, "azurerm_public_ip")
				nics := findResourcesByTypeVMNew(tfState, "azurerm_network_interface")
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
			}
			suite.TestCases = append(suite.TestCases, TestCaseVMNew{
				Classname: "WindowsVMTests",
				Name:      test.Name,
				Status:    status,
				Failure: func() *FailureVMNew {
					if !result {
						return &FailureVMNew{Message: failureMsg, Type: "error"}
					}
					return nil
				}(),
			})
		})
	}

	output, err := xml.MarshalIndent(suite, "", "  ")
	if err != nil {
		t.Fatalf("❌ Failed to generate XML report: %v", err)
	}
	reportFile := "reports/new_windows_vm_report.xml"
	if err := os.WriteFile(reportFile, output, 0644); err != nil {
		t.Fatalf("❌ Failed to write XML report: %v", err)
	}
	fmt.Println("📄 windows_vm_report.xml written successfully")
}
