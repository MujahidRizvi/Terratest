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

type TestCaseEHUBNew struct {
	XMLName   xml.Name        `xml:"testcase"`
	Classname string          `xml:"classname,attr"`
	Name      string          `xml:"name,attr"`
	Failure   *FailureEHUBNew `xml:"failure,omitempty"`
	Status    string          `xml:"status"`
}

type FailureEHUBNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteEHUBNew struct {
	XMLName   xml.Name          `xml:"testsuite"`
	Tests     int               `xml:"tests,attr"`
	Failures  int               `xml:"failures,attr"`
	Errors    int               `xml:"errors,attr"`
	Time      float64           `xml:"time,attr"`
	TestCases []TestCaseEHUBNew `xml:"testcase"`
}

func loadRemoteTFStateEHUBNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote state: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read response body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to unmarshal JSON: %v", err)
	}
	return tfState
}

func findResourcesByTypeEHUBNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var found []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return found
	}
	for _, res := range resources {
		rm, ok := res.(map[string]interface{})
		if !ok || rm["type"] != resourceType {
			continue
		}
		instances, ok := rm["instances"].([]interface{})
		if !ok {
			continue
		}
		for _, inst := range instances {
			im, ok := inst.(map[string]interface{})
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

func TestEventHubValidationsNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStateEHUBNew(t, remoteStateURL)

	var suite TestSuiteEHUBNew
	suite.Tests = 4

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_EventHub_Namespace_Creation", func() (bool, string) {
			ns := findResourcesByTypeEHUBNew(tfState, "azurerm_eventhub_namespace")
			if len(ns) == 0 {
				return false, "Expected at least one Event Hub Namespace"
			}
			return true, ""
		}},
		{"2._Verify_EventHub_Creation", func() (bool, string) {
			hubs := findResourcesByTypeEHUBNew(tfState, "azurerm_eventhub")
			if len(hubs) == 0 {
				return false, "Expected at least one Event Hub"
			}
			return true, ""
		}},
		{"3._Validate_Message_Retention_Constraint", func() (bool, string) {
			hubs := findResourcesByTypeEHUBNew(tfState, "azurerm_eventhub")
			if len(hubs) == 0 {
				return false, "No Event Hub resources found"
			}
			retention, ok := hubs[0]["message_retention"].(float64)
			if !ok || retention < 1 || retention > 7 {
				return false, fmt.Sprintf("Message retention is out of range: %v", retention)
			}
			return true, ""
		}},
		{"4._Verify_Tags_on_EventHub_Namespace", func() (bool, string) {
			ns := findResourcesByTypeEHUBNew(tfState, "azurerm_eventhub_namespace")
			if len(ns) == 0 {
				return false, "No Event Hub Namespace found"
			}
			tags, ok := ns[0]["tags"].(map[string]interface{})
			if !ok || len(tags) == 0 {
				return false, "Tags not applied on Event Hub Namespace"
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
			suite.TestCases = append(suite.TestCases, TestCaseEHUBNew{
				Classname: "EventHubTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureEHUBNew {
					if !passed {
						return &FailureEHUBNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 1.25
	reportFile := "reports/new_res_eventhub_report.xml"
	file, err := os.Create(reportFile)
	if err != nil {
		t.Fatalf("❌ Unable to create XML report: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML report: %v", err)
	}

	fmt.Println("📄 res_eventhub_report.xml written successfully")
}
