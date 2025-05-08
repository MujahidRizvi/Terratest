package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"
)

type TestCaseEPP struct {
	XMLName   xml.Name    `xml:"testcase"`
	Classname string      `xml:"classname,attr"`
	Name      string      `xml:"name,attr"`
	Failure   *FailureEPP `xml:"failure,omitempty"`
	Status    string      `xml:"status"`
}

type FailureEPP struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteEPP struct {
	XMLName   xml.Name      `xml:"testsuite"`
	Tests     int           `xml:"tests,attr"`
	Failures  int           `xml:"failures,attr"`
	Errors    int           `xml:"errors,attr"`
	Time      float64       `xml:"time,attr"`
	TestCases []TestCaseEPP `xml:"testcase"`
}

func loadTFStateEPP(t *testing.T, path string) map[string]interface{} {
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

func mergeStatesEPP(tfState1, tfState2 map[string]interface{}) map[string]interface{} {
	merged := make(map[string]interface{})
	for k, v := range tfState1 {
		merged[k] = v
	}
	for k, v := range tfState2 {
		merged[k] = v
	}
	return merged
}

func findResourcesByTypeEPP(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var foundResources []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return foundResources
	}
	for _, res := range resources {
		resourceMap, ok := res.(map[string]interface{})
		if !ok || resourceMap["type"] != resourceType {
			continue
		}
		instances, ok := resourceMap["instances"].([]interface{})
		if !ok {
			continue
		}
		for _, inst := range instances {
			instMap, ok := inst.(map[string]interface{})
			if !ok {
				continue
			}
			attributes, ok := instMap["attributes"].(map[string]interface{})
			if ok {
				foundResources = append(foundResources, attributes)
			}
		}
	}
	return foundResources
}

func TestEppStateBased(t *testing.T) {
	suite := TestSuiteEPP{Tests: 5}
	tfState1 := loadTFStateEPP(t, "../terraform.tfstate")
	tfState2 := loadTFStateEPP(t, "../terra.tfstate")
	merged := mergeStatesEPP(tfState1, tfState2)

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Verify_Resource_Group_Existence_and_Properties", func() (bool, string) {
			rgs := findResourcesByTypeEPP(merged, "azurerm_resource_group")
			for _, rg := range rgs {
				if strings.Contains(rg["name"].(string), "epp-rg") {
					return true, ""
				}
			}
			return false, "EPP Resource Group 'epp-rg' not found"
		}},

		{"2._Verify_Event_Hub_Namespace_and_Event_Hub", func() (bool, string) {
			nss := findResourcesByTypeEPP(merged, "azurerm_eventhub_namespace")
			ehs := findResourcesByTypeEPP(merged, "azurerm_eventhub")
			foundNS := false
			for _, ns := range nss {
				if strings.Contains(ns["name"].(string), "epphubspace-ns") {
					foundNS = true
					break
				}
			}
			if !foundNS {
				return false, "Event Hub Namespace 'epphubspace-ns' not found"
			}
			for _, eh := range ehs {
				if strings.Contains(eh["name"].(string), "epphub-eh") {
					return true, ""
				}
			}
			return false, "Event Hub 'epphub-eh' not found"
		}},

		{"3._Verify_Private_Endpoint_Existence_and_Configuration", func() (bool, string) {
			peps := findResourcesByTypeEPP(merged, "azurerm_private_endpoint")
			for _, pep := range peps {
				if strings.Contains(pep["name"].(string), "epphub-pep") {
					return true, ""
				}
			}
			return false, "Private Endpoint 'epphub-pep' not found"
		}},

		// 4. Verify Tag Consistency Across All Resources
		{
			Name: "4._Verify_Tag_Consistency_Across_All_Resources",
			TestFunc: func() (bool, string) {
				expectedTags := map[string]string{
					"project": "API Ecosystem",
				}

				resources, ok := merged["resources"].([]interface{})
				if !ok {
					return false, "No resources found in the merged state"
				}

				for _, res := range resources {
					resMap, ok := res.(map[string]interface{})
					if !ok {
						continue
					}

					instances, ok := resMap["instances"].([]interface{})
					if !ok {
						continue
					}

					for _, inst := range instances {
						instMap, ok := inst.(map[string]interface{})
						if !ok {
							continue
						}

						attrs, ok := instMap["attributes"].(map[string]interface{})
						if !ok {
							continue
						}

						tags, ok := attrs["tags"].(map[string]interface{})
						if !ok {
							return false, fmt.Sprintf("Resource '%v' missing tags", attrs["name"])
						}

						for key, expectedVal := range expectedTags {
							actualValRaw, exists := tags[key]
							if !exists {
								return false, fmt.Sprintf("Resource '%v' missing expected tag '%v'", attrs["name"], key)
							}
							actualVal, ok := actualValRaw.(string)
							if !ok || actualVal != expectedVal {
								return false, fmt.Sprintf("Resource '%v' tag '%v' mismatch: expected '%v', got '%v'", attrs["name"], key, expectedVal, actualVal)
							}
						}
					}
				}
				return true, ""
			},
		},

		// 5. Verify Event Hub Namespace Network Access
		{
			Name: "5._Verify_Event_Hub_Namespace_Network_Access",
			TestFunc: func() (bool, string) {
				namespaces := findResourcesByTypeEPP(merged, "azurerm_eventhub_namespace")
				if len(namespaces) == 0 {
					return false, "No Event Hub Namespace found"
				}

				for _, ns := range namespaces {
					name, _ := ns["name"].(string)

					// Safely extract and validate the 'public_network_access_enabled' attribute
					valRaw, exists := ns["public_network_access_enabled"]
					if !exists {
						return false, fmt.Sprintf("Event Hub Namespace '%v' missing 'public_network_access_enabled' attribute", name)
					}

					enabled, ok := valRaw.(bool)
					if !ok {
						return false, fmt.Sprintf("Event Hub Namespace '%v' has non-boolean 'public_network_access_enabled': %v", name, valRaw)
					}

					if enabled {
						return false, fmt.Sprintf("Event Hub Namespace '%v' has public network access ENABLED (should be disabled)", name)
					}
				}
				return true, ""
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			result, msg := test.TestFunc()
			status := "PASS"
			if !result {
				status = "FAIL"
				suite.Failures++
				suite.TestCases = append(suite.TestCases, TestCaseEPP{
					Classname: "Terraform Test",
					Name:      test.Name,
					Failure:   &FailureEPP{Message: msg, Type: "error"},
					Status:    status,
				})
			} else {
				suite.TestCases = append(suite.TestCases, TestCaseEPP{
					Classname: "Terraform Test",
					Name:      test.Name,
					Status:    status,
				})
			}
		})
	}

	xmlOut, err := xml.MarshalIndent(suite, "", "  ")
	if err != nil {
		t.Fatalf("❌ Failed to generate XML report: %v", err)
	}
	err = os.WriteFile("reports\\epp_report.xml", xmlOut, 0644)
	if err != nil {
		t.Fatalf("❌ Failed to write XML report: %v", err)
	}
	fmt.Println("✅ EPP report saved to epp_report.xml")
}
