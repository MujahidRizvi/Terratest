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

type TestCaseSPKNew struct {
	XMLName   xml.Name       `xml:"testcase"`
	Classname string         `xml:"classname,attr"`
	Name      string         `xml:"name,attr"`
	Failure   *FailureSPKNew `xml:"failure,omitempty"`
	Status    string         `xml:"status"`
}

type FailureSPKNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteSPKNew struct {
	XMLName   xml.Name         `xml:"testsuite"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Errors    int              `xml:"errors,attr"`
	Time      float64          `xml:"time,attr"`
	TestCases []TestCaseSPKNew `xml:"testcase"`
}

func loadRemoteTFStateSPKNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch state: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse state: %v", err)
	}
	return tfState
}

func findResourcesByTypeSPKNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var found []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return found
	}
	for _, res := range resources {
		m, ok := res.(map[string]interface{})
		if !ok || m["type"] != resourceType {
			continue
		}
		instances, ok := m["instances"].([]interface{})
		if !ok {
			continue
		}
		for _, i := range instances {
			im, ok := i.(map[string]interface{})
			if !ok {
				continue
			}
			if attrs, ok := im["attributes"].(map[string]interface{}); ok {
				found = append(found, attrs)
			}
		}
	}
	return found
}

func TestSpokeInfraStateBasedNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStateSPKNew(t, remoteStateURL)

	var suite TestSuiteSPKNew
	suite.Tests = 11

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Resource_Group_Existence_and_Properties", func() (bool, string) {
			rgs := findResourcesByTypeSPKNew(tfState, "azurerm_resource_group")
			for _, rg := range rgs {
				name, _ := rg["name"].(string)
				if strings.Contains(name, "spk-rg") {
					if location, _ := rg["location"].(string); location != "" {
						return true, ""
					}
				}
			}
			return false, "spk-rg not found"
		}},
		{"2._Verify_Virtual_Network_Existence", func() (bool, string) {
			vnets := findResourcesByTypeSPKNew(tfState, "azurerm_virtual_network")
			for _, vn := range vnets {
				name, _ := vn["name"].(string)
				if strings.Contains(name, "spoke-vnet") {
					return true, ""
				}
			}
			return false, "spoke-vnet not found"
		}},
		{"3._Verify_Bastion_Network_Security_Group", func() (bool, string) {
			nsgs := findResourcesByTypeSPKNew(tfState, "azurerm_network_security_group")
			for _, nsg := range nsgs {
				name, _ := nsg["name"].(string)
				if strings.Contains(name, "bst-nsg") {
					return true, ""
				}
			}
			return false, "bst-nsg not found"
		}},
		{"4._Verify_Bastion_Subnet_Existence", func() (bool, string) {
			sns := findResourcesByTypeSPKNew(tfState, "azurerm_subnet")
			for _, sn := range sns {
				name, _ := sn["name"].(string)
				if strings.Contains(name, "bst-snet") {
					return true, ""
				}
			}
			return false, "bst-snet not found"
		}},
		{"5._Verify_EXP_NSG_and_Subnet", func() (bool, string) {
			foundNSG, foundSNET := false, false
			for _, nsg := range findResourcesByTypeSPKNew(tfState, "azurerm_network_security_group") {
				if strings.Contains(nsg["name"].(string), "exp-nsg") {
					foundNSG = true
				}
			}
			for _, sn := range findResourcesByTypeSPKNew(tfState, "azurerm_subnet") {
				if strings.Contains(sn["name"].(string), "exp-snet") {
					foundSNET = true
				}
			}
			if foundNSG && foundSNET {
				return true, ""
			}
			return false, "EXP NSG or Subnet not found"
		}},
		{"6._Verify_PROC_NSG_and_Subnet", func() (bool, string) {
			foundNSG, foundSNET := false, false
			for _, nsg := range findResourcesByTypeSPKNew(tfState, "azurerm_network_security_group") {
				if strings.Contains(nsg["name"].(string), "proc-nsg") {
					foundNSG = true
				}
			}
			for _, sn := range findResourcesByTypeSPKNew(tfState, "azurerm_subnet") {
				if strings.Contains(sn["name"].(string), "proc-snet") {
					foundSNET = true
				}
			}
			if foundNSG && foundSNET {
				return true, ""
			}
			return false, "PROC NSG or Subnet not found"
		}},
		{"7._Verify_PROCFAPP_NSG_and_Subnet_with_Delegation", func() (bool, string) {
			foundNSG, foundSNET := false, false
			for _, nsg := range findResourcesByTypeSPKNew(tfState, "azurerm_network_security_group") {
				if strings.Contains(nsg["name"].(string), "procfapp-nsg") {
					foundNSG = true
				}
			}
			for _, sn := range findResourcesByTypeSPKNew(tfState, "azurerm_subnet_with_delegation") {
				if strings.Contains(sn["name"].(string), "procfapp-snet") {
					foundSNET = true
				}
			}
			if foundNSG && foundSNET {
				return true, ""
			}
			return false, "PROCFAPP NSG or delegated subnet not found"
		}},
		{"8._Verify_SYS_NSG_and_Subnet", func() (bool, string) {
			foundNSG, foundSNET := false, false
			for _, nsg := range findResourcesByTypeSPKNew(tfState, "azurerm_network_security_group") {
				if strings.Contains(nsg["name"].(string), "sys-nsg") {
					foundNSG = true
				}
			}
			for _, sn := range findResourcesByTypeSPKNew(tfState, "azurerm_subnet") {
				if strings.Contains(sn["name"].(string), "sys-snet") {
					foundSNET = true
				}
			}
			if foundNSG && foundSNET {
				return true, ""
			}
			return false, "SYS NSG or Subnet not found"
		}},
		{"9._Verify_SYSFAPP_NSG_and_Subnet_with_Delegation", func() (bool, string) {
			foundNSG, foundSNET := false, false
			for _, nsg := range findResourcesByTypeSPKNew(tfState, "azurerm_network_security_group") {
				if strings.Contains(nsg["name"].(string), "sysfapp-nsg") {
					foundNSG = true
				}
			}
			for _, sn := range findResourcesByTypeSPKNew(tfState, "azurerm_subnet_with_delegation") {
				if strings.Contains(sn["name"].(string), "sysfapp-snet") {
					foundSNET = true
				}
			}
			if foundNSG && foundSNET {
				return true, ""
			}
			return false, "SYSFAPP NSG or delegated subnet not found"
		}},
		{"10._Verify_Shared_Resources_NSG_and_Subnet", func() (bool, string) {
			foundNSG, foundSNET := false, false
			for _, nsg := range findResourcesByTypeSPKNew(tfState, "azurerm_network_security_group") {
				if strings.Contains(nsg["name"].(string), "srcs-nsg") {
					foundNSG = true
				}
			}
			for _, sn := range findResourcesByTypeSPKNew(tfState, "azurerm_subnet") {
				if strings.Contains(sn["name"].(string), "srcs-snet") {
					foundSNET = true
				}
			}
			if foundNSG && foundSNET {
				return true, ""
			}
			return false, "Shared NSG or subnet not found"
		}},
		{"11._Verify_EPP_NSG_and_Subnet", func() (bool, string) {
			foundNSG, foundSNET := false, false
			for _, nsg := range findResourcesByTypeSPKNew(tfState, "azurerm_network_security_group") {
				if strings.Contains(nsg["name"].(string), "epp-nsg") {
					foundNSG = true
				}
			}
			for _, sn := range findResourcesByTypeSPKNew(tfState, "azurerm_subnet") {
				if strings.Contains(sn["name"].(string), "epp-snet") {
					foundSNET = true
				}
			}
			if foundNSG && foundSNET {
				return true, ""
			}
			return false, "EPP NSG or Subnet not found"
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
			suite.TestCases = append(suite.TestCases, TestCaseSPKNew{
				Classname: "SpokeInfraTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureSPKNew {
					if !passed {
						return &FailureSPKNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 2.75
	reportFile := "reports//new_mod_spoke_report.xml"
	file, err := os.Create(reportFile)
	if err != nil {
		t.Fatalf("❌ Unable to create report: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML: %v", err)
	}
	fmt.Println("📄 Spoke test report written to", reportFile)
}
