package unit

import (
	"testing"
)

func TestMapKeysExtraction(t *testing.T) {
	// Test the concept of listing map keys (used in allCatsHandlers.go)
	testMaps := []map[string]interface{}{
		{"cat1": "Fluffy", "cat2": "Whiskers"},
		{"a": 1, "b": 2, "c": 3},
		{},
		{"single": "key"},
	}

	expectedLengths := []int{2, 3, 0, 1}

	for i, testMap := range testMaps {
		keys := getMapKeys(testMap)
		if len(keys) != expectedLengths[i] {
			t.Errorf("Map %d: expected %d keys, got %d", i, expectedLengths[i], len(keys))
		}

		// Verify all keys are present
		for key := range testMap {
			found := false
			for _, k := range keys {
				if k == key {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Map %d: key '%s' not found in result", i, key)
			}
		}
	}
}

func TestCatDataStructures(t *testing.T) {
	// Test cat data structure handling
	cats := map[string]interface{}{
		"1": map[string]interface{}{"name": "Fluffy", "age": 3},
		"2": map[string]interface{}{"name": "Whiskers", "age": 5},
		"3": map[string]interface{}{"name": "Shadow", "age": 2},
	}

	// Test getting all cat IDs
	ids := getMapKeys(cats)
	if len(ids) != 3 {
		t.Errorf("Expected 3 cats, got %d", len(ids))
	}

	// Test individual cat access
	for id, catData := range cats {
		cat, ok := catData.(map[string]interface{})
		if !ok {
			t.Errorf("Cat %s: data should be a map", id)
			continue
		}

		name, exists := cat["name"]
		if !exists {
			t.Errorf("Cat %s: should have a name", id)
		}

		age, exists := cat["age"]
		if !exists {
			t.Errorf("Cat %s: should have an age", id)
		}

		// Verify data types
		if _, ok := name.(string); !ok {
			t.Errorf("Cat %s: name should be string, got %T", id, name)
		}

		if _, ok := age.(int); !ok {
			t.Errorf("Cat %s: age should be int, got %T", id, age)
		}
	}
}

func TestCRUDOperationsConcepts(t *testing.T) {
	// Test CRUD operations concepts for cats
	cats := make(map[string]interface{})

	// Test Create
	newCat := map[string]interface{}{
		"name": "TestCat",
		"age":  1,
	}
	cats["test1"] = newCat

	if len(cats) != 1 {
		t.Errorf("After create: expected 1 cat, got %d", len(cats))
	}

	// Test Read
	cat, exists := cats["test1"]
	if !exists {
		t.Error("After create: cat should exist")
	}

	catMap, ok := cat.(map[string]interface{})
	if !ok {
		t.Error("Cat should be a map")
	} else {
		if catMap["name"] != "TestCat" {
			t.Errorf("Expected name 'TestCat', got %v", catMap["name"])
		}
	}

	// Test Update
	updatedCat := map[string]interface{}{
		"name": "UpdatedCat",
		"age":  2,
	}
	cats["test1"] = updatedCat

	cat = cats["test1"]
	catMap, _ = cat.(map[string]interface{})
	if catMap["name"] != "UpdatedCat" {
		t.Errorf("After update: expected name 'UpdatedCat', got %v", catMap["name"])
	}

	// Test Delete
	delete(cats, "test1")
	if len(cats) != 0 {
		t.Errorf("After delete: expected 0 cats, got %d", len(cats))
	}

	_, exists = cats["test1"]
	if exists {
		t.Error("After delete: cat should not exist")
	}
}

func TestURLPathParameterExtraction(t *testing.T) {
	// Test path parameter extraction concept (used in oneCatHandlers.go)
	testPaths := []struct {
		path     string
		expected string
		valid    bool
	}{
		{"/cats/123", "123", true},
		{"/cats/abc", "abc", true},
		{"/cats/", "", false},
		{"/cats", "", false},
		{"/cats/123/extra", "123", true}, // Would extract first parameter
	}

	for _, test := range testPaths {
		t.Run(test.path, func(t *testing.T) {
			// Mock path parameter extraction
			param := extractPathParam(test.path, "/cats/")

			if test.valid {
				if param != test.expected {
					t.Errorf("Expected param '%s', got '%s'", test.expected, param)
				}
			} else {
				if param != "" {
					t.Errorf("Expected empty param for invalid path, got '%s'", param)
				}
			}
		})
	}
}

// Helper functions for testing concepts

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func extractPathParam(path, prefix string) string {
	if len(path) <= len(prefix) {
		return ""
	}

	if path[:len(prefix)] != prefix {
		return ""
	}

	param := path[len(prefix):]

	// Extract until next slash
	for i, char := range param {
		if char == '/' {
			param = param[:i]
			break
		}
	}

	return param
}
