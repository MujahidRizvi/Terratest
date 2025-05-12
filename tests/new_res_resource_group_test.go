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

type TestCaseRGNew struct {
	XMLName   xml.Name      `xml:"testcase"`
	Classname string        `xml:"classname,attr"`
	Name      string        `xml:"name,attr"`
	Failure   *FailureRGNew `xml:"failure,omitempty"`
	Status    string        `xml:"status"`
}

type FailureRGNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteRGNew struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Time      float64         `xml:"time,attr"`
	TestCases []TestCaseRGNew `xml:"testcase"`
}

func loadRemoteTFStateRGNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Could not fetch remote state: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read state body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse state JSON: %v", err)
	}
	return tfState
}

func findResourcesByTypeRGNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var out []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return out
	}
	for _, r := range resources {
		rm, ok := r.(map[string]interface{})
		if !ok || rm["type"] != resourceType {
			continue
		}
		for _, inst := range rm["instances"].([]interface{}) {
			if im, ok := inst.(map[string]interface{}); ok {
				if attrs, ok := im["attributes"].(map[string]interface{}); ok {
					out = append(out, attrs)
				}
			}
		}
	}
	return out
}

func TestResourceGroupValidationNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"

	tfState := loadRemoteTFStateRGNew(t, remoteStateURL)

	var suite TestSuiteRGNew
	suite.Tests = 2

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_Resource_Group_Exists_with_Name_and_Location",
			func() (bool, string) {
				rgs := findResourcesByTypeRGNew(tfState, "azurerm_resource_group")
				if len(rgs) == 0 {
					return false, "No resource group found"
				}
				for _, rg := range rgs {
					if rg["name"] != "" && rg["location"] != "" {
						return true, ""
					}
				}
				return false, "Resource group missing name or location"
			},
		},
		{
			"2._Verify_Tags_on_Resource_Group",
			func() (bool, string) {
				rgs := findResourcesByTypeRGNew(tfState, "azurerm_resource_group")
				for _, rg := range rgs {
					if tags, ok := rg["tags"].(map[string]interface{}); ok {
						if tags["Project"] != nil && tags["Environment"] != nil {
							return true, ""
						}
					}
				}
				return false, "Resource group missing expected tags"
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
			suite.TestCases = append(suite.TestCases, TestCaseRGNew{
				Classname: "ResourceGroupTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureRGNew {
					if !pass {
						return &FailureRGNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 0.92
	file, err := os.Create("reports/new_res_resource_group_report.xml")
	if err != nil {
		t.Fatalf("❌ Failed to create XML file: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML: %v", err)
	}

	fmt.Println("📄 res_rg_report.xml written successfully")
}
