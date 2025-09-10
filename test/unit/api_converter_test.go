package unit

import (
	"encoding/json"
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYml2JsonFunctionality(t *testing.T) {
	// Test the YAML to JSON conversion logic
	// Since the original function reads from openapi.yml and writes to stdout,
	// we'll test the core conversion logic

	// Create test YAML data
	testYAML := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      summary: Test endpoint
`

	// Test YAML unmarshaling
	var data interface{}
	err := yaml.Unmarshal([]byte(testYAML), &data)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Test JSON marshaling
	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Verify the JSON contains expected content
	jsonStr := string(jsonData)
	expectedFields := []string{"openapi", "info", "title", "Test API", "paths"}

	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("JSON output should contain '%s', got: %s", field, jsonStr)
		}
	}
}

func TestYml2JsonWithValidFile(t *testing.T) {
	// Create a temporary YAML file for testing
	tempYAML := `
openapi: 3.0.0
info:
  title: Cats API
  version: 1.0.0
`

	// Create temp file
	tmpFile, err := os.CreateTemp("", "test_openapi*.yml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test data
	if _, err := tmpFile.Write([]byte(tempYAML)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	// Test file reading and conversion
	yfile, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	var data interface{}
	err = yaml.Unmarshal(yfile, &data)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML from file: %v", err)
	}

	// Verify data structure
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		t.Error("YAML should unmarshal to a map")
		return
	}

	if dataMap["openapi"] != "3.0.0" {
		t.Errorf("Expected openapi version 3.0.0, got %v", dataMap["openapi"])
	}

	info, exists := dataMap["info"]
	if !exists {
		t.Error("YAML should contain info section")
		return
	}

	infoMap, ok := info.(map[string]interface{})
	if !ok {
		t.Error("Info should be a map")
		return
	}

	if infoMap["title"] != "Cats API" {
		t.Errorf("Expected title 'Cats API', got %v", infoMap["title"])
	}
}

func TestYml2JsonErrorHandling(t *testing.T) {
	// Test with invalid YAML
	invalidYAML := `
openapi: 3.0.0
info:
  title: Test API
  - invalid: yaml: structure
`

	var data interface{}
	err := yaml.Unmarshal([]byte(invalidYAML), &data)
	if err == nil {
		t.Error("Expected error when unmarshaling invalid YAML")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsInner(s, substr))))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
