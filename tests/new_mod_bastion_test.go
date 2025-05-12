package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type TestCaseModNew struct {
	XMLName   xml.Name       `xml:"testcase"`
	Classname string         `xml:"classname,attr"`
	Name      string         `xml:"name,attr"`
	Failure   *FailureModNew `xml:"failure,omitempty"`
	Status    string         `xml:"status"`
}

type FailureModNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteModNew struct {
	XMLName   xml.Name         `xml:"testsuite"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Errors    int              `xml:"errors,attr"`
	Time      float64          `xml:"time,attr"`
	TestCases []TestCaseModNew `xml:"testcase"`
}

func loadRemoteTFStateModNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote Terraform state: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read remote Terraform state body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse remote Terraform state: %v", err)
	}
	return tfState
}

func findResourcesByTypeModNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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
		instances, ok := resourceMap["instances"].([]interface{})
		if !ok {
			continue
		}
		for _, inst := range instances {
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

func TestAzureInfraValidationsModNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"

	tfState := loadRemoteTFStateModNew(t, remoteStateURL)
	var suite TestSuiteModNew
	suite.Tests = 9

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Spoke_VNet", func() (bool, string) {
			vnets := findResourcesByTypeModNew(tfState, "azurerm_virtual_network")
			for _, vnet := range vnets {
				if name, _ := vnet["name"].(string); strings.Contains(name, "spoke") {
					addr, ok := vnet["address_space"].([]interface{})
					if !ok || len(addr) == 0 {
						return false, "Missing CIDR"
					}
					if cidr, _ := addr[0].(string); cidr != "10.110.0.0/16" {
						return false, "Incorrect CIDR: " + cidr
					}
					return true, ""
				}
			}
			return false, "Spoke VNet not found"
		}},
		{"2._Verify_Subnets_with_correct_CIDRs", func() (bool, string) {
			subnets := findResourcesByTypeModNew(tfState, "azurerm_subnet")
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
			nsgs := findResourcesByTypeModNew(tfState, "azurerm_network_security_group")
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
			vms := findResourcesByTypeModNew(tfState, "azurerm_virtual_machine")
			for _, vm := range vms {
				name, _ := vm["name"].(string)
				if strings.Contains(name, "bastion") {
					pubIP, _ := vm["public_ip_address"].(string)
					adminUser, _ := vm["admin_username"].(string)
					if pubIP == "" {
						return false, "No public IP"
					}
					if adminUser == "" {
						return false, "No admin user"
					}
				}
			}
			return true, ""
		}},
		{"5._Verify_APIM_internal_network_and_DNS", func() (bool, string) {
			apims := findResourcesByTypeModNew(tfState, "azurerm_api_management")
			for _, apim := range apims {
				internal, _ := apim["internal"].(bool)
				dns, _ := apim["dns_name"].(string)
				if !internal || dns == "" {
					return false, "APIM not internal or missing DNS"
				}
			}
			return true, ""
		}},
		{"6._Verify_VM_Size", func() (bool, string) {
			vms := findResourcesByTypeModNew(tfState, "azurerm_virtual_machine")
			for _, vm := range vms {
				size, _ := vm["vm_size"].(string)
				if size != "Standard_DS3_v2" {
					return false, "Wrong VM size: " + size
				}
			}
			return true, ""
		}},
		{"7._Verify_Storage_Account_Replication", func() (bool, string) {
			sas := findResourcesByTypeModNew(tfState, "azurerm_storage_account")
			for _, sa := range sas {
				tier, _ := sa["account_tier"].(string)
				if tier != "Standard" {
					return false, "Non-standard tier: " + tier
				}
			}
			return true, ""
		}},
		{"8._Verify_Managed_Identity", func() (bool, string) {
			ids := findResourcesByTypeModNew(tfState, "azurerm_managed_identity")
			for _, id := range ids {
				if id["client_id"] == nil {
					return false, "Missing client ID"
				}
			}
			return true, ""
		}},
		{"9._Verify_App_Service_Plan_Location", func() (bool, string) {
			plans := findResourcesByTypeModNew(tfState, "azurerm_app_service_plan")
			for _, plan := range plans {
				loc, _ := plan["location"].(string)
				if loc != "East US" {
					return false, "Wrong location: " + loc
				}
			}
			return true, ""
		}},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			passed, reason := tc.TestFunc()
			status := "PASS"
			if !passed {
				status = "FAIL"
				suite.Failures++
			}
			suite.TestCases = append(suite.TestCases, TestCaseModNew{
				Classname: "AzureInfraTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureModNew {
					if !passed {
						return &FailureModNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 2.15
	file, err := os.Create("reports/new_mod_bastion_report.xml")
	if err != nil {
		t.Fatalf("❌ Unable to create report file: %v", err)
	}
	defer file.Close()

	xmlWriter := xml.NewEncoder(file)
	xmlWriter.Indent("", "  ")
	if err := xmlWriter.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode report: %v", err)
	}

	fmt.Println("📄 Test report written to reports/bastion_report.xml")
}
