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

type TestCaseLogAnalyticsNew struct {
	XMLName   xml.Name                `xml:"testcase"`
	Classname string                  `xml:"classname,attr"`
	Name      string                  `xml:"name,attr"`
	Failure   *FailureLogAnalyticsNew `xml:"failure,omitempty"`
	Status    string                  `xml:"status"`
}

type FailureLogAnalyticsNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteLogAnalyticsNew struct {
	XMLName   xml.Name                  `xml:"testsuite"`
	Tests     int                       `xml:"tests,attr"`
	Failures  int                       `xml:"failures,attr"`
	Errors    int                       `xml:"errors,attr"`
	Time      float64                   `xml:"time,attr"`
	TestCases []TestCaseLogAnalyticsNew `xml:"testcase"`
}

func loadRemoteTFStateLogAnalyticsNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote state: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read remote state body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse remote state JSON: %v", err)
	}
	return tfState
}

func findResourcesByTypeLogAnalyticsNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var found []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return found
	}
	for _, res := range resources {
		rMap, ok := res.(map[string]interface{})
		if !ok || rMap["type"] != resourceType {
			continue
		}
		for _, inst := range rMap["instances"].([]interface{}) {
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

func TestLogAnalyticsWorkspaceValidationNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"

	tfState := loadRemoteTFStateLogAnalyticsNew(t, remoteStateURL)

	var suite TestSuiteLogAnalyticsNew
	suite.Tests = 3

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_Log_Analytics_Workspace_Exists_with_Correct_Properties",
			func() (bool, string) {
				ws := findResourcesByTypeLogAnalyticsNew(tfState, "azurerm_log_analytics_workspace")
				if len(ws) == 0 {
					return false, "Expected a Log Analytics Workspace"
				}
				if ws[0]["sku"] != "PerGB2018" || ws[0]["retention_in_days"] != float64(30) {
					return false, fmt.Sprintf("Expected sku=PerGB2018 & retention=30, got sku=%v, retention=%v", ws[0]["sku"], ws[0]["retention_in_days"])
				}
				return true, ""
			},
		},
		{
			"2._Verify_Tags_on_Log_Analytics_Workspace",
			func() (bool, string) {
				ws := findResourcesByTypeLogAnalyticsNew(tfState, "azurerm_log_analytics_workspace")
				if len(ws) == 0 {
					return false, "Workspace not found"
				}
				tags, ok := ws[0]["tags"].(map[string]interface{})
				if !ok || tags["Project"] != "API Ecosystem" {
					return false, "Missing tag: Project=API Ecosystem"
				}
				return true, ""
			},
		},
		{
			"3._Verify_Terraform_Outputs_for_Workspace_Name_and_ID",
			func() (bool, string) {
				outputs, ok := tfState["outputs"].(map[string]interface{})
				if !ok {
					return false, "Outputs block not found"
				}
				idOut, ok1 := outputs["log_analytics_workspace_id"].(map[string]interface{})
				nameOut, ok2 := outputs["log_analytics_workspace_name"].(map[string]interface{})
				if !ok1 || idOut["value"] == "" {
					return false, "Missing or empty log_analytics_workspace_id"
				}
				if !ok2 || nameOut["value"] == "" {
					return false, "Missing or empty log_analytics_workspace_name"
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
			suite.TestCases = append(suite.TestCases, TestCaseLogAnalyticsNew{
				Classname: "LogAnalyticsTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureLogAnalyticsNew {
					if !passed {
						return &FailureLogAnalyticsNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 1.1
	file, err := os.Create("reports/new_res_log_analytics_new.xml")
	if err != nil {
		t.Fatalf("❌ Failed to create XML file: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to write XML: %v", err)
	}

	fmt.Println("📄 res_log_analytics_new.xml written successfully")
}
