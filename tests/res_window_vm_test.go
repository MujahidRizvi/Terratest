package test

func RunWindowsVMValidationTests(tfState map[string]interface{}) []TestCase {
	tests := []GenericTest{
		{
			"1._Verify_Windows_VM_Exists_with_Correct_Configuration",
			"WindowsVMTests",
			func() (bool, string) {
				vms := findResourcesByType(tfState, "azurerm_windows_virtual_machine")
				if len(vms) == 0 {
					return false, "Windows VM not found"
				}
				for _, vm := range vms {
					if vm["name"] != nil && vm["network_interface_ids"] != nil && vm["os_disk"] != nil {
						return true, ""
					}
				}
				return false, "Windows VM exists but missing required configuration"
			},
		},
		{
			"2._Verify_Network_Interface_Attached_to_VM",
			"WindowsVMTests",
			func() (bool, string) {
				nics := findResourcesByType(tfState, "azurerm_network_interface")
				vms := findResourcesByType(tfState, "azurerm_windows_virtual_machine")
				if len(nics) == 0 || len(vms) == 0 {
					return false, "NIC or VM not found"
				}
				nicID := nics[0]["id"]
				for _, vm := range vms {
					if ids, ok := vm["network_interface_ids"].([]interface{}); ok {
						for _, id := range ids {
							if id == nicID {
								return true, ""
							}
						}
					}
				}
				return false, "NIC not attached to VM"
			},
		},
		{
			"3._Verify_Public_IP_Creation_and_Association",
			"WindowsVMTests",
			func() (bool, string) {
				publicIPs := findResourcesByType(tfState, "azurerm_public_ip")
				nics := findResourcesByType(tfState, "azurerm_network_interface")
				if len(publicIPs) == 0 {
					return false, "Public IP not created"
				}
				if len(nics) == 0 {
					return false, "NIC not found for validation"
				}
				for _, nic := range nics {
					if ipConfigs, ok := nic["ip_configuration"].([]interface{}); ok {
						for _, cfg := range ipConfigs {
							cfgMap, ok := cfg.(map[string]interface{})
							if ok && cfgMap["public_ip_address_id"] != nil {
								return true, ""
							}
						}
					}
				}
				return false, "Public IP not associated with NIC"
			},
		},
	}

	return executeTestCases(tests)
}
