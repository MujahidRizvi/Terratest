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

type TestCaseSYSNew struct {
	XMLName   xml.Name       `xml:"testcase"`
	Classname string         `xml:"classname,attr"`
	Name      string         `xml:"name,attr"`
	Failure   *FailureSYSNew `xml:"failure,omitempty"`
	Status    string         `xml:"status"`
}

type FailureSYSNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteSYSNew struct {
	XMLName   xml.Name         `xml:"testsuite"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Errors    int              `xml:"errors,attr"`
	Time      float64          `xml:"time,attr"`
	TestCases []TestCaseSYSNew `xml:"testcase"`
}

func loadRemoteTFStateSYSNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch state: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse JSON: %v", err)
	}
	return tfState
}

func findResourcesByTypeSYSNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestSysStateBasedNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStateSYSNew(t, remoteStateURL)

	var suite TestSuiteSYSNew
	suite.Tests = 3

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_System_Resource_Group_Existence_and_Properties", func() (bool, string) {
			rgs := findResourcesByTypeSYSNew(tfState, "azurerm_resource_group")
			for _, rg := range rgs {
				if strings.Contains(rg["name"].(string), "sys-rg") {
					if loc, _ := rg["location"].(string); loc != "" {
						return true, ""
					}
				}
			}
			return false, "sys-rg not found or invalid"
		}},
		{"2._Verify_Azure_Function_App_Existence_and_Configuration", func() (bool, string) {
			fapps := findResourcesByTypeSYSNew(tfState, "azurerm_windows_function_app")
			for _, fa := range fapps {
				if strings.Contains(fa["name"].(string), "sysReady-fapp") {
					if loc, _ := fa["location"].(string); loc != "" {
						return true, ""
					}
				}
			}
			return false, "sysReady-fapp not found or invalid"
		}},
		{"3._Verify_Private_Endpoint_Existence_and_Configuration", func() (bool, string) {
			peps := findResourcesByTypeSYSNew(tfState, "azurerm_private_endpoint")
			for _, pep := range peps {
				if strings.Contains(strings.ToLower(pep["name"].(string)), "sysfappready-pep") {
					if conn, ok := pep["private_service_connection"].([]interface{}); ok && len(conn) > 0 {
						return true, ""
					}
				}
			}
			return false, "sysfappReady-pep not found or misconfigured"
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
			suite.TestCases = append(suite.TestCases, TestCaseSYSNew{
				Classname: "SystemInfraTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureSYSNew {
					if !passed {
						return &FailureSYSNew{Message: reason, Type: "error"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 1.62
	file, err := os.Create("reports/new_mod_sys_report.xml")
	if err != nil {
		t.Fatalf("❌ Failed to create report file: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML: %v", err)
	}
	fmt.Println("📄 SYS report written to reports/sys_report.xml")
}
