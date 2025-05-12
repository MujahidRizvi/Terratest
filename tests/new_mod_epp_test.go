package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
)

type TestCaseStateNew struct {
	XMLName   xml.Name         `xml:"testcase"`
	Classname string           `xml:"classname,attr"`
	Name      string           `xml:"name,attr"`
	Failure   *FailureStateNew `xml:"failure,omitempty"`
	Status    string           `xml:"status"`
}

type FailureStateNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteStateNew struct {
	XMLName   xml.Name           `xml:"testsuite"`
	Tests     int                `xml:"tests,attr"`
	Failures  int                `xml:"failures,attr"`
	Errors    int                `xml:"errors,attr"`
	Time      float64            `xml:"time,attr"`
	TestCases []TestCaseStateNew `xml:"testcase"`
}

func loadRemoteTFStateNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to download terraform state from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("❌ Failed to download terraform state from %s, status code: %d", url, resp.StatusCode)
	}

	var tfState map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&tfState); err != nil {
		t.Fatalf("❌ Failed to parse terraform state from %s: %v", url, err)
	}
	return tfState
}

func findResourcesByTypeStateNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestStateFromURLBasedNew(t *testing.T) {
	suite := TestSuiteStateNew{Tests: 5}
	stateURL := "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStateNew(t, stateURL)

	tests := []struct {
		Name     string
		TestFunc func(map[string]interface{}) (bool, string)
	}{
		{"1._Verify_Resource_Group_Existence_and_Properties_New", func(state map[string]interface{}) (bool, string) {
			rgs := findResourcesByTypeStateNew(state, "azurerm_resource_group")
			found := false
			for _, rg := range rgs {
				if strings.Contains(rg["name"].(string), "epp-rg") {
					found = true
					break
				}
			}
			if !found {
				return false, "EPP Resource Group 'epp-rg' not found"
			}
			return true, ""
		}},

		{"2._Verify_Event_Hub_Namespace_and_Event_Hub_New", func(state map[string]interface{}) (bool, string) {
			nss := findResourcesByTypeStateNew(state, "azurerm_eventhub_namespace")
			ehs := findResourcesByTypeStateNew(state, "azurerm_eventhub")
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
			foundEH := false
			for _, eh := range ehs {
				if strings.Contains(eh["name"].(string), "epphub-eh") {
					foundEH = true
					break
				}
			}
			if !foundEH {
				return false, "Event Hub 'epphub-eh' not found"
			}
			return true, ""
		}},

		{"3._Verify_Private_Endpoint_Existence_and_Configuration_New", func(state map[string]interface{}) (bool, string) {
			peps := findResourcesByTypeStateNew(state, "azurerm_private_endpoint")
			found := false
			for _, pep := range peps {
				if strings.Contains(pep["name"].(string), "epphub-pep") {
					found = true
					break
				}
			}
			if !found {
				return false, "Private Endpoint 'epphub-pep' not found"
			}
			return true, ""
		}},

		// 4. Verify Tag Consistency Across All Resources
		{
			Name: "4._Verify_Tag_Consistency_Across_All_Resources_New",
			TestFunc: func(state map[string]interface{}) (bool, string) {
				expectedTags := map[string]string{
					"project": "API Ecosystem",
				}

				resources, ok := state["resources"].([]interface{})
				if !ok {
					return false, "No resources found in the state"
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
			Name: "5._Verify_Event_Hub_Namespace_Network_Access_New",
			TestFunc: func(state map[string]interface{}) (bool, string) {
				namespaces := findResourcesByTypeStateNew(state, "azurerm_eventhub_namespace")
				if len(namespaces) == 0 {
					return false, "No Event Hub Namespace found"
				}

				for _, ns := range namespaces {
					name, _ := ns["name"].(string)

					instances, ok := ns["instances"].([]interface{})
					if !ok || len(instances) == 0 {
						continue
					}
					instance := instances[0].(map[string]interface{})
					attrs, ok := instance["attributes"].(map[string]interface{})
					if !ok {
						continue
					}

					valRaw, exists := attrs["public_network_access_enabled"]
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
			result, msg := test.TestFunc(tfState)
			status := "PASS"
			if !result {
				status = "FAIL"
				suite.Failures++
				suite.TestCases = append(suite.TestCases, TestCaseStateNew{
					Classname: "Terraform Test",
					Name:      test.Name,
					Failure:   &FailureStateNew{Message: msg, Type: "error"},
					Status:    status,
				})
			} else {
				suite.TestCases = append(suite.TestCases, TestCaseStateNew{
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
	err = os.WriteFile("reports\\newwwww_epp_report.xml", xmlOut, 0644)
	if err != nil {
		t.Fatalf("❌ Failed to write XML report: %v", err)
	}
	fmt.Println("✅ EPP report saved to epp_report.xml")
}
