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

type TestCaseFAPP8New struct {
	XMLName   xml.Name         `xml:"testcase"`
	Classname string           `xml:"classname,attr"`
	Name      string           `xml:"name,attr"`
	Failure   *FailureFAPP8New `xml:"failure,omitempty"`
	Status    string           `xml:"status"`
}

type FailureFAPP8New struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteFAPP8New struct {
	XMLName   xml.Name           `xml:"testsuite"`
	Tests     int                `xml:"tests,attr"`
	Failures  int                `xml:"failures,attr"`
	Errors    int                `xml:"errors,attr"`
	Time      float64            `xml:"time,attr"`
	TestCases []TestCaseFAPP8New `xml:"testcase"`
}

func loadRemoteTFStateFAPP8New(t *testing.T, url string) map[string]interface{} {
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
		t.Fatalf("❌ Failed to unmarshal JSON state: %v", err)
	}
	return tfState
}

func mergeStatesFAPP8New(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeFAPP8New(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestFunctionAppNetCore8ISOValidationsNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"

	tfState1 := loadRemoteTFStateFAPP8New(t, remoteStateURL)
	// In this example, assume that merging two remote states is not necessary, but if you need to merge two sources,
	// you can call mergeStatesFAPP8New with two different calls.
	tfState2 := loadRemoteTFStateFAPP8New(t, remoteStateURL)
	merged := mergeStatesFAPP8New(tfState1, tfState2)

	var suite TestSuiteFAPP8New
	suite.Tests = 4

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Azure_Storage_Account_Creation_and_Configuration", func() (bool, string) {
			sa := findResourcesByTypeFAPP8New(merged, "azurerm_storage_account")
			if len(sa) == 0 {
				return false, "Expected a storage account resource"
			}
			if sa[0]["account_tier"] == "" || sa[0]["account_replication_type"] == "" {
				return false, "Storage account missing configuration"
			}
			return true, ""
		}},
		{"2._Verify_Azure_App_Service_Plan_Creation_and_Configuration", func() (bool, string) {
			asp := findResourcesByTypeFAPP8New(merged, "azurerm_service_plan")
			if len(asp) == 0 {
				return false, "Expected an App Service Plan resource"
			}
			if asp[0]["os_type"] != "Windows" {
				return false, fmt.Sprintf("Expected OS type Windows, got %v", asp[0]["os_type"])
			}
			return true, ""
		}},
		{"3._Verify_Azure_Windows_Function_App_Deployment_and_Settings", func() (bool, string) {
			fn := findResourcesByTypeFAPP8New(merged, "azurerm_windows_function_app")
			if len(fn) == 0 {
				return false, "Expected a Windows Function App resource"
			}
			config, ok := fn[0]["site_config"].([]interface{})
			if !ok || len(config) == 0 {
				return false, "Missing site_config in function app"
			}
			return true, ""
		}},
		{"4._Verify_Virtual_Network_Integration_and_Public_Access_Setting", func() (bool, string) {
			fn := findResourcesByTypeFAPP8New(merged, "azurerm_windows_function_app")
			if len(fn) == 0 {
				return false, "No function app found"
			}
			if fn[0]["virtual_network_subnet_id"] == nil {
				return false, "Function App not integrated with VNet"
			}
			if fn[0]["public_network_access_enabled"] != false {
				return false, "Public network access should be disabled"
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
			suite.TestCases = append(suite.TestCases, TestCaseFAPP8New{
				Classname: "FunctionAppNetCore8ISOTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureFAPP8New {
					if !passed {
						return &FailureFAPP8New{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 1.23

	reportFile := "reports/new_res_function_app_netcore8iso_report.xml"
	file, err := os.Create(reportFile)
	if err != nil {
		t.Fatalf("❌ Failed to create XML report: %v", err)
	}
	defer file.Close()

	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to write XML report: %v", err)
	}

	fmt.Printf("📄 XML test report written to %s\n", reportFile)
}
