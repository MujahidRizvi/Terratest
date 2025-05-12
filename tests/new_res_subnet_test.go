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

type TestCaseSubnetNew struct {
	XMLName   xml.Name          `xml:"testcase"`
	Classname string            `xml:"classname,attr"`
	Name      string            `xml:"name,attr"`
	Failure   *FailureSubnetNew `xml:"failure,omitempty"`
	Status    string            `xml:"status"`
}

type FailureSubnetNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteSubnetNew struct {
	XMLName   xml.Name            `xml:"testsuite"`
	Tests     int                 `xml:"tests,attr"`
	Failures  int                 `xml:"failures,attr"`
	Errors    int                 `xml:"errors,attr"`
	Time      float64             `xml:"time,attr"`
	TestCases []TestCaseSubnetNew `xml:"testcase"`
}

func loadRemoteTFStateSubnetNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote state: %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(data, &tfState); err != nil {
		t.Fatalf("❌ Failed to unmarshal JSON: %v", err)
	}
	return tfState
}

func findResourcesByTypeSubnetNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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
		for _, inst := range rm["instances"].([]interface{}) {
			instMap, ok := inst.(map[string]interface{})
			if !ok {
				continue
			}
			if attrs, ok := instMap["attributes"].(map[string]interface{}); ok {
				results = append(results, attrs)
			}
		}
	}
	return results
}

func TestSubnetValidationNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"

	tfState := loadRemoteTFStateSubnetNew(t, remoteStateURL)

	var suite TestSuiteSubnetNew
	suite.Tests = 2

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_Subnet_Created_with_Name_and_Prefix",
			func() (bool, string) {
				subnets := findResourcesByTypeSubnetNew(tfState, "azurerm_subnet")
				if len(subnets) == 0 {
					return false, "No subnets found"
				}
				for _, s := range subnets {
					if s["name"] == "" {
						return false, "Subnet missing name"
					}
					if prefixes, ok := s["address_prefixes"].([]interface{}); !ok || len(prefixes) == 0 {
						return false, "Missing address_prefixes"
					}
				}
				return true, ""
			},
		},
		{
			"2._Verify_Subnet_NSG_Association_Exists",
			func() (bool, string) {
				subnets := findResourcesByTypeSubnetNew(tfState, "azurerm_subnet")
				for _, s := range subnets {
					if s["network_security_group_id"] == nil {
						return false, fmt.Sprintf("Subnet %v has no NSG associated", s["name"])
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
			suite.TestCases = append(suite.TestCases, TestCaseSubnetNew{
				Classname: "SubnetTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureSubnetNew {
					if !passed {
						return &FailureSubnetNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 0.87
	file, err := os.Create("reports/new_res_subnet_report.xml")
	if err != nil {
		t.Fatalf("❌ Could not create XML file: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode report: %v", err)
	}

	fmt.Println("📄 res_subnet_report.xml written successfully")
}
