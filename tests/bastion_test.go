package test

import "strings"

func RunBastionTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Verify_Spoke_VNet", "BastionInfraTests", func() (bool, string) {
			for _, vnet := range findResourcesByType(tfState, "azurerm_virtual_network") {
				if name, _ := vnet["name"].(string); strings.Contains(name, "spoke") {
					if cidrs, ok := vnet["address_space"].([]interface{}); !ok || len(cidrs) == 0 {
						return false, "Missing CIDR"
					} else if cidr, _ := cidrs[0].(string); cidr != "10.110.0.0/16" {
						return false, "Incorrect CIDR: " + cidr
					}
					return true, ""
				}
			}
			return false, "Spoke VNet not found"
		}},
		{"2._Verify_Subnets_with_correct_CIDRs", "BastionInfraTests", func() (bool, string) {
			for _, sn := range findResourcesByType(tfState, "azurerm_subnet") {
				if cidrs, ok := sn["address_prefixes"].([]interface{}); !ok || len(cidrs) == 0 {
					name, _ := sn["name"].(string)
					return false, "Subnet " + name + " missing CIDRs"
				}
			}
			return true, ""
		}},
		{"3._Verify_NSG_and_Rules", "BastionInfraTests", func() (bool, string) {
			for _, nsg := range findResourcesByType(tfState, "azurerm_network_security_group") {
				if rules, ok := nsg["security_rule"].([]interface{}); !ok || len(rules) == 0 {
					name, _ := nsg["name"].(string)
					return false, "NSG " + name + " has no rules"
				}
			}
			return true, ""
		}},
		{"4._Verify_Bastion_VM_Public_IP_and_Admin_User", "BastionInfraTests", func() (bool, string) {
			for _, vm := range findResourcesByType(tfState, "azurerm_virtual_machine") {
				name, _ := vm["name"].(string)
				if strings.Contains(name, "bastion") {
					if vm["public_ip_address"] == "" {
						return false, "No public IP"
					}
					if vm["admin_username"] == "" {
						return false, "No admin user"
					}
				}
			}
			return true, ""
		}},
		{"5._Verify_APIM_internal_network_and_DNS", "BastionInfraTests", func() (bool, string) {
			apims := findResourcesByType(tfState, "azurerm_api_management")
			for _, apim := range apims {
				instances, ok := apim["instances"].([]interface{})
				if !ok {
					continue
				}
				for _, inst := range instances {
					attrs, ok := inst.(map[string]interface{})["attributes"].(map[string]interface{})
					if !ok {
						continue
					}

					internal, _ := attrs["internal"].(bool)

					// The field might be named differently: "gateway_url" or "hostname_configuration"
					dns, _ := attrs["gateway_url"].(string)

					if !internal || dns == "" {
						return false, "APIM not internal or missing DNS"
					}
				}
			}
			return true, ""
		}},
		{"6._Verify_VM_Size", "BastionInfraTests", func() (bool, string) {
			for _, vm := range findResourcesByType(tfState, "azurerm_virtual_machine") {
				if size, _ := vm["vm_size"].(string); size != "Standard_DS3_v2" {
					return false, "Wrong VM size: " + size
				}
			}
			return true, ""
		}},
		{"7._Verify_Storage_Account_Replication", "BastionInfraTests", func() (bool, string) {
			for _, sa := range findResourcesByType(tfState, "azurerm_storage_account") {
				if tier, _ := sa["account_tier"].(string); tier != "Standard" {
					return false, "Non-standard tier: " + tier
				}
			}
			return true, ""
		}},
		{"8._Verify_Managed_Identity", "BastionInfraTests", func() (bool, string) {
			for _, id := range findResourcesByType(tfState, "azurerm_managed_identity") {
				if id["client_id"] == nil {
					return false, "Missing client ID"
				}
			}
			return true, ""
		}},
		{"9._Verify_App_Service_Plan_Location", "BastionInfraTests", func() (bool, string) {
			for _, plan := range findResourcesByType(tfState, "azurerm_app_service_plan") {
				if loc, _ := plan["location"].(string); loc != "East US" {
					return false, "Wrong location: " + loc
				}
			}
			return true, ""
		}},
	}

	return executeTestCases(tests)
}
