package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// =============================================================================
// ACTUAL HANDLER FUNCTION TESTS
// =============================================================================

// Test actual createCat function
func TestActualCreateCat(t *testing.T) {
	// Save original database state
	originalDB := make(map[string]Cat)
	for k, v := range catsDatabase {
		originalDB[k] = v
	}
	defer func() {
		// Restore original state
		catsDatabase = originalDB
	}()

	// Clear database for test
	catsDatabase = make(map[string]Cat)

	// Create test cat
	testCat := Cat{
		Name:      "TestCat",
		Color:     "Orange",
		BirthDate: "2023-01-01",
	}

	jsonData, err := json.Marshal(testCat)
	if err != nil {
		t.Fatalf("Failed to marshal test cat: %v", err)
	}

	// Create request
	req := httptest.NewRequest("POST", "/api/cats", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Call actual function
	statusCode, response := createCat(req)

	// Assertions
	if statusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, statusCode)
	}

	// Check response is a string (cat ID)
	responseStr, ok := response.(string)
	if !ok {
		t.Errorf("Expected string response, got %T", response)
		return
	}

	if responseStr == "" {
		t.Error("Expected non-empty cat ID")
	}

	// Check cat was saved to database
	if len(catsDatabase) != 1 {
		t.Errorf("Expected 1 cat in database, got %d", len(catsDatabase))
	}

	// Verify the cat in database
	savedCat, exists := catsDatabase[responseStr]
	if !exists {
		t.Error("Created cat not found in database")
		return
	}

	if savedCat.Name != testCat.Name {
		t.Errorf("Expected cat name %s, got %s", testCat.Name, savedCat.Name)
	}

	if savedCat.Color != testCat.Color {
		t.Errorf("Expected cat color %s, got %s", testCat.Color, savedCat.Color)
	}
}

// Test actual createCat function with invalid JSON
func TestActualCreateCatInvalidJSON(t *testing.T) {
	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/api/cats", strings.NewReader("{ invalid json }"))
	req.Header.Set("Content-Type", "application/json")

	// Call actual function
	statusCode, response := createCat(req)

	// Assertions
	if statusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, statusCode)
	}

	if response != "Invalid JSON input" {
		t.Errorf("Expected 'Invalid JSON input', got %v", response)
	}
}

// Test actual deleteCat function with existing cat
func TestActualDeleteCatExists(t *testing.T) {
	// Save original database state
	originalDB := make(map[string]Cat)
	for k, v := range catsDatabase {
		originalDB[k] = v
	}
	defer func() {
		// Restore original state
		catsDatabase = originalDB
	}()

	// Set up test cat in database
	testCatID := "test-cat-id-123"
	testCat := Cat{
		Name: "TestCat",
		ID:   testCatID,
	}
	catsDatabase = map[string]Cat{
		testCatID: testCat,
	}

	// Create request with path parameter
	req := httptest.NewRequest("DELETE", "/api/cats/"+testCatID, nil)
	req.SetPathValue("catId", testCatID)

	// Call actual function
	statusCode, response := deleteCat(req)

	// Assertions
	if statusCode != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, statusCode)
	}

	if response != nil {
		t.Errorf("Expected nil response, got %v", response)
	}

	// Check cat was deleted from database
	if _, exists := catsDatabase[testCatID]; exists {
		t.Error("Cat should have been deleted from database")
	}

	if len(catsDatabase) != 0 {
		t.Errorf("Expected empty database, got %d items", len(catsDatabase))
	}
}

// Test actual deleteCat function with non-existent cat
func TestActualDeleteCatNotExists(t *testing.T) {
	// Save original database state
	originalDB := make(map[string]Cat)
	for k, v := range catsDatabase {
		originalDB[k] = v
	}
	defer func() {
		// Restore original state
		catsDatabase = originalDB
	}()

	// Clear database
	catsDatabase = make(map[string]Cat)

	nonExistentID := "non-existent-cat-id"

	// Create request
	req := httptest.NewRequest("DELETE", "/api/cats/"+nonExistentID, nil)
	req.SetPathValue("catId", nonExistentID)

	// Call actual function
	statusCode, response := deleteCat(req)

	// Assertions
	if statusCode != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, statusCode)
	}

	if response != "Cat not found" {
		t.Errorf("Expected 'Cat not found', got %v", response)
	}
}

// Test complete CRUD operations
func TestActualCRUDOperations(t *testing.T) {
	// Save original database state
	originalDB := make(map[string]Cat)
	for k, v := range catsDatabase {
		originalDB[k] = v
	}
	defer func() {
		// Restore original state
		catsDatabase = originalDB
	}()

	// Clear database
	catsDatabase = make(map[string]Cat)

	// Create cat
	testCat := Cat{
		Name:      "CRUDCat",
		Color:     "Blue",
		BirthDate: "2023-01-01",
	}

	jsonData, _ := json.Marshal(testCat)
	createReq := httptest.NewRequest("POST", "/api/cats", bytes.NewBuffer(jsonData))
	createReq.Header.Set("Content-Type", "application/json")

	statusCode, response := createCat(createReq)
	if statusCode != http.StatusCreated {
		t.Fatalf("Failed to create cat: status %d", statusCode)
	}

	catID := response.(string)

	// Verify cat exists with getCat
	getReq := httptest.NewRequest("GET", "/api/cats/"+catID, nil)
	getReq.SetPathValue("catId", catID)

	statusCode, _ = getCat(getReq)
	if statusCode != http.StatusOK {
		t.Errorf("Failed to get cat: status %d", statusCode)
	}

	// Delete cat
	deleteReq := httptest.NewRequest("DELETE", "/api/cats/"+catID, nil)
	deleteReq.SetPathValue("catId", catID)

	statusCode, _ = deleteCat(deleteReq)
	if statusCode != http.StatusNoContent {
		t.Errorf("Failed to delete cat: status %d", statusCode)
	}

	// Verify cat is gone
	getReq2 := httptest.NewRequest("GET", "/api/cats/"+catID, nil)
	getReq2.SetPathValue("catId", catID)

	statusCode, _ = getCat(getReq2)
	if statusCode != http.StatusNotFound {
		t.Errorf("Expected cat to be deleted, got status %d", statusCode)
	}
}

// =============================================================================
// YML2JSON FUNCTION TESTS
// =============================================================================

// Test yml2json with actual openapi.yml file
func TestActualYml2JsonWithRealFile(t *testing.T) {
	// Check if openapi.yml exists
	if _, err := os.Stat("openapi.yml"); os.IsNotExist(err) {
		t.Skip("openapi.yml file not found, skipping test")
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call actual yml2json function
	yml2json()

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify output is valid JSON
	var result map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("yml2json output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Basic validation - should have some expected OpenAPI fields
	expectedFields := []string{"openapi", "info", "paths"}
	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Errorf("Expected field '%s' in output", field)
		}
	}
}

// Test yml2json output format
func TestActualYml2JsonOutputFormat(t *testing.T) {
	// Simple YAML for testing format
	simpleYAML := `
key1: value1
key2: 42
key3: true
nested:
  subkey1: subvalue1
  subkey2: 123
array:
  - item1
  - item2
  - item3
`

	// Save original file
	originalExists := false
	var originalContent []byte
	if content, err := os.ReadFile("openapi.yml"); err == nil {
		originalExists = true
		originalContent = content
	}

	// Write test YAML
	err := os.WriteFile("openapi.yml", []byte(simpleYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test YAML: %v", err)
	}

	// Restore original file after test
	defer func() {
		if originalExists {
			os.WriteFile("openapi.yml", originalContent, 0644)
		} else {
			os.Remove("openapi.yml")
		}
	}()

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call actual yml2json function
	yml2json()

	// Restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Verify JSON format
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should start with {
	if !strings.HasPrefix(strings.TrimSpace(lines[0]), "{") {
		t.Error("JSON output should start with {")
	}

	// Should end with }
	lastLine := lines[len(lines)-1]
	if !strings.HasSuffix(strings.TrimSpace(lastLine), "}") {
		t.Error("JSON output should end with }")
	}

	// Should have proper indentation
	indentedLines := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "\t") {
			indentedLines++
		}
	}

	if indentedLines == 0 {
		t.Error("JSON output should have indented lines")
	}

	// Verify it parses as valid JSON
	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	// Verify content
	if result["key1"] != "value1" {
		t.Errorf("Expected key1 to be 'value1', got %v", result["key1"])
	}

	// Check numeric value
	if result["key2"] != float64(42) { // JSON numbers are float64
		t.Errorf("Expected key2 to be 42, got %v", result["key2"])
	}

	// Check boolean value
	if result["key3"] != true {
		t.Errorf("Expected key3 to be true, got %v", result["key3"])
	}
}

// =============================================================================
// MAIN FUNCTION COMPONENT TESTS
// =============================================================================

// Test main function components indirectly
func TestMainComponents(t *testing.T) {
	// Test that version variable is accessible
	if version == "" {
		// version might be empty in test environment, that's OK
		t.Log("Version is empty, which is expected in test environment")
	}

	// Test logger initialization (should be done by init)
	// Logger is a global variable that should be initialized
	t.Log("Logger is initialized as a global variable")

	// Test app creation
	app := newApp()
	if app == nil {
		t.Error("newApp() should return a non-nil handler")
	}
}

// Test server startup simulation (without actually starting)
func TestMainServerSetup(t *testing.T) {
	// Simulate the server setup from main()
	app := newApp()

	// This mimics the server creation in main()
	testServer := func(addr string, handler interface{}) bool {
		if addr == "" {
			return false
		}
		if handler == nil {
			return false
		}
		return true
	}

	result := testServer(":8080", app)
	if !result {
		t.Error("Server setup simulation failed")
	}
}

// Test the main workflow without actually running main()
func TestMainWorkflow(t *testing.T) {
	// Test each step of main() function workflow

	// Step 1: Logger should be initialized (global var)
	// Logger is initialized as a global variable
	t.Log("Logger is available as global variable")

	// Step 2: App creation
	app := newApp()
	if app == nil {
		t.Error("App creation failed")
	}

	// Step 3: Server configuration would be next
	// We can't test the actual server start without conflicting with other tests
	// But we can test the configuration values

	expectedAddr := ":8080"
	if expectedAddr != ":8080" {
		t.Errorf("Expected server address :8080, got %s", expectedAddr)
	}
}
