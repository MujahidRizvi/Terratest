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

type TestCaseAppGateWayNew struct {
	XMLName   xml.Name              `xml:"testcase"`
	Classname string                `xml:"classname,attr"`
	Name      string                `xml:"name,attr"`
	Failure   *FailureAppGateWayNew `xml:"failure,omitempty"`
	Status    string                `xml:"status"`
}

type FailureAppGateWayNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteAppGateWayNew struct {
	XMLName   xml.Name                `xml:"testsuite"`
	Tests     int                     `xml:"tests,attr"`
	Failures  int                     `xml:"failures,attr"`
	Errors    int                     `xml:"errors,attr"`
	Time      float64                 `xml:"time,attr"`
	TestCases []TestCaseAppGateWayNew `xml:"testcase"`
}

func loadRemoteTFStateAppGateWayNew(t *testing.T, url string) map[string]interface{} {
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

func findResourcesByTypeAppGateWayNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestAppGateWayValidationNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStateAppGateWayNew(t, remoteStateURL)

	var suite TestSuiteAppGateWayNew
	suite.Tests = 5

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Check_Application_Gateway_Exists", func() (bool, string) {
			gw := findResourcesByTypeAppGateWayNew(tfState, "azurerm_application_gateway")
			if len(gw) == 0 {
				return false, "Expected at least one Application Gateway"
			}
			return true, ""
		}},
		{"2._Check_AppGW_Has_HTTP2_Enabled", func() (bool, string) {
			gw := findResourcesByTypeAppGateWayNew(tfState, "azurerm_application_gateway")
			if len(gw) == 0 {
				return false, "No Application Gateway to check HTTP2"
			}
			if http2, ok := gw[0]["enable_http2"].(bool); !ok || !http2 {
				return false, "Expected HTTP2 to be enabled"
			}
			return true, ""
		}},
		{"3._Check_SKU_Name_And_Tier", func() (bool, string) {
			gw := findResourcesByTypeAppGateWayNew(tfState, "azurerm_application_gateway")
			if len(gw) == 0 {
				return false, "No Application Gateway to check SKU"
			}
			sku, ok := gw[0]["sku"].([]interface{})
			if !ok || len(sku) == 0 {
				return false, "SKU block is missing or empty"
			}
			skuMap := sku[0].(map[string]interface{})
			if skuMap["name"] != "Standard_v2" || skuMap["tier"] != "Standard_v2" {
				return false, fmt.Sprintf("Expected SKU name/tier to be Standard_v2, got: %v / %v", skuMap["name"], skuMap["tier"])
			}
			return true, ""
		}},
		{"4._Check_AppGW_IP_Config_Subnet", func() (bool, string) {
			gw := findResourcesByTypeAppGateWayNew(tfState, "azurerm_application_gateway")
			if len(gw) == 0 {
				return false, "No Application Gateway to check IP configuration"
			}
			ipConfigs, ok := gw[0]["gateway_ip_configuration"].([]interface{})
			if !ok || len(ipConfigs) == 0 {
				return false, "No IP configuration found"
			}
			ipConfig := ipConfigs[0].(map[string]interface{})
			if ipConfig["subnet_id"] == nil {
				return false, "Expected subnet_id in IP configuration"
			}
			return true, ""
		}},
		{"5._Check_Tags_Are_Set", func() (bool, string) {
			gw := findResourcesByTypeAppGateWayNew(tfState, "azurerm_application_gateway")
			if len(gw) == 0 {
				return false, "No Application Gateway found"
			}
			tags, ok := gw[0]["tags"].(map[string]interface{})
			if !ok || len(tags) == 0 {
				return false, "Tags block is missing or empty"
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
			suite.TestCases = append(suite.TestCases, TestCaseAppGateWayNew{
				Classname: "AppGateWayTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureAppGateWayNew {
					if !passed {
						return &FailureAppGateWayNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 1.0
	reportFile := "reports/new_res_appgateway_report.xml"
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

	fmt.Println("📄 appgateway_report.xml written successfully")
}
