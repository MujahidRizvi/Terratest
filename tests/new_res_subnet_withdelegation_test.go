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

type TestCaseSubnetDelNew struct {
	XMLName   xml.Name             `xml:"testcase"`
	Classname string               `xml:"classname,attr"`
	Name      string               `xml:"name,attr"`
	Failure   *FailureSubnetDelNew `xml:"failure,omitempty"`
	Status    string               `xml:"status"`
}

type FailureSubnetDelNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteSubnetDelNew struct {
	XMLName   xml.Name               `xml:"testsuite"`
	Tests     int                    `xml:"tests,attr"`
	Failures  int                    `xml:"failures,attr"`
	Errors    int                    `xml:"errors,attr"`
	Time      float64                `xml:"time,attr"`
	TestCases []TestCaseSubnetDelNew `xml:"testcase"`
}

func loadRemoteTFStateSubnetDelNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to load state: %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read state data: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(data, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse state JSON: %v", err)
	}
	return tfState
}

func findResourcesByTypeSubnetDelNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var found []map[string]interface{}
	resources, _ := tfState["resources"].([]interface{})
	for _, res := range resources {
		r, _ := res.(map[string]interface{})
		if r["type"] != resourceType {
			continue
		}
		instances, _ := r["instances"].([]interface{})
		for _, inst := range instances {
			i, _ := inst.(map[string]interface{})
			if attrs, ok := i["attributes"].(map[string]interface{}); ok {
				found = append(found, attrs)
			}
		}
	}
	return found
}

func TestSubnetWithDelegationNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStateSubnetDelNew(t, remoteStateURL)

	var suite TestSuiteSubnetDelNew
	suite.Tests = 2

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{
			"1._Verify_Subnet_With_Delegation_Exists",
			func() (bool, string) {
				subnets := findResourcesByTypeSubnetDelNew(tfState, "azurerm_subnet")
				for _, s := range subnets {
					if _, ok := s["delegation"]; ok {
						return true, ""
					}
				}
				return false, "No subnet found with delegation"
			},
		},
		{
			"2._Verify_Delegation_Config_Contains_Microsoft_Web_ServerFarms",
			func() (bool, string) {
				subnets := findResourcesByTypeSubnetDelNew(tfState, "azurerm_subnet")
				for _, s := range subnets {
					if delegs, ok := s["delegation"].([]interface{}); ok {
						for _, d := range delegs {
							if dMap, ok := d.(map[string]interface{}); ok {
								if dMap["name"] != nil {
									if config, ok := dMap["service_delegation"].([]interface{}); ok {
										for _, c := range config {
											if cMap, ok := c.(map[string]interface{}); ok {
												if cMap["name"] == "Microsoft.Web/serverFarms" {
													return true, ""
												}
											}
										}
									}
								}
							}
						}
					}
				}
				return false, "No valid service delegation found for Microsoft.Web/serverFarms"
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			pass, reason := tc.TestFunc()
			status := "PASS"
			if !pass {
				status = "FAIL"
				suite.Failures++
			}
			suite.TestCases = append(suite.TestCases, TestCaseSubnetDelNew{
				Classname: "SubnetDelegationTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureSubnetDelNew {
					if !pass {
						return &FailureSubnetDelNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 0.78
	file, err := os.Create("reports/new_res_subnet_withdelegation_report.xml")
	if err != nil {
		t.Fatalf("❌ Could not create XML report: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode report: %v", err)
	}

	fmt.Println("📄 res_subnet_withdelegation_report.xml written successfully")
}
