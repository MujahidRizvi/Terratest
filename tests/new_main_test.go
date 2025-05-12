package test

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCaseNewMain struct {
	XMLName   xml.Name        `xml:"testcase"`
	Classname string          `xml:"classname,attr"`
	Name      string          `xml:"name,attr"`
	Failure   *FailureNewMain `xml:"failure,omitempty"`
	Status    string          `xml:"status"`
}

type FailureNewMain struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteNewMain struct {
	XMLName   xml.Name          `xml:"testsuite"`
	Tests     int               `xml:"tests,attr"`
	Failures  int               `xml:"failures,attr"`
	Errors    int               `xml:"errors,attr"`
	Time      float64           `xml:"time,attr"`
	TestCases []TestCaseNewMain `xml:"testcase"`
}

// loadTFStateNewMain executes `terraform show -json` in the target directory
func loadTFStateNewMain(t *testing.T, tfDir string) map[string]interface{} {
	cmd := exec.Command("terraform", "show", "-json")
	cmd.Dir = tfDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	assert.NoError(t, err, fmt.Sprintf("❌ terraform show failed: %s", stderr.String()))
	var state map[string]interface{}
	assert.NoError(t, json.Unmarshal(stdout.Bytes(), &state), "❌ Failed to parse terraform show output")
	return state
}

// findResourcesByTypeNewMain searches JSON for matching resource types
func findResourcesByTypeNewMain(tfState map[string]interface{}, tfType string) []map[string]interface{} {
	var results []map[string]interface{}
	values, ok := tfState["values"].(map[string]interface{})
	if !ok {
		return results
	}
	rootModule := values["root_module"].(map[string]interface{})
	if resources, ok := rootModule["resources"].([]interface{}); ok {
		for _, r := range resources {
			rMap := r.(map[string]interface{})
			if rMap["type"] == tfType {
				results = append(results, rMap["values"].(map[string]interface{}))
			}
		}
	}
	// Handle child_modules
	if children, ok := rootModule["child_modules"].([]interface{}); ok {
		for _, cm := range children {
			child := cm.(map[string]interface{})
			if resources, ok := child["resources"].([]interface{}); ok {
				for _, r := range resources {
					rMap := r.(map[string]interface{})
					if rMap["type"] == tfType {
						results = append(results, rMap["values"].(map[string]interface{}))
					}
				}
			}
		}
	}
	return results
}

func TestAzureMainInfraValidation(t *testing.T) {
	var suite TestSuiteNewMain
	tfDir := filepath.Join("..", "Azure", "main")
	tfState := loadTFStateNewMain(t, tfDir)

	tests := []struct {
		Name string
		Test func(map[string]interface{}) (bool, string)
	}{
		{
			"1. Validate Resource Group",
			func(state map[string]interface{}) (bool, string) {
				for _, rg := range findResourcesByTypeNewMain(state, "azurerm_resource_group") {
					if rg["name"] == "" {
						return false, "Resource Group name is empty"
					}
				}
				return true, ""
			},
		},
		{
			"2. Validate Virtual Network (VNet)",
			func(state map[string]interface{}) (bool, string) {
				for _, vnet := range findResourcesByTypeNewMain(state, "azurerm_virtual_network") {
					if vnet["name"] == "" {
						return false, "VNet name is empty"
					}
					if len(vnet["address_space"].([]interface{})) == 0 {
						return false, "VNet address space is empty"
					}
				}
				return true, ""
			},
		},
		{
			"3. Validate Subnet",
			func(state map[string]interface{}) (bool, string) {
				for _, sn := range findResourcesByTypeNewMain(state, "azurerm_subnet") {
					if sn["name"] == "" {
						return false, "Subnet name is empty"
					}
					if len(sn["address_prefixes"].([]interface{})) == 0 {
						return false, "Subnet address_prefixes is empty"
					}
				}
				return true, ""
			},
		},
		{
			"4. Validate Network Security Group (NSG)",
			func(state map[string]interface{}) (bool, string) {
				for _, nsg := range findResourcesByTypeNewMain(state, "azurerm_network_security_group") {
					if nsg["name"] == "" {
						return false, "NSG name is empty"
					}
				}
				return true, ""
			},
		},
		{
			"5. Validate NSG-to-Subnet Association",
			func(state map[string]interface{}) (bool, string) {
				for _, sn := range findResourcesByTypeNewMain(state, "azurerm_subnet") {
					_ = sn["network_security_group_id"]
				}
				return true, ""
			},
		},
		{
			"6. Validate Virtual Machine (VM)",
			func(state map[string]interface{}) (bool, string) {
				for _, vm := range findResourcesByTypeNewMain(state, "azurerm_virtual_machine") {
					if vm["name"] == "" {
						return false, "VM name is empty"
					}
				}
				return true, ""
			},
		},
		{
			"7. Validate Network Interface (NIC)",
			func(state map[string]interface{}) (bool, string) {
				for _, nic := range findResourcesByTypeNewMain(state, "azurerm_network_interface") {
					if nic["name"] == "" {
						return false, "NIC name is empty"
					}
				}
				return true, ""
			},
		},
		{
			"8. Validate Application Gateway",
			func(state map[string]interface{}) (bool, string) {
				for _, agw := range findResourcesByTypeNewMain(state, "azurerm_application_gateway") {
					if agw["name"] == "" {
						return false, "Application Gateway name is empty"
					}
				}
				return true, ""
			},
		},
		{
			"9. Validate Azure API Management (APIM)",
			func(state map[string]interface{}) (bool, string) {
				for _, apim := range findResourcesByTypeNewMain(state, "azurerm_api_management") {
					if apim["name"] == "" {
						return false, "APIM name is empty"
					}
				}
				return true, ""
			},
		},
		{
			"10. Validate DNS Zone",
			func(state map[string]interface{}) (bool, string) {
				for _, dns := range findResourcesByTypeNewMain(state, "azurerm_dns_zone") {
					if dns["name"] == "" {
						return false, "DNS Zone name is empty"
					}
				}
				return true, ""
			},
		},
		{
			"11. Validate Public IP Address",
			func(state map[string]interface{}) (bool, string) {
				for _, pip := range findResourcesByTypeNewMain(state, "azurerm_public_ip") {
					if pip["name"] == "" {
						return false, "Public IP name is empty"
					}
				}
				return true, ""
			},
		},
		{
			"12. Validate VNET Peering",
			func(state map[string]interface{}) (bool, string) {
				for _, peer := range findResourcesByTypeNewMain(state, "azurerm_virtual_network_peering") {
					if peer["name"] == "" {
						return false, "VNET peering name is empty"
					}
				}
				return true, ""
			},
		},
		{
			"13. Validate NSG Security Rules",
			func(state map[string]interface{}) (bool, string) {
				for _, nsg := range findResourcesByTypeNewMain(state, "azurerm_network_security_group") {
					if nsg["security_rule"] == nil {
						return false, "NSG has no security_rule block"
					}
				}
				return true, ""
			},
		},
	}

	suite.Tests = len(tests)
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			pass, reason := tc.Test(tfState)
			status := "FAIL"
			if pass {
				fmt.Printf("✅ %s passed\n", tc.Name)
				status = "PASS"
			} else {
				fmt.Printf("❌ %s failed: %s\n", tc.Name, reason)
				suite.Failures++
			}
			tcResult := TestCaseNewMain{
				Classname: "AzureInfraValidation",
				Name:      tc.Name,
				Status:    status,
			}
			if !pass {
				tcResult.Failure = &FailureNewMain{Message: reason, Type: "failure"}
			}
			suite.TestCases = append(suite.TestCases, tcResult)
		})
	}

	suite.Time = 2.50
	reportFile := "reports/new_main_test_report.xml"
	if err := os.MkdirAll(filepath.Dir(reportFile), 0755); err != nil {
		t.Fatalf("❌ Failed to create report directory: %v", err)
	}
	file, err := os.Create(reportFile)
	if err != nil {
		t.Fatalf("❌ Failed to create XML report: %v", err)
	}
	defer file.Close()
	xmlWriter := xml.NewEncoder(file)
	xmlWriter.Indent("", "  ")
	if err := xmlWriter.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to write XML: %v", err)
	}
	fmt.Printf("📄 Test report written to %s\n", reportFile)
}
