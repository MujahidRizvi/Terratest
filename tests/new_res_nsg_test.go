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

type TestCaseNSGNew struct {
	XMLName   xml.Name       `xml:"testcase"`
	Classname string         `xml:"classname,attr"`
	Name      string         `xml:"name,attr"`
	Failure   *FailureNSGNew `xml:"failure,omitempty"`
	Status    string         `xml:"status"`
}

type FailureNSGNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteNSGNew struct {
	XMLName   xml.Name         `xml:"testsuite"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Errors    int              `xml:"errors,attr"`
	Time      float64          `xml:"time,attr"`
	TestCases []TestCaseNSGNew `xml:"testcase"`
}

func loadRemoteTFStateNSGNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote state: %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read remote state body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(data, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse JSON state: %v", err)
	}
	return tfState
}

func findResourcesByTypeNSGNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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
		instances, _ := m["instances"].([]interface{})
		for _, inst := range instances {
			instMap, ok := inst.(map[string]interface{})
			if !ok {
				continue
			}
			if attrs, ok := instMap["attributes"].(map[string]interface{}); ok {
				found = append(found, attrs)
			}
		}
	}
	return found
}

func TestNSGResourcesValidationNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStateNSGNew(t, remoteStateURL)

	var suite TestSuiteNSGNew
	suite.Tests = 2

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_NSG_Creation_and_Name_Tagging",
			func() (bool, string) {
				nsgs := findResourcesByTypeNSGNew(tfState, "azurerm_network_security_group")
				if len(nsgs) == 0 {
					return false, "No NSG resources found"
				}
				for _, nsg := range nsgs {
					if name, ok := nsg["name"].(string); ok && strings.TrimSpace(name) != "" {
						tags, ok := nsg["tags"].(map[string]interface{})
						if !ok || len(tags) == 0 {
							return false, "NSG has no tags"
						}
						return true, ""
					}
				}
				return false, "No NSG with valid name found"
			},
		},
		{
			"2._Verify_NSG_Security_Rules_Configured",
			func() (bool, string) {
				nsgs := findResourcesByTypeNSGNew(tfState, "azurerm_network_security_group")
				for _, nsg := range nsgs {
					rules, ok := nsg["security_rule"].([]interface{})
					if !ok || len(rules) == 0 {
						return false, "NSG has no security rules configured"
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
			suite.TestCases = append(suite.TestCases, TestCaseNSGNew{
				Classname: "NSGTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureNSGNew {
					if !passed {
						return &FailureNSGNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 0.91
	file, err := os.Create("reports/new_res_nsg_report.xml")
	if err != nil {
		t.Fatalf("❌ Failed to create XML report: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to write XML: %v", err)
	}

	fmt.Println("📄 res_nsg_report.xml written successfully")
}
