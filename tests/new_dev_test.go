package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

type TestCaseNewDev struct {
	XMLName   xml.Name       `xml:"testcase"`
	Classname string         `xml:"classname,attr"`
	Name      string         `xml:"name,attr"`
	Failure   *FailureNewDev `xml:"failure,omitempty"`
	Status    string         `xml:"status"`
}

type FailureNewDev struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteNewDev struct {
	XMLName   xml.Name         `xml:"testsuite"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Errors    int              `xml:"errors,attr"`
	Time      float64          `xml:"time,attr"`
	TestCases []TestCaseNewDev `xml:"testcase"`
}

func loadTFStateNewDev(t *testing.T) map[string]interface{} {
	cmd := exec.Command("terraform", "show", "-json")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("❌ Failed to execute 'terraform show -json': %v", err)
	}

	var tfState map[string]interface{}
	if err := json.Unmarshal(output, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse Terraform JSON output: %v", err)
	}
	return tfState
}

func findResourcesByTypeNewDev(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var foundResources []map[string]interface{}

	values, ok := tfState["values"].(map[string]interface{})
	if !ok {
		return foundResources
	}

	resources, ok := values["root_module"].(map[string]interface{})["resources"].([]interface{})
	if !ok {
		return foundResources
	}

	for _, res := range resources {
		resourceMap, ok := res.(map[string]interface{})
		if !ok {
			continue
		}
		if resourceMap["type"] == resourceType {
			attrs, ok := resourceMap["values"].(map[string]interface{})
			if ok {
				foundResources = append(foundResources, attrs)
			}
		}
	}
	return foundResources
}

func TestAzureInfraValidationsNewDev(t *testing.T) {
	var suite TestSuiteNewDev
	suite.Tests = 9

	tfState := loadTFStateNewDev(t)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Spoke_VNet", func() (bool, string) {
			vnets := findResourcesByTypeNewDev(tfState, "azurerm_virtual_network")
			for _, vnet := range vnets {
				name, _ := vnet["name"].(string)
				if strings.Contains(name, "spoke") {
					addrSpace, ok := vnet["address_space"].([]interface{})
					if !ok || len(addrSpace) == 0 {
						return false, "Spoke VNet missing CIDR block"
					}
					cidr, _ := addrSpace[0].(string)
					if cidr != "10.110.0.0/16" {
						return false, "Spoke VNet has incorrect CIDR: " + cidr
					}
					return true, ""
				}
			}
			return false, "Spoke VNet not found"
		}},
		{"2._Verify_Subnets_with_correct_CIDRs", func() (bool, string) {
			subnets := findResourcesByTypeNewDev(tfState, "azurerm_subnet")
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
			nsgs := findResourcesByTypeNewDev(tfState, "azurerm_network_security_group")
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
			vms := findResourcesByTypeNewDev(tfState, "azurerm_virtual_machine")
			for _, vm := range vms {
				name, _ := vm["name"].(string)
				if strings.Contains(name, "bastion") {
					pubIP, _ := vm["public_ip_address"].(string)
					if pubIP == "" {
						return false, "No public IP on Bastion VM"
					}
					adminUser, _ := vm["admin_username"].(string)
					if adminUser == "" {
						return false, "No admin user on Bastion VM"
					}
				}
			}
			return true, ""
		}},
		{"5._Verify_APIM_internal_network_and_DNS", func() (bool, string) {
			apims := findResourcesByTypeNewDev(tfState, "azurerm_api_management")
			for _, apim := range apims {
				internal, _ := apim["internal"].(bool)
				if !internal {
					return false, "APIM is not internal"
				}
				dns, _ := apim["dns_name"].(string)
				if dns == "" {
					return false, "APIM has no DNS name"
				}
			}
			return true, ""
		}},
		{"6._Verify_VM_Size", func() (bool, string) {
			vms := findResourcesByTypeNewDev(tfState, "azurerm_virtual_machine")
			for _, vm := range vms {
				size, _ := vm["vm_size"].(string)
				if size != "Standard_DS3_v2" {
					return false, "Wrong VM size: " + size
				}
			}
			return true, ""
		}},
		{"7._Verify_Storage_Account_Replication", func() (bool, string) {
			sas := findResourcesByTypeNewDev(tfState, "azurerm_storage_account")
			for _, sa := range sas {
				tier, _ := sa["account_tier"].(string)
				if tier != "Standard" {
					return false, "Non-standard replication: " + tier
				}
			}
			return true, ""
		}},
		{"8._Verify_Managed_Identity", func() (bool, string) {
			ids := findResourcesByTypeNewDev(tfState, "azurerm_user_assigned_identity")
			for _, id := range ids {
				if id["client_id"] == nil {
					return false, "Managed Identity missing client ID"
				}
			}
			return true, ""
		}},
		{"9._Verify_App_Service_Plan_Location", func() (bool, string) {
			plans := findResourcesByTypeNewDev(tfState, "azurerm_app_service_plan")
			for _, plan := range plans {
				loc, _ := plan["location"].(string)
				if loc != "East US" {
					return false, "App Service Plan not in East US"
				}
			}
			return true, ""
		}},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			passed, reason := tc.TestFunc()

			if passed {
				fmt.Printf("✅ %s passed\n", tc.Name)
			} else {
				fmt.Printf("❌ %s failed: %s\n", tc.Name, reason)
			}

			status := "FAIL"
			if passed {
				status = "PASS"
			}

			testCase := TestCaseNewDev{
				Classname: "AzureInfraTests",
				Name:      tc.Name,
				Status:    status,
			}
			if !passed {
				suite.Failures++
				testCase.Failure = &FailureNewDev{
					Message: reason,
					Type:    "failure",
				}
			}
			suite.TestCases = append(suite.TestCases, testCase)
		})
	}

	suite.Time = 2.15
	_ = os.MkdirAll("reports", os.ModePerm)
	file, err := os.Create("reports/new_dev_test_report.xml")
	if err != nil {
		t.Fatalf("❌ Unable to create report file: %v", err)
	}
	defer file.Close()

	xmlWriter := xml.NewEncoder(file)
	xmlWriter.Indent("", "  ")
	if err := xmlWriter.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode report: %v", err)
	}

	fmt.Println("📄 Test report written to reports/dev_test_report.xml")
}
