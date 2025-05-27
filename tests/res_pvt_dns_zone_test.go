package test

import (
	"strings"
)

func RunPrivateDNSZoneTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{
			"1._Verify_Private_DNS_Zone_Created_with_Proper_Name",
			"PrivateDNSZoneTests",
			func() (bool, string) {
				zones := findResourcesByType(tfState, "azurerm_private_dns_zone")
				if len(zones) == 0 {
					return false, "No private DNS zones found"
				}
				for _, z := range zones {
					if name, _ := z["name"].(string); strings.HasSuffix(name, ".azure-api.net") {
						return true, ""
					}
				}
				return false, "No private DNS zone ends with .azure-api.net"
			},
		},
		{
			"2._Verify_DNS_Zone_Links_and_Tags",
			"PrivateDNSZoneTests",
			func() (bool, string) {
				links := findResourcesByType(tfState, "azurerm_private_dns_zone_virtual_network_link")
				if len(links) == 0 {
					return false, "No virtual network link found for DNS zone"
				}
				for _, l := range links {
					tags, ok := l["tags"].(map[string]interface{})
					if !ok || tags["Environment"] == nil {
						return false, "Missing expected Environment tag"
					}
				}
				return true, ""
			},
		},
	}

	return executeTestCases(tests)
}
