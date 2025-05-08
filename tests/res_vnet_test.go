package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

// Structs renamed for VNET module context
type TestCaseVNET struct {
	XMLName   xml.Name     `xml:"testcase"`
	Classname string       `xml:"classname,attr"`
	Name      string       `xml:"name,attr"`
	Failure   *FailureVNET `xml:"failure,omitempty"`
	Status    string       `xml:"status"`
}

type FailureVNET struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type"`
}

type TestSuiteVNET struct {
	XMLName   xml.Name       `xml:"testsuite"`
	Tests     int            `xml:"tests,attr"`
	Failures  int            `xml:"failures,attr"`
	Errors    int            `xml:"errors,attr"`
	Time      float64        `xml:"time,attr"`
	TestCases []TestCaseVNET `xml:"testcase"`
}

func loadTFStateVNET(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesVNET(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeVNET(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestVnetStateBased(t *testing.T) {
	var suite TestSuiteVNET
	suite.Tests = 4

	tfState1 := loadTFStateVNET(t, "../terraform.tfstate")
	tfState2 := loadTFStateVNET(t, "../terra.tfstate")
	merged := mergeStatesVNET(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_Virtual_Network_Exists_with_Correct_Address_Space_and_Tags",
			func() (bool, string) {
				vnets := findResourcesByTypeVNET(merged, "azurerm_virtual_network")
				for _, vnet := range vnets {
					addressSpace, _ := vnet["address_space"].([]interface{})
					tags, _ := vnet["tags"].(map[string]interface{})
					if len(addressSpace) > 0 && tags["Project"] != nil {
						return true, ""
					}
				}
				return false, "Virtual network with correct address space and tags not found"
			},
		},
		{
			"2._Verify_Virtual_Network_Peering_from_Spoke_to_Hub",
			func() (bool, string) {
				peerings := findResourcesByTypeVNET(merged, "azurerm_virtual_network_peering")
				for _, peer := range peerings {
					name, _ := peer["name"].(string)
					if strings.Contains(name, "to-hub") {
						return true, ""
					}
				}
				return false, "VNet peering from spoke to hub not found"
			},
		},
		{
			"3._Verify_Virtual_Network_Peering_from_Hub_to_Spoke",
			func() (bool, string) {
				peerings := findResourcesByTypeVNET(merged, "azurerm_virtual_network_peering")
				for _, peer := range peerings {
					name, _ := peer["name"].(string)
					if strings.Contains(name, "hub") && strings.Contains(name, "to") {
						return true, ""
					}
				}
				return false, "VNet peering from hub to spoke not found"
			},
		},
		{
			"4._Verify_Private_DNS_Zone_Virtual_Network_Links_Creation",
			func() (bool, string) {
				links := findResourcesByTypeVNET(merged, "azurerm_private_dns_zone_virtual_network_link")
				if len(links) == 0 {
					return false, "Private DNS zone virtual network links not found"
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
				suite.TestCases = append(suite.TestCases, TestCaseVNET{
					Classname: "Terraform Test",
					Name:      test.Name,
					Failure: &FailureVNET{
						Message: failureMsg,
						Type:    "error",
					},
					Status: status,
				})
			} else {
				suite.TestCases = append(suite.TestCases, TestCaseVNET{
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
	reportFile := "reports\\res_vnet_report.xml"
	if err := os.WriteFile(reportFile, output, 0644); err != nil {
		t.Fatalf("❌ Failed to write JUnit XML report: %v", err)
	}
	fmt.Println("✅ VNET report saved to", reportFile)
}
