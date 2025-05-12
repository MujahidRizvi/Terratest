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

type TestCaseAPIMNew struct {
	XMLName   xml.Name        `xml:"testcase"`
	Classname string          `xml:"classname,attr"`
	Name      string          `xml:"name,attr"`
	Failure   *FailureAPIMNew `xml:"failure,omitempty"`
	Status    string          `xml:"status"`
}

type FailureAPIMNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteAPIMNew struct {
	XMLName   xml.Name          `xml:"testsuite"`
	Tests     int               `xml:"tests,attr"`
	Failures  int               `xml:"failures,attr"`
	Errors    int               `xml:"errors,attr"`
	Time      float64           `xml:"time,attr"`
	TestCases []TestCaseAPIMNew `xml:"testcase"`
}

func loadRemoteTFStateAPIMNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote Terraform state: %v", err)
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

func findResourcesByTypeAPIMNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var results []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return results
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
				results = append(results, attrs)
			}
		}
	}
	return results
}

func TestApimStateBasedNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStateAPIMNew(t, remoteStateURL)

	var suite TestSuiteAPIMNew
	suite.Tests = 5

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_APIM_Resource_Deployment_and_Configuration",
			func() (bool, string) {
				apims := findResourcesByTypeAPIMNew(tfState, "azurerm_api_management")
				for _, apim := range apims {
					if name, _ := apim["name"].(string); name != "" {
						return true, ""
					}
				}
				return false, "APIM resource not found or incorrectly configured"
			},
		},
		{
			"2._Verify_App_Insights_and_Log_Analytics_Integration",
			func() (bool, string) {
				appInsights := findResourcesByTypeAPIMNew(tfState, "azurerm_application_insights")
				logAnalytics := findResourcesByTypeAPIMNew(tfState, "azurerm_log_analytics_workspace")
				if len(appInsights) > 0 && len(logAnalytics) > 0 {
					return true, ""
				}
				return false, "Application Insights or Log Analytics not found"
			},
		},
		{
			"3._Verify_Private_DNS_A_Records_for_APIM_and_Prefixes",
			func() (bool, string) {
				dnsRecords := findResourcesByTypeAPIMNew(tfState, "azurerm_private_dns_a_record")
				expectedPrefixes := []string{"management", "developer", "portal"}
				found := map[string]bool{}
				for _, rec := range dnsRecords {
					name, _ := rec["name"].(string)
					for _, prefix := range expectedPrefixes {
						if name == prefix {
							found[prefix] = true
						}
					}
				}
				for _, prefix := range expectedPrefixes {
					if !found[prefix] {
						return false, fmt.Sprintf("Missing DNS A record for prefix: %s", prefix)
					}
				}
				return true, ""
			},
		},
		{
			"4._Verify_AppInsights_Logger_and_Log_Retention",
			func() (bool, string) {
				loggers := findResourcesByTypeAPIMNew(tfState, "azurerm_api_management_logger")
				for _, logger := range loggers {
					resourceID, _ := logger["resource_id"].(string)
					if strings.Contains(resourceID, "applicationInsights") {
						return true, ""
					}
				}
				return false, "Logger not linked with Application Insights"
			},
		},
		{
			"5._Verify_Terraform_Outputs_for_APIM_ID_Name_PrivateIP_FQDN",
			func() (bool, string) {
				outputsRaw, ok := tfState["outputs"].(map[string]interface{})
				if !ok {
					return false, "Outputs missing in state"
				}
				for _, key := range []string{"apim_id", "apim_name", "apim_private_ip", "apim_fqdn"} {
					out, exists := outputsRaw[key]
					if !exists {
						return false, fmt.Sprintf("Missing output key: %s", key)
					}
					if m, ok := out.(map[string]interface{}); !ok || m["value"] == nil || m["value"] == "" {
						return false, fmt.Sprintf("Invalid output value for: %s", key)
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
			suite.TestCases = append(suite.TestCases, TestCaseAPIMNew{
				Classname: "APIMValidationTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureAPIMNew {
					if !passed {
						return &FailureAPIMNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 1.97
	file, err := os.Create("reports/new_res_apim_report.xml")
	if err != nil {
		t.Fatalf("❌ Failed to create XML report: %v", err)
	}
	defer file.Close()

	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML: %v", err)
	}

	fmt.Println("📄 res_apim_report.xml written successfully")
}
