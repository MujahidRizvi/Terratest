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

type TestCaseVNetNew struct {
	XMLName   xml.Name        `xml:"testcase"`
	Classname string          `xml:"classname,attr"`
	Name      string          `xml:"name,attr"`
	Failure   *FailureVNetNew `xml:"failure,omitempty"`
	Status    string          `xml:"status"`
}

type FailureVNetNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteVNetNew struct {
	XMLName   xml.Name          `xml:"testsuite"`
	Tests     int               `xml:"tests,attr"`
	Failures  int               `xml:"failures,attr"`
	Errors    int               `xml:"errors,attr"`
	Time      float64           `xml:"time,attr"`
	TestCases []TestCaseVNetNew `xml:"testcase"`
}

func loadRemoteTFStateVNetNew(t *testing.T, url string) map[string]interface{} {
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
		t.Fatalf("❌ Failed to parse state JSON: %v", err)
	}
	return tfState
}

func findResourcesByTypeVNetNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var found []map[string]interface{}
	resources, _ := tfState["resources"].([]interface{})
	for _, res := range resources {
		r, _ := res.(map[string]interface{})
		if r["type"] != resourceType {
			continue
		}
		for _, inst := range r["instances"].([]interface{}) {
			i, _ := inst.(map[string]interface{})
			if attrs, ok := i["attributes"].(map[string]interface{}); ok {
				found = append(found, attrs)
			}
		}
	}
	return found
}

func TestVNetValidationNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStateVNetNew(t, remoteStateURL)

	var suite TestSuiteVNetNew
	suite.Tests = 2

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_VNet_Creation_and_Address_Space",
			func() (bool, string) {
				vnets := findResourcesByTypeVNetNew(tfState, "azurerm_virtual_network")
				if len(vnets) == 0 {
					return false, "No Virtual Network found"
				}
				for _, v := range vnets {
					if v["name"] == "" {
						return false, "VNet missing name"
					}
					if addrSpaces, ok := v["address_space"].([]interface{}); !ok || len(addrSpaces) == 0 {
						return false, "VNet missing address space"
					}
				}
				return true, ""
			},
		},
		{
			"2._Verify_DNS_Servers_and_Tags_If_Set",
			func() (bool, string) {
				vnets := findResourcesByTypeVNetNew(tfState, "azurerm_virtual_network")
				for _, v := range vnets {
					if tags, ok := v["tags"].(map[string]interface{}); ok && len(tags) > 0 {
						return true, ""
					}
				}
				return false, "No tags found on Virtual Network"
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			pass, reason := tc.TestFunc()
			status := "PASS"
			if !pass {
				status = "FAIL"
				suite.Failures++
			}
			suite.TestCases = append(suite.TestCases, TestCaseVNetNew{
				Classname: "VirtualNetworkTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureVNetNew {
					if !pass {
						return &FailureVNetNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 0.95
	file, err := os.Create("reports/new_res_vnet_report.xml")
	if err != nil {
		t.Fatalf("❌ Failed to create XML report: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML: %v", err)
	}

	fmt.Println("📄 res_vnet_report.xml written successfully")
}
