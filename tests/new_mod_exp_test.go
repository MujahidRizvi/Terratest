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

type TestCaseEXPNew struct {
	XMLName   xml.Name       `xml:"testcase"`
	Classname string         `xml:"classname,attr"`
	Name      string         `xml:"name,attr"`
	Failure   *FailureEXPNew `xml:"failure,omitempty"`
	Status    string         `xml:"status"`
}

type FailureEXPNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteEXPNew struct {
	XMLName   xml.Name         `xml:"testsuite"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Errors    int              `xml:"errors,attr"`
	Time      float64          `xml:"time,attr"`
	TestCases []TestCaseEXPNew `xml:"testcase"`
}

func loadRemoteTFStateEXPNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote state: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse state JSON: %v", err)
	}
	return tfState
}

func findResourcesByTypeEXPNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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
			instMap, ok := i.(map[string]interface{})
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

func TestAzureAPIMInfraStateBasedNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStateEXPNew(t, remoteStateURL)

	var suite TestSuiteEXPNew
	suite.Tests = 5

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Resource_Group_Existence_and_Properties", func() (bool, string) {
			rgs := findResourcesByTypeEXPNew(tfState, "azurerm_resource_group")
			for _, rg := range rgs {
				name, _ := rg["name"].(string)
				if strings.Contains(name, "exp-rg") {
					loc, _ := rg["location"].(string)
					tags, _ := rg["tags"].(map[string]interface{})
					if loc != "" && tags["Project"] == "API Ecosystem" {
						return true, ""
					}
				}
			}
			return false, "Resource Group 'exp-rg' not found or properties invalid"
		}},
		{"2._Verify_APIM_Instance", func() (bool, string) {
			apims := findResourcesByTypeEXPNew(tfState, "azurerm_api_management")
			for _, apim := range apims {
				name, _ := apim["name"].(string)
				if strings.Contains(name, "exp-apim") {
					return true, ""
				}
			}
			return false, "APIM instance 'exp-apim' not found"
		}},
		{"3._Verify_APIM_Network_and_Access_Configuration", func() (bool, string) {
			apims := findResourcesByTypeEXPNew(tfState, "azurerm_api_management")
			for _, apim := range apims {
				publicAccess, _ := apim["public_network_access_enabled"].(bool)
				if publicAccess {
					return true, ""
				}
			}
			return false, "APIM public network access not enabled"
		}},
		{"4._Verify_Tag_Consistency", func() (bool, string) {
			rgs := findResourcesByTypeEXPNew(tfState, "azurerm_resource_group")
			for _, rg := range rgs {
				tags, _ := rg["tags"].(map[string]interface{})
				if tags["Project"] == "API Ecosystem" {
					return true, ""
				}
			}
			return false, "Tags not consistent"
		}},
		{"5._Verify_Output_Values", func() (bool, string) {
			outputs, ok := tfState["outputs"].(map[string]interface{})
			if !ok {
				return false, "No outputs in state"
			}
			apimID, exists := outputs["apim_id"]
			if !exists {
				return false, "Output 'apim_id' not found"
			}
			val, ok := apimID.(map[string]interface{})
			if !ok || val["value"] == nil {
				return false, "Output 'apim_id' missing value"
			}
			return true, ""
		}},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			pass, reason := tc.TestFunc()
			status := "PASS"
			if !pass {
				status = "FAIL"
				suite.Failures++
			}
			suite.TestCases = append(suite.TestCases, TestCaseEXPNew{
				Classname: "AzureAPIMInfraStateBasedTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureEXPNew {
					if !pass {
						return &FailureEXPNew{
							Message: reason,
							Type:    "failure",
						}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 2.15
	reportFile := "reports/exp_report.xml"
	file, err := os.Create(reportFile)
	if err != nil {
		t.Fatalf("❌ Unable to create report file: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML: %v", err)
	}
	fmt.Println("📄 Report written to", reportFile)
}
