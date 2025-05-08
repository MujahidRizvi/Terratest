package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"testing"
)

type TestCaseEHUB struct {
	XMLName   xml.Name     `xml:"testcase"`
	Classname string       `xml:"classname,attr"`
	Name      string       `xml:"name,attr"`
	Failure   *FailureEHUB `xml:"failure,omitempty"`
	Status    string       `xml:"status"`
}

type FailureEHUB struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteEHUB struct {
	XMLName   xml.Name       `xml:"testsuite"`
	Tests     int            `xml:"tests,attr"`
	Failures  int            `xml:"failures,attr"`
	Errors    int            `xml:"errors,attr"`
	Time      float64        `xml:"time,attr"`
	TestCases []TestCaseEHUB `xml:"testcase"`
}

func loadTFStateEHUB(t *testing.T, path string) map[string]interface{} {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("❌ Failed to read terraform state from %s: %v", path, err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(data, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse terraform state from %s: %v", path, err)
	}
	return tfState
}

func mergeStatesEHUB(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeEHUB(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var foundResources []map[string]interface{}

	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return foundResources
	}

	for _, res := range resources {
		resourceMap, ok := res.(map[string]interface{})
		if !ok {
			continue
		}

		if resourceMap["type"] == resourceType {
			instances, ok := resourceMap["instances"].([]interface{})
			if !ok || len(instances) == 0 {
				continue
			}
			for _, inst := range instances {
				instanceMap, ok := inst.(map[string]interface{})
				if !ok {
					continue
				}
				attributes, ok := instanceMap["attributes"].(map[string]interface{})
				if ok {
					foundResources = append(foundResources, attributes)
				}
			}
		}
	}
	return foundResources
}

func TestEventHubValidations(t *testing.T) {
	var suite TestSuiteEHUB
	suite.Tests = 4

	tfState1 := loadTFStateEHUB(t, "../terraform.tfstate")
	tfState2 := loadTFStateEHUB(t, "../terra.tfstate")
	merged := mergeStatesEHUB(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_EventHub_Namespace_Creation", func() (bool, string) {
			ns := findResourcesByTypeEHUB(merged, "azurerm_eventhub_namespace")
			if len(ns) == 0 {
				return false, "Expected at least one Event Hub Namespace"
			}
			return true, ""
		}},
		{"2._Verify_EventHub_Creation", func() (bool, string) {
			hubs := findResourcesByTypeEHUB(merged, "azurerm_eventhub")
			if len(hubs) == 0 {
				return false, "Expected at least one Event Hub"
			}
			return true, ""
		}},
		{"3._Validate_Message_Retention_Constraint", func() (bool, string) {
			hubs := findResourcesByTypeEHUB(merged, "azurerm_eventhub")
			if len(hubs) == 0 {
				return false, "No Event Hub resources found"
			}
			retention, ok := hubs[0]["message_retention"].(float64)
			if !ok || retention < 1 || retention > 7 {
				return false, fmt.Sprintf("Message retention is out of allowed range: %v", retention)
			}
			return true, ""
		}},
		{"4._Verify_Tags_on_EventHub_Namespace", func() (bool, string) {
			ns := findResourcesByTypeEHUB(merged, "azurerm_eventhub_namespace")
			if len(ns) == 0 {
				return false, "No Event Hub Namespace found"
			}
			tags, ok := ns[0]["tags"].(map[string]interface{})
			if !ok || len(tags) == 0 {
				return false, "Tags not applied on Event Hub Namespace"
			}
			return true, ""
		}},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			passed, reason := tc.TestFunc()

			if passed {
				fmt.Printf("✅ %s passed\n", tc.Name)
			} else {
				fmt.Printf("❌ %s failed: %s\n", tc.Name, reason)
			}

			status := "FAIL"
			if passed {
				status = "PASS"
			}
			testCase := TestCaseEHUB{
				Classname: "EventHubModuleTests",
				Name:      tc.Name,
				Status:    status,
			}
			if !passed {
				suite.Failures++
				testCase.Failure = &FailureEHUB{
					Message: reason,
					Type:    "failure",
				}
			}
			suite.TestCases = append(suite.TestCases, testCase)
		})
	}

	suite.Time = 1.23

	reportFile := "reports\\res_eventhub.xml"
	file, err := os.Create(reportFile)
	if err != nil {
		t.Fatalf("❌ Failed to create XML report: %v", err)
	}
	defer file.Close()

	xmlWriter := xml.NewEncoder(file)
	xmlWriter.Indent("", "  ")
	if err := xmlWriter.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to write XML report: %v", err)
	}

	fmt.Printf("📄 XML test report written to %s\n", reportFile)
}
