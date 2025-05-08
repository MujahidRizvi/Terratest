package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

// Structs renamed for Private DNS Zone context
type TestCasePvtDnsZone struct {
	XMLName   xml.Name           `xml:"testcase"`
	Classname string             `xml:"classname,attr"`
	Name      string             `xml:"name,attr"`
	Failure   *FailurePvtDnsZone `xml:"failure,omitempty"`
	Status    string             `xml:"status"`
}

type FailurePvtDnsZone struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuitePvtDnsZone struct {
	XMLName   xml.Name             `xml:"testsuite"`
	Tests     int                  `xml:"tests,attr"`
	Failures  int                  `xml:"failures,attr"`
	Errors    int                  `xml:"errors,attr"`
	Time      float64              `xml:"time,attr"`
	TestCases []TestCasePvtDnsZone `xml:"testcase"`
}

// Functions renamed and made read-only
func loadTFStatePvtDnsZone(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesPvtDnsZone(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypePvtDnsZone(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestPvtDnsZoneStateBased(t *testing.T) {
	var suite TestSuitePvtDnsZone
	suite.Tests = 3

	tfState1 := loadTFStatePvtDnsZone(t, "../terraform.tfstate")
	tfState2 := loadTFStatePvtDnsZone(t, "../terra.tfstate")
	merged := mergeStatesPvtDnsZone(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_Private_DNS_Zones_Creation",
			func() (bool, string) {
				privateDnsZones := findResourcesByTypePvtDnsZone(merged, "azurerm_private_dns_zone")
				if len(privateDnsZones) > 0 {
					return true, ""
				}
				return false, "Private DNS Zones not found"
			},
		},
		{
			"2._Verify_Private_DNS_Zone_Virtual_Network_Links_Creation",
			func() (bool, string) {
				vnetLinks := findResourcesByTypePvtDnsZone(merged, "azurerm_private_dns_zone_virtual_network_link")
				if len(vnetLinks) > 0 {
					return true, ""
				}
				return false, "Private DNS Zone Virtual Network Links not found"
			},
		},
		{
			Name: "3._Verify_Tags_on_Created_Resources",
			TestFunc: func() (bool, string) {
				privateDnsZones := findResourcesByTypePvtDnsZone(merged, "azurerm_private_dns_zone")
				vnetLinks := findResourcesByTypePvtDnsZone(merged, "azurerm_private_dns_zone_virtual_network_link")

				// Check if tags are applied correctly
				for _, zone := range privateDnsZones {
					tags, _ := zone["tags"].(map[string]interface{})
					if len(tags) == 0 {
						return false, "No tags found on private DNS zone"
					}
				}
				for _, link := range vnetLinks {
					tags, _ := link["tags"].(map[string]interface{})
					if len(tags) == 0 {
						return false, "No tags found on virtual network link"
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
				suite.TestCases = append(suite.TestCases, TestCasePvtDnsZone{
					Classname: "Terraform Test",
					Name:      test.Name,
					Failure: &FailurePvtDnsZone{
						Message: failureMsg,
						Type:    "error",
					},
					Status: status,
				})
			} else {
				suite.TestCases = append(suite.TestCases, TestCasePvtDnsZone{
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
	reportFile := "reports\\res_pvt_dnszone_report.xml"
	if err := os.WriteFile(reportFile, output, 0644); err != nil {
		t.Fatalf("❌ Failed to write JUnit XML report: %v", err)
	}
	fmt.Println("✅ Private DNS Zone report saved to", reportFile)
}
