package test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

type TestCasePublicIPNew struct {
	XMLName   xml.Name            `xml:"testcase"`
	Classname string              `xml:"classname,attr"`
	Name      string              `xml:"name,attr"`
	Failure   *FailurePublicIPNew `xml:"failure,omitempty"`
	Status    string              `xml:"status"`
}

type FailurePublicIPNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuitePublicIPNew struct {
	XMLName   xml.Name              `xml:"testsuite"`
	Tests     int                   `xml:"tests,attr"`
	Failures  int                   `xml:"failures,attr"`
	Errors    int                   `xml:"errors,attr"`
	Time      float64               `xml:"time,attr"`
	TestCases []TestCasePublicIPNew `xml:"testcase"`
}

func loadRemoteTFStatePublicIPNew(t *testing.T, url string) map[string]interface{} {
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

func findResourcesByTypePublicIPNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
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

func TestPublicIPValidationNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStatePublicIPNew(t, remoteStateURL)

	var suite TestSuitePublicIPNew
	suite.Tests = 3

	publicIPs := findResourcesByTypePublicIPNew(tfState, "azurerm_public_ip")

	tests := []struct {
		Name     string
		TestFunc func() (bool, string)
	}{
		{"1._Validate_Public_IP_Exists", func() (bool, string) {
			if len(publicIPs) == 0 {
				return false, "No Public IP resource found"
			}
			return true, ""
		}},
		{"2._Validate_Public_IP_Tags", func() (bool, string) {
			if len(publicIPs) == 0 {
				return false, "No Public IP resource found"
			}
			tags, ok := publicIPs[0]["tags"].(map[string]interface{})
			if !ok || len(tags) == 0 {
				return false, "Expected tags on Public IP"
			}
			return true, ""
		}},
		{"3._Validate_Public_IP_Allocation_Method", func() (bool, string) {
			if len(publicIPs) == 0 {
				return false, "No Public IP resource found"
			}
			if method, ok := publicIPs[0]["allocation_method"].(string); !ok || method != "Static" {
				return false, fmt.Sprintf("Expected allocation_method 'Static', got: %v", publicIPs[0]["allocation_method"])
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
			suite.TestCases = append(suite.TestCases, TestCasePublicIPNew{
				Classname: "PublicIPTests",
				Name:      tc.Name,
				Status:    status,
				Failure: func() *FailurePublicIPNew {
					if !passed {
						return &FailurePublicIPNew{Message: reason, Type: "failure"}
					}
					return nil
				}(),
			})
		})
	}

	suite.Time = 1.0
	reportFile := "reports/public_ip_test_report.xml"
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

	fmt.Println("📄 public_ip_test_report.xml written successfully")
}
