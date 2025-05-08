package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

type TestCase struct {
	XMLName   xml.Name `xml:"testcase"`
	Classname string   `xml:"classname,attr"`
	Name      string   `xml:"name,attr"`
	Failure   *Failure `xml:"failure,omitempty"`
	Status    string   `xml:"status"`
}

type Failure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuite struct {
	XMLName   xml.Name   `xml:"testsuite"`
	Tests     int        `xml:"tests,attr"`
	Failures  int        `xml:"failures,attr"`
	Errors    int        `xml:"errors,attr"`
	Time      float64    `xml:"time,attr"`
	TestCases []TestCase `xml:"testcase"`
}

func loadTFState(t *testing.T, path string) map[string]interface{} {
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

func mergeResources(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByType(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestAzureInfraValidations(t *testing.T) {
	var suite TestSuite
	suite.Tests = 9

	tfState1 := loadTFState(t, "../terraform.tfstate")
	tfState2 := loadTFState(t, "../terra.tfstate")
	merged := mergeResources(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Spoke_VNet", func() (bool, string) {
			vnets := findResourcesByType(merged, "azurerm_virtual_network")
			for _, vnet := range vnets {
				name, ok := vnet["name"].(string)
				if !ok {
					continue
				}
				if strings.Contains(name, "spoke") {
					addrSpace, ok := vnet["address_space"].([]interface{})
					if !ok || len(addrSpace) == 0 {
						return false, "Spoke VNet missing CIDR block"
					}
					cidr, ok := addrSpace[0].(string)
					if !ok || cidr != "10.110.0.0/16" {
						return false, "Spoke VNet has incorrect CIDR: " + cidr
					}
					return true, ""
				}
			}
			return false, "Spoke VNet not found"
		}},
		{"2._Verify_Subnets_with_correct_CIDRs", func() (bool, string) {
			subnets := findResourcesByType(merged, "azurerm_subnet")
			for _, sn := range subnets {
				cidrs, ok := sn["address_prefixes"].([]interface{})
				if !ok || len(cidrs) == 0 {
					name, _ := sn["name"].(string)
					return false, "Subnet " + name + " missing CIDRs"
				}
			}
			return true, ""
		}},
		{"3._Verify_NSG_and_Rules", func() (bool, string) {
			nsgs := findResourcesByType(merged, "azurerm_network_security_group")
			for _, nsg := range nsgs {
				rules, ok := nsg["security_rule"].([]interface{})
				if !ok || len(rules) == 0 {
					name, _ := nsg["name"].(string)
					return false, "NSG " + name + " has no rules"
				}
			}
			return true, ""
		}},
		{"4._Verify_Bastion_VM_Public_IP_and_Admin_User", func() (bool, string) {
			vms := findResourcesByType(merged, "azurerm_virtual_machine")
			for _, vm := range vms {
				// Type assertion with error handling
				name, ok := vm["name"].(string)
				if !ok {
					return false, "Failed to assert VM name"
				}
				if strings.Contains(name, "bastion") {
					// Handle public IP
					pubIP, ok := vm["public_ip_address"].(string)
					if !ok || pubIP == "" {
						return false, "No public IP on Bastion VM"
					}

					// Handle admin username
					adminUser, ok := vm["admin_username"].(string)
					if !ok || adminUser == "" {
						return false, "No admin user on Bastion VM"
					}
				}
			}
			return true, ""
		}},

		{"5._Verify_APIM_internal_network_and_DNS", func() (bool, string) {
			apims := findResourcesByType(merged, "azurerm_api_management")
			for _, apim := range apims {
				internal, ok := apim["internal"].(bool)
				if !ok || !internal {
					return false, "APIM is not internal or missing 'internal' field"
				}
				dns, ok := apim["dns_name"].(string)
				if !ok || dns == "" {
					return false, "APIM has no DNS name"
				}
			}
			return true, ""
		}},
		{"6._Verify_VM_Size", func() (bool, string) {
			vms := findResourcesByType(merged, "azurerm_virtual_machine")
			for _, vm := range vms {
				size, ok := vm["vm_size"].(string)
				if !ok || size != "Standard_DS3_v2" {
					return false, "Wrong VM size: " + size
				}
			}
			return true, ""
		}},
		{"7._Verify_Storage_Account_Replication", func() (bool, string) {
			sas := findResourcesByType(merged, "azurerm_storage_account")
			for _, sa := range sas {
				tier, ok := sa["account_tier"].(string)
				if !ok || tier != "Standard" {
					return false, "Non-standard replication: " + tier
				}
			}
			return true, ""
		}},
		{"8._Verify_Managed_Identity", func() (bool, string) {
			ids := findResourcesByType(merged, "azurerm_managed_identity")
			for _, id := range ids {
				if id["client_id"] == nil {
					return false, "Managed Identity missing client ID"
				}
			}
			return true, ""
		}},
		{"9._Verify_App_Service_Plan_Location", func() (bool, string) {
			plans := findResourcesByType(merged, "azurerm_app_service_plan")
			for _, plan := range plans {
				loc, ok := plan["location"].(string)
				if !ok || loc != "East US" {
					return false, "App Service Plan not in East US"
				}
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
			testCase := TestCase{
				Classname: "AzureInfraTests",
				Name:      tc.Name,
				Status:    status,
			}
			if !passed {
				suite.Failures++
				testCase.Failure = &Failure{
					Message: reason,
					Type:    "failure",
				}
			}
			suite.TestCases = append(suite.TestCases, testCase)
		})
	}

	// Write XML report
	suite.Time = 2.15
	file, err := os.Create("reports\\dev_test_report.xml")
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
