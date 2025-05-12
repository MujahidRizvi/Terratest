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

type TestCaseDNSZoneNew struct {
	XMLName   xml.Name           `xml:"testcase"`
	Classname string             `xml:"classname,attr"`
	Name      string             `xml:"name,attr"`
	Failure   *FailureDNSZoneNew `xml:"failure,omitempty"`
	Status    string             `xml:"status"`
}

type FailureDNSZoneNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteDNSZoneNew struct {
	XMLName   xml.Name             `xml:"testsuite"`
	Tests     int                  `xml:"tests,attr"`
	Failures  int                  `xml:"failures,attr"`
	Errors    int                  `xml:"errors,attr"`
	Time      float64              `xml:"time,attr"`
	TestCases []TestCaseDNSZoneNew `xml:"testcase"`
}

func loadRemoteTFStateDNSZoneNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote Terraform state: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read state body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to unmarshal JSON: %v", err)
	}
	return tfState
}

func findResourcesByTypeDNSZoneNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var results []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return results
	}
	for _, res := range resources {
		r, ok := res.(map[string]interface{})
		if !ok || r["type"] != resourceType {
			continue
		}
		instances, _ := r["instances"].([]interface{})
		for _, i := range instances {
			instMap, ok := i.(map[string]interface{})
			if !ok {
				continue
			}
			if attrs, ok := instMap["attributes"].(map[string]interface{}); ok {
				results = append(results, attrs)
			}
		}
	}
	return results
}

func TestPrivateDNSZoneValidationNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"

	tfState := loadRemoteTFStateDNSZoneNew(t, remoteStateURL)

	var suite TestSuiteDNSZoneNew
	suite.Tests = 2

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_Private_DNS_Zone_Created_with_Proper_Name",
			func() (bool, string) {
				zones := findResourcesByTypeDNSZoneNew(tfState, "azurerm_private_dns_zone")
				if len(zones) == 0 {
					return false, "No private DNS zones found"
				}
				for _, z := range zones {
					if name, _ := z["name"].(string); strings.HasSuffix(name, ".azure-api.net") {
						return true, ""
					}
				}
				return false, "No private DNS zone ends with .azure-api.net"
			},
		},
		{
			"2._Verify_DNS_Zone_Links_and_Tags",
			func() (bool, string) {
				links := findResourcesByTypeDNSZoneNew(tfState, "azurerm_private_dns_zone_virtual_network_link")
				if len(links) == 0 {
					return false, "No virtual network link found for DNS zone"
				}
				for _, l := range links {
					tags, ok := l["tags"].(map[string]interface{})
					if !ok || tags["Environment"] == nil {
						return false, "Missing expected Environment tag"
					}
				}
				return true, ""
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			passed, reason := tc.TestFunc()
			status := "PASS"
			if !passed {
				status = "FAIL"
				suite.Failures++
			}
			suite.TestCases = append(suite.TestCases, TestCaseDNSZoneNew{
				Classname: "PrivateDNSZoneTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureDNSZoneNew {
					if !passed {
						return &FailureDNSZoneNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 1.02
	reportFile := "reports/new_res_pvt_dns_zone_report.xml"
	file, err := os.Create(reportFile)
	if err != nil {
		t.Fatalf("❌ Failed to create XML file: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML report: %v", err)
	}

	fmt.Println("📄 res_pvt_dns_zone_report.xml written successfully")
}
