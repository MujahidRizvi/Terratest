package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

type TestCaseSPK struct {
	XMLName   xml.Name    `xml:"testcase"`
	Classname string      `xml:"classname,attr"`
	Name      string      `xml:"name,attr"`
	Failure   *FailureSPK `xml:"failure,omitempty"`
	Status    string      `xml:"status"`
}

type FailureSPK struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteSPK struct {
	XMLName   xml.Name      `xml:"testsuite"`
	Tests     int           `xml:"tests,attr"`
	Failures  int           `xml:"failures,attr"`
	Errors    int           `xml:"errors,attr"`
	Time      float64       `xml:"time,attr"`
	TestCases []TestCaseSPK `xml:"testcase"`
}

func loadTFStateSPK(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesSPK(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeSPK(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestSpokeInfraStateBased(t *testing.T) {
	var suite TestSuiteSPK
	suite.Tests = 11

	tfState1 := loadTFStateSPK(t, "../terraform.tfstate")
	tfState2 := loadTFStateSPK(t, "../terra.tfstate")
	merged := mergeStatesSPK(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Resource_Group_Existence_and_Properties", func() (bool, string) {
			resourceGroups := findResourcesByTypeSPK(merged, "azurerm_resource_group")
			for _, rg := range resourceGroups {
				name, _ := rg["name"].(string)
				if strings.Contains(name, "spk-rg") {
					if location, _ := rg["location"].(string); location != "" {
						return true, ""
					}
				}
			}
			return false, "Resource Group 'spk-rg' not found or invalid"
		}},
		{"2._Verify_Virtual_Network_Existence", func() (bool, string) {
			vnets := findResourcesByTypeSPK(merged, "azurerm_virtual_network")
			for _, vn := range vnets {
				name, _ := vn["name"].(string)
				if strings.Contains(name, "spoke-vnet") {
					return true, ""
				}
			}
			return false, "Virtual Network 'spoke-vnet' not found"
		}},
		{"3._Verify_Bastion_Network_Security_Group", func() (bool, string) {
			nsgs := findResourcesByTypeSPK(merged, "azurerm_network_security_group")
			for _, nsg := range nsgs {
				name, _ := nsg["name"].(string)
				if strings.Contains(name, "bst-nsg") {
					return true, ""
				}
			}
			return false, "Bastion NSG 'bst-nsg' not found"
		}},
		{"4._Verify_Bastion_Subnet_Existence", func() (bool, string) {
			subnets := findResourcesByTypeSPK(merged, "azurerm_subnet")
			for _, sn := range subnets {
				name, _ := sn["name"].(string)
				if strings.Contains(name, "bst-snet") {
					return true, ""
				}
			}
			return false, "Bastion Subnet 'bst-snet' not found"
		}},
		{"5._Verify_EXP_NSG_and_Subnet", func() (bool, string) {
			nsgs := findResourcesByTypeSPK(merged, "azurerm_network_security_group")
			for _, nsg := range nsgs {
				name, _ := nsg["name"].(string)
				if strings.Contains(name, "exp-nsg") {
					return true, ""
				}
			}
			subnets := findResourcesByTypeSPK(merged, "azurerm_subnet")
			for _, sn := range subnets {
				name, _ := sn["name"].(string)
				if strings.Contains(name, "exp-snet") {
					return true, ""
				}
			}
			return false, "EXP NSG or Subnet not found"
		}},
		{"6._Verify_PROC_NSG_and_Subnet", func() (bool, string) {
			nsgs := findResourcesByTypeSPK(merged, "azurerm_network_security_group")
			for _, nsg := range nsgs {
				name, _ := nsg["name"].(string)
				if strings.Contains(name, "proc-nsg") {
					return true, ""
				}
			}
			subnets := findResourcesByTypeSPK(merged, "azurerm_subnet")
			for _, sn := range subnets {
				name, _ := sn["name"].(string)
				if strings.Contains(name, "proc-snet") {
					return true, ""
				}
			}
			return false, "PROC NSG or Subnet not found"
		}},
		{"7._Verify_PROCFAPP_NSG_and_Subnet_with_Delegation", func() (bool, string) {
			nsgs := findResourcesByTypeSPK(merged, "azurerm_network_security_group")
			for _, nsg := range nsgs {
				name, _ := nsg["name"].(string)
				if strings.Contains(name, "procfapp-nsg") {
					return true, ""
				}
			}
			subnets := findResourcesByTypeSPK(merged, "azurerm_subnet_with_delegation")
			for _, sn := range subnets {
				name, _ := sn["name"].(string)
				if strings.Contains(name, "procfapp-snet") {
					return true, ""
				}
			}
			return false, "PROCFAPP NSG or Subnet with Delegation not found"
		}},
		{"8._Verify_SYS_NSG_and_Subnet", func() (bool, string) {
			nsgs := findResourcesByTypeSPK(merged, "azurerm_network_security_group")
			for _, nsg := range nsgs {
				name, _ := nsg["name"].(string)
				if strings.Contains(name, "sys-nsg") {
					return true, ""
				}
			}
			subnets := findResourcesByTypeSPK(merged, "azurerm_subnet")
			for _, sn := range subnets {
				name, _ := sn["name"].(string)
				if strings.Contains(name, "sys-snet") {
					return true, ""
				}
			}
			return false, "SYS NSG or Subnet not found"
		}},
		{"9._Verify_SYSFAPP_NSG_and_Subnet_with_Delegation", func() (bool, string) {
			nsgs := findResourcesByTypeSPK(merged, "azurerm_network_security_group")
			for _, nsg := range nsgs {
				name, _ := nsg["name"].(string)
				if strings.Contains(name, "sysfapp-nsg") {
					return true, ""
				}
			}
			subnets := findResourcesByTypeSPK(merged, "azurerm_subnet_with_delegation")
			for _, sn := range subnets {
				name, _ := sn["name"].(string)
				if strings.Contains(name, "sysfapp-snet") {
					return true, ""
				}
			}
			return false, "SYSFAPP NSG or Subnet with Delegation not found"
		}},
		{"10._Verify_Shared_Resources_NSG_and_Subnet", func() (bool, string) {
			nsgs := findResourcesByTypeSPK(merged, "azurerm_network_security_group")
			for _, nsg := range nsgs {
				name, _ := nsg["name"].(string)
				if strings.Contains(name, "srcs-nsg") {
					return true, ""
				}
			}
			subnets := findResourcesByTypeSPK(merged, "azurerm_subnet")
			for _, sn := range subnets {
				name, _ := sn["name"].(string)
				if strings.Contains(name, "srcs-snet") {
					return true, ""
				}
			}
			return false, "Shared Resources NSG or Subnet not found"
		}},
		{"11._Verify_EPP_NSG_and_Subnet", func() (bool, string) {
			nsgs := findResourcesByTypeSPK(merged, "azurerm_network_security_group")
			for _, nsg := range nsgs {
				name, _ := nsg["name"].(string)
				if strings.Contains(name, "epp-nsg") {
					return true, ""
				}
			}
			subnets := findResourcesByTypeSPK(merged, "azurerm_subnet")
			for _, sn := range subnets {
				name, _ := sn["name"].(string)
				if strings.Contains(name, "epp-snet") {
					return true, ""
				}
			}
			return false, "EPP NSG or Subnet not found"
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
				suite.TestCases = append(suite.TestCases, TestCaseSPK{
					Classname: "Terraform Test",
					Name:      test.Name,
					Failure: &FailureSPK{
						Message: failureMsg,
						Type:    "error",
					},
					Status: status,
				})
			} else {
				suite.TestCases = append(suite.TestCases, TestCaseSPK{
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
	reportFile := "reports\\spoke_report.xml"
	err = os.WriteFile(reportFile, output, 0644)
	if err != nil {
		t.Fatalf("❌ Failed to write Spoke XML report: %v", err)
	}

	// Optionally print the output
	fmt.Println("Spoke report saved to", reportFile)
}
