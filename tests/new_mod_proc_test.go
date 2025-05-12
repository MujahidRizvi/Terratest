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

type TestCasePROCNew struct {
	XMLName   xml.Name        `xml:"testcase"`
	Classname string          `xml:"classname,attr"`
	Name      string          `xml:"name,attr"`
	Failure   *FailurePROCNew `xml:"failure,omitempty"`
	Status    string          `xml:"status"`
}

type FailurePROCNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuitePROCNew struct {
	XMLName   xml.Name          `xml:"testsuite"`
	Tests     int               `xml:"tests,attr"`
	Failures  int               `xml:"failures,attr"`
	Errors    int               `xml:"errors,attr"`
	Time      float64           `xml:"time,attr"`
	TestCases []TestCasePROCNew `xml:"testcase"`
}

func loadRemoteTFStatePROCNew(t *testing.T, url string) map[string]interface{} {
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
		t.Fatalf("❌ Failed to parse state: %v", err)
	}
	return tfState
}

func findResourcesByTypePROCNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestAzureFunctionProcInfraStateBasedNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStatePROCNew(t, remoteStateURL)

	var suite TestSuitePROCNew
	suite.Tests = 4

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Resource_Group_Existence_and_Properties", func() (bool, string) {
			rgs := findResourcesByTypePROCNew(tfState, "azurerm_resource_group")
			for _, rg := range rgs {
				name, _ := rg["name"].(string)
				if strings.Contains(name, "proc-rg") {
					if location, _ := rg["location"].(string); location != "" {
						return true, ""
					}
				}
			}
			return false, "proc-rg not found or invalid"
		}},
		{"2._Verify_Function_App_Instance", func() (bool, string) {
			funcs := findResourcesByTypePROCNew(tfState, "azurerm_windows_function_app")
			for _, fn := range funcs {
				name, _ := fn["name"].(string)
				if strings.Contains(name, "procReady-fapp") {
					return true, ""
				}
			}
			return false, "procReady-fapp not found"
		}},
		{"3._Verify_Storage_Account_Existence", func() (bool, string) {
			sts := findResourcesByTypePROCNew(tfState, "azurerm_storage_account")
			for _, sa := range sts {
				name, _ := sa["name"].(string)
				if strings.Contains(name, "prfpreadystg") {
					return true, ""
				}
			}
			return false, "prfpreadystg not found"
		}},
		{"4._Verify_Private_Endpoint", func() (bool, string) {
			endpoints := findResourcesByTypePROCNew(tfState, "azurerm_private_endpoint")
			for _, ep := range endpoints {
				name, _ := ep["name"].(string)
				if strings.Contains(name, "prfpReady-pep") {
					return true, ""
				}
			}
			return false, "prfpReady-pep not found"
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
			suite.TestCases = append(suite.TestCases, TestCasePROCNew{
				Classname: "AzureFunctionProcInfraStateBasedTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailurePROCNew {
					if !passed {
						return &FailurePROCNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 1.87
	reportFile := "reports/new_mod_proc_report.xml"
	file, err := os.Create(reportFile)
	if err != nil {
		t.Fatalf("❌ Unable to create test report: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML: %v", err)
	}
	fmt.Println("📄 Test report written to", reportFile)
}
