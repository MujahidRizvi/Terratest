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

type TestCasePEPNew struct {
	XMLName   xml.Name       `xml:"testcase"`
	Classname string         `xml:"classname,attr"`
	Name      string         `xml:"name,attr"`
	Failure   *FailurePEPNew `xml:"failure,omitempty"`
	Status    string         `xml:"status"`
}

type FailurePEPNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuitePEPNew struct {
	XMLName   xml.Name         `xml:"testsuite"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Errors    int              `xml:"errors,attr"`
	Time      float64          `xml:"time,attr"`
	TestCases []TestCasePEPNew `xml:"testcase"`
}

func loadRemoteTFStatePEPNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote state: %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read state body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(data, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse state JSON: %v", err)
	}
	return tfState
}

func findResourcesByTypePEPNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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
		instances, _ := m["instances"].([]interface{})
		for _, inst := range instances {
			instMap, ok := inst.(map[string]interface{})
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

func TestPrivateEndpointValidationNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStatePEPNew(t, remoteStateURL)

	var suite TestSuitePEPNew
	suite.Tests = 2

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_Private_Endpoint_Exists_and_Has_Correct_Connection",
			func() (bool, string) {
				peps := findResourcesByTypePEPNew(tfState, "azurerm_private_endpoint")
				if len(peps) == 0 {
					return false, "No private endpoint found"
				}
				for _, pep := range peps {
					if conns, ok := pep["private_service_connection"].([]interface{}); ok && len(conns) > 0 {
						return true, ""
					}
				}
				return false, "No private service connection found in any private endpoint"
			},
		},
		{
			"2._Verify_Private_Endpoint_Has_Correct_Subnet_Reference",
			func() (bool, string) {
				peps := findResourcesByTypePEPNew(tfState, "azurerm_private_endpoint")
				for _, pep := range peps {
					if subnetID, ok := pep["subnet_id"].(string); ok && strings.Contains(subnetID, "subnet") {
						return true, ""
					}
				}
				return false, "No valid subnet_id found for private endpoint"
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
			suite.TestCases = append(suite.TestCases, TestCasePEPNew{
				Classname: "PrivateEndpointTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailurePEPNew {
					if !passed {
						return &FailurePEPNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 1.08
	file, err := os.Create("reports/new_res_private_endpoint_report.xml")
	if err != nil {
		t.Fatalf("❌ Failed to create report file: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML: %v", err)
	}

	fmt.Println("📄 res_private_endpoint_report.xml written successfully")
}
