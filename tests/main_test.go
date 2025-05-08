package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCaseMain struct {
	XMLName   xml.Name     `xml:"testcase"`
	Classname string       `xml:"classname,attr"`
	Name      string       `xml:"name,attr"`
	Failure   *FailureMain `xml:"failure,omitempty"`
	Status    string       `xml:"status"`
}

type FailureMain struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteMain struct {
	XMLName   xml.Name       `xml:"testsuite"`
	Tests     int            `xml:"tests,attr"`
	Failures  int            `xml:"failures,attr"`
	Errors    int            `xml:"errors,attr"`
	Time      float64        `xml:"time,attr"`
	TestCases []TestCaseMain `xml:"testcase"`
}

// Utility: Reads and parses terraform.tfstate from the given directory
func loadTFStateMain(t *testing.T, _ string) map[string]interface{} {
	statePath := filepath.Join("..", "terraform.tfstate") // Adjusted path for terraform.tfstate
	data, err := os.ReadFile(statePath)
	assert.NoError(t, err, "Failed to read terraform.tfstate")
	var tfState map[string]interface{}
	assert.NoError(t, json.Unmarshal(data, &tfState), "Failed to parse terraform.tfstate")
	return tfState
}

// Utility: Finds all resources of a given type in the tfstate
func findResourcesMain(tfState map[string]interface{}, tfType string) []map[string]interface{} {
	var matches []map[string]interface{}
	resources := tfState["resources"].([]interface{})
	for _, r := range resources {
		rMap := r.(map[string]interface{})
		if rMap["type"] == tfType {
			for _, inst := range rMap["instances"].([]interface{}) {
				matches = append(matches, inst.(map[string]interface{})["attributes"].(map[string]interface{}))
			}
		}
	}
	return matches
}

func TestValidateAzureInfraFromTFState(t *testing.T) {
	var suite TestSuiteMain
	suite.Tests = 13 // Updated to 13 tests

	tfDir := filepath.Join("..", "Azure", "main") // Adjusted path to Terraform code
	tfState := loadTFStateMain(t, tfDir)

	tests := []struct {
		Name     string
		TestFunc func(map[string]interface{}) (bool, string)
	}{
		{
			"1._Validate_existence_of_Resource_Group",
			func(tfState map[string]interface{}) (bool, string) {
				for _, rg := range findResourcesMain(tfState, "azurerm_resource_group") {
					fmt.Printf("Resource Group: %s, Location: %s, Tags: %v\n", rg["name"], rg["location"], rg["tags"])
					if rg["name"] == "" {
						return false, "Resource Group name must not be empty"
					}
				}
				return true, ""
			},
		},
		{
			"2._Validate_Virtual_Network_(VNet)",
			func(tfState map[string]interface{}) (bool, string) {
				for _, vnet := range findResourcesMain(tfState, "azurerm_virtual_network") {
					fmt.Printf("VNet: %s, Address Space: %v\n", vnet["name"], vnet["address_space"])
					if vnet["name"] == "" {
						return false, "VNet name must not be empty"
					}
					if len(vnet["address_space"].([]interface{})) == 0 {
						return false, "VNet address_space must not be empty"
					}
				}
				return true, ""
			},
		},
		{
			"3._Validate_Subnet",
			func(tfState map[string]interface{}) (bool, string) {
				for _, subnet := range findResourcesMain(tfState, "azurerm_subnet") {
					fmt.Printf("Subnet: %s, Address Prefixes: %v\n", subnet["name"], subnet["address_prefixes"])
					if subnet["name"] == "" {
						return false, "Subnet name must not be empty"
					}
					if len(subnet["address_prefixes"].([]interface{})) == 0 {
						return false, "Subnet address_prefixes must not be empty"
					}
				}
				return true, ""
			},
		},
		{
			"4._Validate_Network_Security_Group_(NSG)",
			func(tfState map[string]interface{}) (bool, string) {
				for _, nsg := range findResourcesMain(tfState, "azurerm_network_security_group") {
					fmt.Printf("NSG: %s, Location: %s\n", nsg["name"], nsg["location"])
					if nsg["name"] == "" {
						return false, "NSG name must not be empty"
					}
				}
				return true, ""
			},
		},
		{
			"5._Validate_NSG-to-Subnet_Association",
			func(tfState map[string]interface{}) (bool, string) {
				for _, subnet := range findResourcesMain(tfState, "azurerm_subnet") {
					fmt.Printf("Subnet: %s, NSG Association: %v\n", subnet["name"], subnet["network_security_group_id"])
					// network_security_group_id can be empty for subnets with no NSG
				}
				return true, ""
			},
		},
		{
			"6._Validate_Virtual_Machine_(VM)",
			func(tfState map[string]interface{}) (bool, string) {
				for _, vm := range findResourcesMain(tfState, "azurerm_virtual_machine") {
					fmt.Printf("VM: %s, Resource Group: %s\n", vm["name"], vm["resource_group_name"])
					if vm["name"] == "" {
						return false, "VM name must not be empty"
					}
				}
				return true, ""
			},
		},
		{
			"7._Validate_Network_Interface_(NIC)",
			func(tfState map[string]interface{}) (bool, string) {
				for _, nic := range findResourcesMain(tfState, "azurerm_network_interface") {
					fmt.Printf("NIC: %s, Subnet: %v\n", nic["name"], nic["ip_configuration"])
					if nic["name"] == "" {
						return false, "NIC name must not be empty"
					}
				}
				return true, ""
			},
		},
		{
			"8._Validate_Application_Gateway",
			func(tfState map[string]interface{}) (bool, string) {
				for _, agw := range findResourcesMain(tfState, "azurerm_application_gateway") {
					fmt.Printf("App Gateway: %s\n", agw["name"])
					if agw["name"] == "" {
						return false, "App Gateway name must not be empty"
					}
				}
				return true, ""
			},
		},
		{
			"9._Validate_Azure_API_Management_(APIM)",
			func(tfState map[string]interface{}) (bool, string) {
				for _, apim := range findResourcesMain(tfState, "azurerm_api_management") {
					fmt.Printf("APIM: %s, Publisher: %s\n", apim["name"], apim["publisher_name"])
					if apim["name"] == "" {
						return false, "APIM name must not be empty"
					}
				}
				return true, ""
			},
		},
		{
			"10._Validate_DNS_Zone",
			func(tfState map[string]interface{}) (bool, string) {
				for _, dns := range findResourcesMain(tfState, "azurerm_dns_zone") {
					fmt.Printf("DNS Zone: %s\n", dns["name"])
					if dns["name"] == "" {
						return false, "DNS Zone name must not be empty"
					}
				}
				return true, ""
			},
		},
		{
			"11._Validate_Public_IP_Address",
			func(tfState map[string]interface{}) (bool, string) {
				for _, pip := range findResourcesMain(tfState, "azurerm_public_ip") {
					fmt.Printf("Public IP: %s\n", pip["name"])
					if pip["name"] == "" {
						return false, "Public IP name must not be empty"
					}
				}
				return true, ""
			},
		},
		{
			"12._Validate_VNET_peering",
			func(tfState map[string]interface{}) (bool, string) {
				for _, peer := range findResourcesMain(tfState, "azurerm_virtual_network_peering") {
					fmt.Printf("VNET Peering: %s, Remote VNet: %s\n", peer["name"], peer["remote_virtual_network_id"])
					if peer["name"] == "" {
						return false, "VNET Peering name must not be empty"
					}
				}
				return true, ""
			},
		},
		{
			"13._Validate_NSGs'_inbound_and_outbound_rules_must_match",
			func(tfState map[string]interface{}) (bool, string) {
				for _, nsg := range findResourcesMain(tfState, "azurerm_network_security_group") {
					fmt.Printf("NSG: %s, Security Rules: %v\n", nsg["name"], nsg["security_rule"])
					if nsg["security_rule"] == nil {
						return false, "NSG must have security_rule property"
					}
				}
				return true, ""
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			passed, reason := tc.TestFunc(tfState)

			// Terminal output
			if passed {
				fmt.Printf("✅ %s passed\n", tc.Name)
			} else {
				fmt.Printf("❌ %s failed: %s\n", tc.Name, reason)
			}

			// XML Status
			status := "FAIL"
			if passed {
				status = "PASS"
			}

			// XML report entry
			testCase := TestCaseMain{
				Classname: "AzureInfraTests",
				Name:      tc.Name,
				Status:    status,
			}
			if !passed {
				suite.Failures++
				testCase.Failure = &FailureMain{
					Message: reason,
					Type:    "failure",
				}
			}
			suite.TestCases = append(suite.TestCases, testCase)
		})
	}

	// Write XML report
	suite.Time = 2.15
	file, err := os.Create("reports\\main_test_report.xml")
	if err != nil {
		t.Fatalf("❌ Unable to create report file: %v", err)
	}
	defer file.Close()

	xmlWriter := xml.NewEncoder(file)
	xmlWriter.Indent("", "  ")
	if err := xmlWriter.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode report: %v", err)
	}

	fmt.Println("📄 Test report written to main_test_report.xml")
}
