package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type TestCaseDNSNew struct {
	XMLName   xml.Name       `xml:"testcase"`
	Classname string         `xml:"classname,attr"`
	Name      string         `xml:"name,attr"`
	Failure   *FailureDNSNew `xml:"failure,omitempty"`
	Status    string         `xml:"status"`
}

type FailureDNSNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteDNSNew struct {
	XMLName   xml.Name         `xml:"testsuite"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Errors    int              `xml:"errors,attr"`
	Time      float64          `xml:"time,attr"`
	TestCases []TestCaseDNSNew `xml:"testcase"`
}

func loadRemoteTFStateDNSNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote state: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read response body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to unmarshal JSON: %v", err)
	}
	return tfState
}

func findResourcesByTypeDNSNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var found []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return found
	}
	for _, res := range resources {
		rm, ok := res.(map[string]interface{})
		if !ok || rm["type"] != resourceType {
			continue
		}
		instances, ok := rm["instances"].([]interface{})
		if !ok {
			continue
		}
		for _, inst := range instances {
			im, ok := inst.(map[string]interface{})
			if !ok {
				continue
			}
			if attrs, ok := im["attributes"].(map[string]interface{}); ok {
				found = append(found, attrs)
			}
		}
	}
	return found
}

func TestDNSRecordValidationNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStateDNSNew(t, remoteStateURL)

	var suite TestSuiteDNSNew
	suite.Tests = 4

	dnsRecords := findResourcesByTypeDNSNew(tfState, "azurerm_private_dns_a_record")

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Validate_Main_APIM_A_Record_Creation", func() (bool, string) {
			for _, rec := range dnsRecords {
				if name, ok := rec["name"].(string); ok && strings.Contains(name, "apim") {
					return true, ""
				}
			}
			return false, "Main APIM A record not found"
		}},
		{"2._Validate_Extra_Prefix_DNS_Records_for_each", func() (bool, string) {
			count := 0
			for _, rec := range dnsRecords {
				if name, ok := rec["name"].(string); ok && strings.Contains(name, "prefix") {
					count++
				}
			}
			if count == 0 {
				return false, "No A records with extra prefixes found (for_each logic likely broken)"
			}
			return true, ""
		}},
		{"3._Validate_Empty_Prefix_Not_Provisioned", func() (bool, string) {
			for _, rec := range dnsRecords {
				if name, ok := rec["name"].(string); ok && strings.TrimSpace(name) == "" {
					return false, "Empty DNS record prefix should not be created"
				}
			}
			return true, ""
		}},
		{"4._Validate_Private_IPs_Exist_In_Records", func() (bool, string) {
			for _, rec := range dnsRecords {
				if ips, ok := rec["records"].([]interface{}); !ok || len(ips) == 0 {
					return false, fmt.Sprintf("Record '%v' has no private IPs", rec["name"])
				}
			}
			return true, ""
		}},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			passed, reason := tc.TestFunc()
			status := "PASS"
			if !passed {
				status = "FAIL"
				suite.Failures++
			}
			suite.TestCases = append(suite.TestCases, TestCaseDNSNew{
				Classname: "DNSRecordTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailureDNSNew {
					if !passed {
						return &FailureDNSNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 1.0
	reportFile := "reports/new_res_pvt_dns_rec_apim_test_report.xml"
	file, err := os.Create(reportFile)
	if err != nil {
		t.Fatalf("❌ Unable to create XML report: %v", err)
	}
	defer file.Close()
	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML report: %v", err)
	}

	fmt.Println("📄 dns_records_test_report.xml written successfully")
}
