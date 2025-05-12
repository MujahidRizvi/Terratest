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

type TestCaseSRCNew struct {
	XMLName   xml.Name       `xml:"testcase"`
	Classname string         `xml:"classname,attr"`
	Name      string         `xml:"name,attr"`
	Failure   *FailureSRCNew `xml:"failure,omitempty"`
	Status    string         `xml:"status"`
}

type FailureSRCNew struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
}

type TestSuiteSRCNew struct {
	XMLName   xml.Name         `xml:"testsuite"`
	Tests     int              `xml:"tests,attr"`
	Failures  int              `xml:"failures,attr"`
	Errors    int              `xml:"errors,attr"`
	Time      float64          `xml:"time,attr"`
	TestCases []TestCaseSRCNew `xml:"testcase"`
}

func loadRemoteTFStateSRCNew(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch state: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read body: %v", err)
	}
	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse state: %v", err)
	}
	return tfState
}

func findResourcesByTypeSRCNew(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var found []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return found
	}
	for _, res := range resources {
		m, ok := res.(map[string]interface{})
		if !ok || m["type"] != resourceType {
			continue
		}
		instances, ok := m["instances"].([]interface{})
		if !ok {
			continue
		}
		for _, i := range instances {
			im, ok := i.(map[string]interface{})
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

func TestSysResourceGroupStateBasedNew(t *testing.T) {
	const remoteStateURL = "https://agidamainuaentfsa.blob.core.windows.net/tfstate/Dev/global.tfstate?sp=r&st=2025-05-09T06:39:00Z&se=2099-05-09T14:39:00Z&spr=https&sv=2024-11-04&sr=b&sig=uuplyULRDhrCAP9dVpECP32Ph7de5kPvu47V2OQ2dVs%3D"
	tfState := loadRemoteTFStateSRCNew(t, remoteStateURL)

	var suite TestSuiteSRCNew
	suite.Tests = 1

	t.Run("1._Verify_System_Resource_Group_Existence_and_Properties", func(t *testing.T) {
		passed := false
		reason := ""

		rgs := findResourcesByTypeSRCNew(tfState, "azurerm_resource_group")
		for _, rg := range rgs {
			name, _ := rg["name"].(string)
			if strings.Contains(name, "srcs-rg") {
				loc, _ := rg["location"].(string)
				if loc != "" {
					passed = true
					break
				} else {
					reason = "Location missing for srcs-rg"
				}
			}
		}
		if !passed && reason == "" {
			reason = "srcs-rg not found"
		}

		status := "PASS"
		if !passed {
			status = "FAIL"
			suite.Failures++
		}

		suite.TestCases = append(suite.TestCases, TestCaseSRCNew{
			Classname: "Terraform Test",
			Name:      "1._Verify_System_Resource_Group_Existence_and_Properties",
			Status:    status,
			Failure: func() *FailureSRCNew {
				if !passed {
					return &FailureSRCNew{
						Message: reason,
						Type:    "error",
					}
				}
				return nil
			}(),
		})
	})

	suite.Time = 0.95
	file, err := os.Create("reports/new_mod_src_report.xml")
	if err != nil {
		t.Fatalf("❌ Failed to create report file: %v", err)
	}
	defer file.Close()

	writer := xml.NewEncoder(file)
	writer.Indent("", "  ")
	if err := writer.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML: %v", err)
	}

	fmt.Println("📄 SRC report saved to reports/src_report.xml")
}
