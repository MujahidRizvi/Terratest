package test

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
)

// CLI flags
var (
	envFlag         = flag.String("env", "", "Environment name (e.g. dev, prod)")
	remoteStateFlag = flag.String("remoteStateURL", "", "Remote Terraform state URL")
)

// Config structure for test settings
type Config struct {
	Environment    string `json:"environment"`
	RemoteStateURL string `json:"remote_state_url"`
}

// Load test configuration from flag, env, or config file fallback
func LoadTestConfig(t *testing.T) *Config {
	flag.Parse()

	cfg := &Config{}

	// Optional: load from config.json
	data, err := os.ReadFile("config.json")
	if err == nil {
		_ = json.Unmarshal(data, cfg)
	}

	// Override from flags or env
	if *envFlag != "" {
		cfg.Environment = *envFlag
	} else if env := os.Getenv("TEST_ENV"); env != "" {
		cfg.Environment = env
	}

	if *remoteStateFlag != "" {
		cfg.RemoteStateURL = *remoteStateFlag
	} else if url := os.Getenv("TF_REMOTE_STATE_URL"); url != "" {
		cfg.RemoteStateURL = url
	}

	if cfg.RemoteStateURL == "" {
		t.Fatal("❌ Missing remote state URL. Use -remoteStateURL or set TF_REMOTE_STATE_URL.")
	}

	return cfg
}

// Download and parse remote Terraform state
func loadRemoteTFState(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("❌ Failed to fetch remote Terraform state: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("❌ Failed to read state body: %v", err)
	}

	var tfState map[string]interface{}
	if err := json.Unmarshal(body, &tfState); err != nil {
		t.Fatalf("❌ Failed to parse Terraform state: %v", err)
	}
	return tfState
}

// Return all resources of a given type from the Terraform state
func findResourcesByType(tfState map[string]interface{}, resourceType string) []map[string]interface{} {
	var results []map[string]interface{}
	resources, ok := tfState["resources"].([]interface{})
	if !ok {
		return results
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
				results = append(results, attrs)
			}
		}
	}
	return results
}

// Generic test definition
type GenericTest struct {
	Name     string
	Class    string
	Validate func() (bool, string)
}

// Convert GenericTest into standardized TestCase for reporting
func executeTestCases(tests []GenericTest) []TestCase {
	var cases []TestCase
	for _, tc := range tests {
		pass, msg := tc.Validate()
		status := "PASS"
		var fail *Failure
		if !pass {
			status = "FAIL"
			fail = &Failure{Message: msg, Type: "failure"}
		}
		cases = append(cases, TestCase{
			Classname: tc.Class,
			Name:      tc.Name,
			Status:    status,
			Failure:   fail,
		})
	}
	return cases
}

// Write combined JUnit-style XML report
func writeReport(t *testing.T, suite TestSuite, path string) {
	if err := os.MkdirAll("reports", 0755); err != nil {
		t.Fatalf("❌ Failed to create report directory: %v", err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("❌ Unable to create report file: %v", err)
	}
	defer file.Close()

	xmlWriter := xml.NewEncoder(file)
	xmlWriter.Indent("", "  ")
	if err := xmlWriter.Encode(suite); err != nil {
		t.Fatalf("❌ Failed to encode XML: %v", err)
	}

	fmt.Println("📄 Combined report written to", path)
}
