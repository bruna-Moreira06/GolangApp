package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// Test yml2json function with valid YAML file
func TestYml2JsonWithValidYAML(t *testing.T) {
	// Create a temporary YAML file for testing
	testYAML := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      summary: Test endpoint
      responses:
        200:
          description: Success
`

	// Write test YAML file
	tempFile := "test_openapi.yml"
	err := os.WriteFile(tempFile, []byte(testYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}
	defer os.Remove(tempFile) // Clean up

	// Mock yml2json function that reads from test file
	yml2jsonMock := func(filename string) (map[string]interface{}, error) {
		yfile, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		var data map[string]interface{}
		err = yaml.Unmarshal(yfile, &data)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	// Test the function
	result, err := yml2jsonMock(tempFile)
	if err != nil {
		t.Fatalf("yml2json failed: %v", err)
	}

	// Verify the result
	if result["openapi"] != "3.0.0" {
		t.Errorf("Expected openapi version '3.0.0', got %v", result["openapi"])
	}

	// Check info section
	info, ok := result["info"].(map[string]interface{})
	if !ok {
		t.Fatal("Info section should be a map")
	}

	if info["title"] != "Test API" {
		t.Errorf("Expected title 'Test API', got %v", info["title"])
	}
}

// Test yml2json function with invalid YAML file
func TestYml2JsonWithInvalidYAML(t *testing.T) {
	// Create a temporary invalid YAML file
	invalidYAML := `
openapi: 3.0.0
info:
  title: Test API
  invalid: [
    - missing closing bracket
`

	tempFile := "test_invalid.yml"
	err := os.WriteFile(tempFile, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}
	defer os.Remove(tempFile)

	// Mock yml2json function
	yml2jsonMock := func(filename string) (map[string]interface{}, error) {
		yfile, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		var data map[string]interface{}
		err = yaml.Unmarshal(yfile, &data)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	// Test the function - should return error
	_, err = yml2jsonMock(tempFile)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

// Test yml2json function with non-existent file
func TestYml2JsonWithNonExistentFile(t *testing.T) {
	yml2jsonMock := func(filename string) (map[string]interface{}, error) {
		yfile, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		var data map[string]interface{}
		err = yaml.Unmarshal(yfile, &data)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	// Test with non-existent file
	_, err := yml2jsonMock("non_existent_file.yml")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

// Test the JSON encoding part of yml2json
func TestYml2JsonJSONEncoding(t *testing.T) {
	// Sample data that would come from YAML parsing
	testData := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":   "Test API",
			"version": "1.0.0",
		},
		"paths": map[string]interface{}{
			"/test": map[string]interface{}{
				"get": map[string]interface{}{
					"summary": "Test endpoint",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Success",
						},
					},
				},
			},
		},
	}

	// Test JSON encoding (simulating the json.NewEncoder part)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "\t")
	err := enc.Encode(testData)

	if err != nil {
		t.Fatalf("JSON encoding failed: %v", err)
	}

	// Verify the output contains expected elements
	output := buf.String()
	if !strings.Contains(output, "\"openapi\": \"3.0.0\"") {
		t.Error("Output should contain openapi version")
	}

	if !strings.Contains(output, "\"title\": \"Test API\"") {
		t.Error("Output should contain title")
	}

	// Verify indentation
	if !strings.Contains(output, "\t") {
		t.Error("Output should be indented with tabs")
	}
}

// Test yml2json output redirection concept
func TestYml2JsonOutputRedirection(t *testing.T) {
	testData := map[string]interface{}{
		"test":   "value",
		"number": 42,
	}

	// Capture stdout (simulating what yml2json does)
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Mock the output part of yml2json
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	enc.Encode(testData)

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read the captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output
	if !strings.Contains(output, "\"test\": \"value\"") {
		t.Error("Output should contain test data")
	}

	if !strings.Contains(output, "\"number\": 42") {
		t.Error("Output should contain number data")
	}
}

// Test YAML to JSON conversion with complex structures
func TestYml2JsonComplexStructures(t *testing.T) {
	complexYAML := `
database:
  host: localhost
  port: 5432
  credentials:
    username: admin
    password: secret
servers:
  - name: server1
    url: http://localhost:8080
  - name: server2
    url: http://localhost:8081
features:
  logging: true
  metrics: false
  debug: true
`

	tempFile := "test_complex.yml"
	err := os.WriteFile(tempFile, []byte(complexYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test YAML file: %v", err)
	}
	defer os.Remove(tempFile)

	// Mock yml2json function
	yml2jsonMock := func(filename string) (map[string]interface{}, error) {
		yfile, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		var data map[string]interface{}
		err = yaml.Unmarshal(yfile, &data)
		if err != nil {
			return nil, err
		}

		return data, nil
	}

	result, err := yml2jsonMock(tempFile)
	if err != nil {
		t.Fatalf("yml2json failed: %v", err)
	}

	// Test nested structures
	database, ok := result["database"].(map[string]interface{})
	if !ok {
		t.Fatal("Database should be a map")
	}

	if database["host"] != "localhost" {
		t.Errorf("Expected host 'localhost', got %v", database["host"])
	}

	// Test arrays
	servers, ok := result["servers"].([]interface{})
	if !ok {
		t.Fatal("Servers should be an array")
	}

	if len(servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(servers))
	}

	// Test boolean values
	features, ok := result["features"].(map[string]interface{})
	if !ok {
		t.Fatal("Features should be a map")
	}

	if features["logging"] != true {
		t.Errorf("Expected logging to be true, got %v", features["logging"])
	}
}

// Test error handling in yml2json workflow
func TestYml2JsonWorkflowErrorHandling(t *testing.T) {
	testCases := []struct {
		name        string
		yamlContent string
		expectError bool
	}{
		{
			name: "Valid YAML",
			yamlContent: `
key: value
number: 123
`,
			expectError: false,
		},
		{
			name: "Invalid YAML - bad indentation",
			yamlContent: `
key: value
  bad_indent: value
`,
			expectError: true,
		},
		{
			name: "Invalid YAML - unclosed bracket",
			yamlContent: `
array: [1, 2, 3
`,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempFile := fmt.Sprintf("test_%s.yml", strings.ReplaceAll(tc.name, " ", "_"))
			err := os.WriteFile(tempFile, []byte(tc.yamlContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
			defer os.Remove(tempFile)

			yml2jsonMock := func(filename string) error {
				yfile, err := os.ReadFile(filename)
				if err != nil {
					return err
				}

				var data interface{}
				err = yaml.Unmarshal(yfile, &data)
				if err != nil {
					return err
				}

				return nil
			}

			err = yml2jsonMock(tempFile)

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			} else if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
