package test

import (
	"fmt"
)

func RunEventHubTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Verify_EventHub_Namespace_Creation", "EventHubTests", func() (bool, string) {
			ns := findResourcesByType(tfState, "azurerm_eventhub_namespace")
			if len(ns) == 0 {
				return false, "Expected at least one Event Hub Namespace"
			}
			return true, ""
		}},
		{"2._Verify_EventHub_Creation", "EventHubTests", func() (bool, string) {
			hubs := findResourcesByType(tfState, "azurerm_eventhub")
			if len(hubs) == 0 {
				return false, "Expected at least one Event Hub"
			}
			return true, ""
		}},
		{"3._Validate_Message_Retention_Constraint", "EventHubTests", func() (bool, string) {
			hubs := findResourcesByType(tfState, "azurerm_eventhub")
			if len(hubs) == 0 {
				return false, "No Event Hub resources found"
			}
			retention, ok := hubs[0]["message_retention"].(float64)
			if !ok || retention < 1 || retention > 7 {
				return false, fmt.Sprintf("Message retention is out of range: %v", retention)
			}
			return true, ""
		}},
		{"4._Verify_Tags_on_EventHub_Namespace", "EventHubTests", func() (bool, string) {
			ns := findResourcesByType(tfState, "azurerm_eventhub_namespace")
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

	return executeTestCases(tests)
}
