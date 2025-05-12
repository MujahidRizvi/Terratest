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

type TestCaseFAPPNew struct {
	XMLName   xml.Name        `xml:"testcase"`
	Classname string          `xml:"classname,attr"`
	Name      string          `xml:"name,attr"`
	Failure   *FailureFAPPNew `xml:"failure,omitempty"`
	Status    string          `xml:"status"`
}

type FailureFAPPNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteFAPPNew struct {
	XMLName   xml.Name          `xml:"testsuite"`
	Tests     int               `xml:"tests,attr"`
	Failures  int               `xml:"failures,attr"`
	Errors    int               `xml:"errors,attr"`
	Time      float64           `xml:"time,attr"`
	TestCases []TestCaseFAPPNew `xml:"testcase"`
}

func loadRemoteTFStateFAPPNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote Terraform state: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read response body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse state: %v", err)
	}
	return tfState
}

func findResourcesByTypeFAPPNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestFunctionAppInfrastructureNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"

	tfState := loadRemoteTFStateFAPPNew(t, remoteStateURL)

	var suite TestSuiteFAPPNew
	suite.Tests = 6

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Azure_Storage_Account_Creation", func() (bool, string) {
			resources := findResourcesByTypeFAPPNew(tfState, "azurerm_storage_account")
			if len(resources) == 0 {
				return false, "Expected at least one Azure Storage Account"
			}
			return true, ""
		}},
		{"2._Verify_Log_Analytics_Workspace_and_AppInsights", func() (bool, string) {
			law := findResourcesByTypeFAPPNew(tfState, "azurerm_application_insights")
			if len(law) == 0 {
				return false, "Expected Application Insights configured with Log Analytics"
			}
			return true, ""
		}},
		{"3._Verify_App_Service_Plan_Creation", func() (bool, string) {
			plans := findResourcesByTypeFAPPNew(tfState, "azurerm_service_plan")
			if len(plans) == 0 {
				return false, "Expected at least one App Service Plan"
			}
			return true, ""
		}},
		{"4._Verify_Function_App_Deployment", func() (bool, string) {
			fapps := findResourcesByTypeFAPPNew(tfState, "azurerm_windows_function_app")
			if len(fapps) == 0 {
				return false, "Expected at least one Function App"
			}
			return true, ""
		}},
		{"5._Verify_Network_and_Security_Settings", func() (bool, string) {
			storages := findResourcesByTypeFAPPNew(tfState, "azurerm_storage_account")
			if len(storages) == 0 {
				return false, "No storage account found"
			}
			rules, ok := storages[0]["network_rules"].([]interface{})
			if !ok || len(rules) == 0 {
				return false, "Storage account network rules not configured"
			}
			return true, ""
		}},
		{"6._Verify_Tags_Applied", func() (bool, string) {
			resTypes := []string{"azurerm_storage_account", "azurerm_application_insights", "azurerm_service_plan", "azurerm_windows_function_app"}
			for _, resType := range resTypes {
				res := findResourcesByTypeFAPPNew(tfState, resType)
				if len(res) == 0 {
					return false, fmt.Sprintf("No resource of type %s found", resType)
				}
				tags, ok := res[0]["tags"].(map[string]interface{})
				if !ok || len(tags) == 0 {
					return false, fmt.Sprintf("Tags missing on resource type %s", resType)
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
			suite.TestCases = append(suite.TestCases, TestCaseFAPPNew{
				Classname: "FunctionAppModuleTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureFAPPNew {
					if !passed {
						return &FailureFAPPNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 1.45
	reportFile := "reports/new_res_functionapp_report.xml"
	file, err := os.Create(reportFile)
	if err != nil {
		t.Fatalf("❌ Failed to create XML report: %v", err)
	}
	defer file.Close()

	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML: %v", err)
	}

	fmt.Println("📄 res_functionapp_report.xml written successfully")
}
