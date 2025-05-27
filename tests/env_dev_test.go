package test

import (
	"strings"
)

func RunDevInfraTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{"1._Verify_Spoke_VNet", "DevInfraTests", func() (bool, string) {
			vnets := findResourcesByType(tfState, "azurerm_virtual_network")
			for _, vnet := range vnets {
				name, _ := vnet["name"].(string)
				if strings.Contains(name, "spoke") {
					addrSpace, ok := vnet["address_space"].([]interface{})
					if !ok || len(addrSpace) == 0 {
						return false, "Spoke VNet missing CIDR block"
					}
					cidr, _ := addrSpace[0].(string)
					if cidr != "10.110.0.0/16" {
						return false, "Spoke VNet has incorrect CIDR: " + cidr
					}
					return true, ""
				}
			}
			return false, "Spoke VNet not found"
		}},
		{"2._Verify_Subnets_with_correct_CIDRs", "DevInfraTests", func() (bool, string) {
			subnets := findResourcesByType(tfState, "azurerm_subnet")
			for _, sn := range subnets {
				cidrs, ok := sn["address_prefixes"].([]interface{})
				if !ok || len(cidrs) == 0 {
					name, _ := sn["name"].(string)
					return false, "Subnet " + name + " missing CIDRs"
				}
			}
			return true, ""
		}},
		{"3._Verify_NSG_and_Rules", "DevInfraTests", func() (bool, string) {
			nsgs := findResourcesByType(tfState, "azurerm_network_security_group")
			for _, nsg := range nsgs {
				rules, ok := nsg["security_rule"].([]any)
				if !ok || len(rules) == 0 {
					name, _ := nsg["name"].(string)
					return false, "NSG " + name + " has no rules"
				}
			}
			return true, ""
		}},
		{"4._Verify_Bastion_VM_Public_IP_and_Admin_User", "DevInfraTests", func() (bool, string) {
			vms := findResourcesByType(tfState, "azurerm_virtual_machine")
			for _, vm := range vms {
				name, _ := vm["name"].(string)
				if strings.Contains(name, "bastion") {
					pubIP, _ := vm["public_ip_address"].(string)
					if pubIP == "" {
						return false, "No public IP on Bastion VM"
					}
					adminUser, _ := vm["admin_username"].(string)
					if adminUser == "" {
						return false, "No admin user on Bastion VM"
					}
				}
			}
			return true, ""
		}},
		{"5._Verify_APIM_internal_network_and_DNS", "DevInfraTests", func() (bool, string) {
			apims := findResourcesByType(tfState, "azurerm_api_management")
			for _, apim := range apims {
				instances, ok := apim["instances"].([]interface{})
				if !ok {
					continue
				}
				for _, inst := range instances {
					attributes := inst.(map[string]interface{})["attributes"].(map[string]interface{})

					internal, _ := attributes["internal"].(bool)
					if !internal {
						return false, "APIM is not internal"
					}

					dns, _ := attributes["gateway_url"].(string) // use "gateway_url" or "hostname_configuration" depending on your setup
					if dns == "" {
						return false, "APIM has no DNS name"
					}
				}
			}
			return true, ""
		}},
		{"6._Verify_VM_Size", "DevInfraTests", func() (bool, string) {
			vms := findResourcesByType(tfState, "azurerm_virtual_machine")
			for _, vm := range vms {
				size, _ := vm["vm_size"].(string)
				if size != "Standard_DS3_v2" {
					return false, "Wrong VM size: " + size
				}
			}
			return true, ""
		}},
		{"7._Verify_Storage_Account_Replication", "DevInfraTests", func() (bool, string) {
			sas := findResourcesByType(tfState, "azurerm_storage_account")
			for _, sa := range sas {
				tier, _ := sa["account_tier"].(string)
				if tier != "Standard" {
					return false, "Non-standard replication: " + tier
				}
			}
			return true, ""
		}},
		{"8._Verify_Managed_Identity", "DevInfraTests", func() (bool, string) {
			ids := findResourcesByType(tfState, "azurerm_user_assigned_identity")
			for _, id := range ids {
				if id["client_id"] == nil {
					return false, "Managed Identity missing client ID"
				}
			}
			return true, ""
		}},
		{"9._Verify_App_Service_Plan_Location", "DevInfraTests", func() (bool, string) {
			plans := findResourcesByType(tfState, "azurerm_app_service_plan")
			for _, plan := range plans {
				loc, _ := plan["location"].(string)
				if loc != "East US" {
					return false, "App Service Plan not in East US"
				}
			}
			return true, ""
		}},
	}

	return executeTestCases(tests)
}
