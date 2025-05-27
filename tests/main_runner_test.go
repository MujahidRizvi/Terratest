package test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testModule struct {
	Name string
	Func func(map[string]interface{}) []TestCase
}

func getAllTestModules() []testModule {
	return []testModule{
		{"MainInfra", RunMainInfraTests},
		{"DevInfra", RunDevInfraTests},
		{"Bastion", RunBastionTests},
		{"Proc", RunProcTests},
		{"Spoke", RunSpokeTests},
		{"Epp", RunEppTests},
		{"Exp", RunExpTests},
		{"Src", RunSrcTests},
		{"Sys", RunSysTests},
		{"APIM", RunAPIMTests},
		{"AppGateway", RunAppGatewayTests},
		{"EventHub", RunEventHubTests},
		{"FuncNetCore8ISO", RunFunctionAppNetCore8ISOTests},
		{"FuncApp", RunFunctionAppTests},
		{"LogAnalytics", RunLogAnalyticsTests},
		{"NSG", RunNSGTests},
		{"PrivateEndpoint", RunPrivateEndpointTests},
		{"PublicIP", RunPublicIPTests},
		{"DNS", RunDNSTests},
		{"PrivateDNS", RunPrivateDNSZoneTests},
		{"RG", RunResourceGroupTests},
		{"Subnet", RunSubnetTests},
		{"SubnetDelegation", RunSubnetWithDelegationTests},
		{"VNet", RunVNetValidationTests},
		{"WindowsVM", RunWindowsVMValidationTests},
	}
}

func TestAllModulesFromMain(t *testing.T) {
	cfg := LoadTestConfig(t) // Load config from flag/env/config.json
	tfState := loadRemoteTFState(t, cfg.RemoteStateURL)

	var suite TestSuite
	var mu sync.Mutex

	modules := getAllTestModules()

	for _, mod := range modules {
		mod := mod // capture range variable

		t.Run(mod.Name, func(t *testing.T) {
			t.Parallel() // Run subtest in parallel

			testCases := mod.Func(tfState)

			mu.Lock()
			defer mu.Unlock()

			for _, tc := range testCases {
				assert.NotEmpty(t, tc.Name, "Test case should have a name")
				assert.Contains(t, []string{"PASS", "FAIL"}, tc.Status, "Test case should have valid status")

				suite.TestCases = append(suite.TestCases, tc)
				suite.Tests++
				if tc.Status == "FAIL" {
					suite.Failures++
				}
			}
		})
	}

	t.Cleanup(func() {
		writeReport(t, suite, "reports/overall_modules_parallel_report.xml")
	})
}
