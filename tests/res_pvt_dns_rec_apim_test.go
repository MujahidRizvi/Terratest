package test

import (
	"fmt"
	"strings"
)

func RunDNSTests(tfState map[string]interface{}) []TestCase {

	dnsRecords := findResourcesByType(tfState, "azurerm_private_dns_a_record")

	tests := []GenericTest{
		{
			"1._Validate_Main_APIM_A_Record_Creation",
			"DNSRecordTests",
			func() (bool, string) {
				for _, rec := range dnsRecords {
					if name, ok := rec["name"].(string); ok && strings.Contains(name, "apim") {
						return true, ""
					}
				}
				return false, "Main APIM A record not found"
			},
		},
		{
			"2._Validate_Extra_Prefix_DNS_Records_for_each",
			"DNSRecordTests",
			func() (bool, string) {
				foundKeys := map[string]bool{}
				detectedNames := []string{}

				for _, rec := range dnsRecords {
					name, hasName := rec["name"].(string)
					indexKey, hasIndexKey := rec["index_key"].(string)

					if hasName {
						detectedNames = append(detectedNames, name)
					}

					if hasIndexKey && indexKey != "" {
						foundKeys[indexKey] = true
					} else if hasName {
						// If index_key missing, try to extract prefix dynamically from name.
						// Assuming the prefix is last dot-separated part, or you can customize parsing here.
						parts := strings.Split(name, ".")
						prefix := parts[len(parts)-1]
						foundKeys[prefix] = true
					}
				}

				if len(foundKeys) == 0 {
					return false, fmt.Sprintf("No A records with extra prefixes found (for_each logic likely broken). Found names: %v", detectedNames)
				}

				return true, fmt.Sprintf("Found A records for prefixes/index_keys: %v", foundKeys)
			},
		},
		{
			"3._Validate_Empty_Prefix_Not_Provisioned",
			"DNSRecordTests",
			func() (bool, string) {
				for _, rec := range dnsRecords {
					if name, ok := rec["name"].(string); ok && strings.TrimSpace(name) == "" {
						return false, "Empty DNS record prefix should not be created"
					}
				}
				return true, ""
			},
		},
		{
			"4._Validate_Private_IPs_Exist_In_Records",
			"DNSRecordTests",
			func() (bool, string) {
				for _, rec := range dnsRecords {
					if ips, ok := rec["records"].([]interface{}); !ok || len(ips) == 0 {
						return false, fmt.Sprintf("Record '%v' has no private IPs", rec["name"])
					}
				}
				return true, ""
			},
		},
	}

	return executeTestCases(tests)
}
