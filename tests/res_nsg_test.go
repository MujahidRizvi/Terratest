package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

type TestCaseNSG struct {
	XMLName   xml.Name    `xml:"testcase"`
	Classname string      `xml:"classname,attr"`
	Name      string      `xml:"name,attr"`
	Failure   *FailureNSG `xml:"failure,omitempty"`
	Status    string      `xml:"status"`
}

type FailureNSG struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteNSG struct {
	XMLName   xml.Name      `xml:"testsuite"`
	Tests     int           `xml:"tests,attr"`
	Failures  int           `xml:"failures,attr"`
	Errors    int           `xml:"errors,attr"`
	Time      float64       `xml:"time,attr"`
	TestCases []TestCaseNSG `xml:"testcase"`
}

func loadTFStateNSG(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesNSG(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeNSG(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var found []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
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

func TestNSGValidations(t *testing.T) {
	var suite TestSuiteNSG
	suite.Tests = 3

	tfState1 := loadTFStateNSG(t, "../terraform.tfstate")
	tfState2 := loadTFStateNSG(t, "../terra.tfstate")
	merged := mergeStatesNSG(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_NSG_Resource_Creation", func() (bool, string) {
			nsgs := findResourcesByTypeNSG(merged, "azurerm_network_security_group")
			if len(nsgs) == 0 {
				return false, "Expected at least one NSG resource"
			}
			return true, ""
		}},
		{"2._Verify_NSG_Security_Rules_Configuration", func() (bool, string) {
			rules := findResourcesByTypeNSG(merged, "azurerm_network_security_rule")
			if len(rules) == 0 {
				return false, "No NSG security rules found"
			}
			for _, rule := range rules {
				if rule["priority"] == nil || rule["access"] == nil {
					return false, "Missing priority or access in one of the rules"
				}
			}
			return true, ""
		}},
		{"3._Validate_NSG_Rule_Effectiveness", func() (bool, string) {
			rules := findResourcesByTypeNSG(merged, "azurerm_network_security_rule")
			if len(rules) == 0 {
				return false, "No rules to evaluate"
			}
			for _, rule := range rules {
				access := rule["access"].(string)
				if access != "Allow" && access != "Deny" {
					return false, "Rule access must be either Allow or Deny"
				}
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
			testCase := TestCaseNSG{
				Classname: "NSGModuleTests",
				Name:      tc.Name,
				Status:    status,
			}
			if !passed {
				testCase.Failure = &FailureNSG{Message: reason, Type: "failure"}
			}
			suite.TestCases = append(suite.TestCases, testCase)
		})
	}

	suite.Time = 1.23
	reportFile := "reports\\res_nsg.xml"
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
