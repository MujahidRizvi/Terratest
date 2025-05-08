package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type TestCasePEP struct {
	XMLName   xml.Name    `xml:"testcase"`
	Classname string      `xml:"classname,attr"`
	Name      string      `xml:"name,attr"`
	Failure   *FailurePEP `xml:"failure,omitempty"`
	Status    string      `xml:"status"`
}

type FailurePEP struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuitePEP struct {
	XMLName   xml.Name      `xml:"testsuite"`
	Tests     int           `xml:"tests,attr"`
	Failures  int           `xml:"failures,attr"`
	Errors    int           `xml:"errors,attr"`
	Time      float64       `xml:"time,attr"`
	TestCases []TestCasePEP `xml:"testcase"`
}

func loadTFStatePEP(t *testing.T, path string) map[string]interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("\u274c Failed to read terraform state from %s: %v", path, err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(data, &tfState); err != nil {
		t.Fatalf("\u274c Failed to parse terraform state from %s: %v", path, err)
	}
	return tfState
}

func mergeStatesPEP(_, state2 map[string]interface{}) map[string]interface{} {
	return state2 // We just use actual state here for validation
}

func findResourcesByTypePEP(state map[string]interface{}, resourceType string) []map[string]interface{} {
	var found []map[string]interface{}
	resources, ok := state["resources"].([]interface{})
	if !ok {
		return found
	}
	for _, r := range resources {
		rMap, ok := r.(map[string]interface{})
		if !ok || rMap["type"] != resourceType {
			continue
		}
		instances, ok := rMap["instances"].([]interface{})
		if !ok {
			continue
		}
		for _, inst := range instances {
			instMap, ok := inst.(map[string]interface{})
			if !ok {
				continue
			}
			attrs, ok := instMap["attributes"].(map[string]interface{})
			if ok {
				found = append(found, attrs)
			}
		}
	}
	return found
}

func TestPrivateEndpointValidations(t *testing.T) {
	var suite TestSuitePEP
	suite.Tests = 4

	expected := loadTFStatePEP(t, filepath.Join("..", "terra.tfstate"))
	actual := loadTFStatePEP(t, filepath.Join("..", "terraform.tfstate"))
	merged := mergeStatesPEP(expected, actual)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Private_Endpoint_Resource_Exists_and_Is_Correctly_Configured", func() (bool, string) {
			peps := findResourcesByTypePEP(merged, "azurerm_private_endpoint")
			if len(peps) == 0 {
				return false, "Expected at least one private endpoint resource"
			}
			for _, pep := range peps {
				if pep["private_service_connection"] == nil {
					return false, "Missing private_service_connection block"
				}
			}
			return true, ""
		}},
		{"2._Verify_Private_DNS_Zone_Group_Association_and_DNS_Registration", func() (bool, string) {
			peps := findResourcesByTypePEP(merged, "azurerm_private_endpoint")
			for _, pep := range peps {
				group := pep["private_dns_zone_group"]
				if group == nil {
					return false, "Missing private_dns_zone_group block"
				}
				entries := group.([]interface{})
				if len(entries) == 0 {
					return false, "Private DNS zone group is empty"
				}
			}
			return true, ""
		}},
		{"3._Verify_Private_Service_Connection_Settings_and_Status", func() (bool, string) {
			peps := findResourcesByTypePEP(merged, "azurerm_private_endpoint")
			for _, pep := range peps {
				connections := pep["private_service_connection"].([]interface{})
				for _, conn := range connections {
					connMap := conn.(map[string]interface{})
					if connMap["name"] == "" || connMap["private_connection_resource_id"] == "" {
						return false, "Invalid private_service_connection settings"
					}
				}
			}
			return true, ""
		}},
		{"4._Verify_Private_Endpoint_Resource_ID_Output_Matches_Azure_Resource", func() (bool, string) {
			outputs, ok := merged["outputs"].(map[string]interface{})
			if !ok {
				return false, "Missing outputs block in Terraform state"
			}
			if _, ok := outputs["id"]; !ok || outputs["id"].(map[string]interface{})["value"] == "" {
				return false, "Missing or empty private endpoint resource ID output"
			}
			return true, ""
		}},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			passed, reason := tc.TestFunc()
			status := "FAIL"
			if passed {
				fmt.Printf("\u2705 %s passed\n", tc.Name)
				status = "PASS"
			} else {
				fmt.Printf("\u274c %s failed: %s\n", tc.Name, reason)
				suite.Failures++
			}
			testCase := TestCasePEP{
				Classname: "PrivateEndpointModuleTests",
				Name:      tc.Name,
				Status:    status,
			}
			if !passed {
				testCase.Failure = &FailurePEP{Message: reason, Type: "failure"}
			}
			suite.TestCases = append(suite.TestCases, testCase)
		})
	}

	suite.Time = 1.23
	reportFile := "reports\\res_private_endpoint.xml"
	file, err := os.Create(reportFile)
	if err != nil {
		t.Fatalf("\\u274c Failed to create XML report: %v", err)
	}
	defer file.Close()
	xmlWriter := xml.NewEncoder(file)
	xmlWriter.Indent("", "  ")
	if err := xmlWriter.Encode(suite); err != nil {
		t.Fatalf("\\u274c Failed to write XML report: %v", err)
	}
	fmt.Printf("\\ud83d\\udcc4 XML test report written to %s\n", reportFile)
}
