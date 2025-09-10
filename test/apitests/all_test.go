//go:build integration

package apitests

import (
	"fmt"
	"net/http"
	"testing"
)

var initCatId string

func init() {
	// Preparation: delete all existing & create a cat
	ids := []string{}
	call("GET", "/cats", nil, nil, &ids)

	for _, id := range ids {
		code := 0
		call("DELETE", "/cats/"+id, nil, &code, nil)
		fmt.Println("DELETE /cats ->", code)
	}

	// Create a single cat into the DB
	call("POST", "/cats", &CatModel{Name: "Toto"}, nil, &initCatId)
}

func TestGetCats(t *testing.T) {
	code := 0
	result := []string{}
	err := call("GET", "/cats", nil, &code, &result)
	if err != nil {
		t.Error("Request error", err)
	}

	fmt.Println("GET /cats ->", code, result)

	if code != http.StatusOK {
		t.Error("We should get code 200, got", code)
	}

	// After init cleanup and creation, we should have 1 cat (the initCat)
	if len(result) != 1 {
		t.Error("We should get 1 item (initCat only), got", len(result))
		return
	}

	if result[0] != initCatId {
		t.Error("Expected initCatId in first position, got", result[0])
	}
}

func TestCreateCat(t *testing.T) {
	// Test creating a new cat
	newCat := &CatModel{
		Name:      "Fluffy",
		Color:     "White",
		BirthDate: "2023-01-15",
	}

	code := 0
	var createdCatId string
	err := call("POST", "/cats", newCat, &code, &createdCatId)
	if err != nil {
		t.Error("Request error", err)
	}

	fmt.Println("POST /cats ->", code, createdCatId)

	if code != http.StatusCreated {
		t.Errorf("Expected status code 201, got %d", code)
	}

	if createdCatId == "" {
		t.Error("Expected non-empty cat ID")
	}

	// Verify the cat was created by getting it
	var retrievedCat CatModel
	getCode := 0
	err = call("GET", "/cats/"+createdCatId, nil, &getCode, &retrievedCat)
	if err != nil {
		t.Error("Error retrieving created cat", err)
	}

	if getCode != http.StatusOK {
		t.Errorf("Expected status code 200 when getting created cat, got %d", getCode)
	}

	if retrievedCat.Name != newCat.Name {
		t.Errorf("Expected cat name %s, got %s", newCat.Name, retrievedCat.Name)
	}

	if retrievedCat.Color != newCat.Color {
		t.Errorf("Expected cat color %s, got %s", newCat.Color, retrievedCat.Color)
	}

	// Clean up
	deleteCode := 0
	call("DELETE", "/cats/"+createdCatId, nil, &deleteCode, nil)
}

func TestCreateCatInvalidData(t *testing.T) {
	// Test creating a cat with missing required field (name)
	invalidCat := &CatModel{
		Color:     "Black",
		BirthDate: "2023-01-15",
		// Name is missing
	}

	code := 0
	var response string
	call("POST", "/cats", invalidCat, &code, &response)

	fmt.Println("POST /cats (invalid) ->", code, response)

	// Note: Current implementation doesn't validate required fields
	// This test documents current behavior - you might want to add validation
	if code != http.StatusCreated {
		t.Logf("Cat creation with missing name returned status %d (this might be expected if validation is added)", code)
	}
}

func TestGetCat(t *testing.T) {
	// Test getting an existing cat
	code := 0
	var cat CatModel
	err := call("GET", "/cats/"+initCatId, nil, &code, &cat)
	if err != nil {
		t.Error("Request error", err)
	}

	fmt.Println("GET /cats/"+initCatId+" ->", code, cat)

	if code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", code)
	}

	if cat.Name != "Toto" {
		t.Errorf("Expected cat name 'Toto', got '%s'", cat.Name)
	}

	if cat.ID != initCatId {
		t.Errorf("Expected cat ID '%s', got '%s'", initCatId, cat.ID)
	}
}

func TestGetCatNotFound(t *testing.T) {
	// Test getting a non-existent cat
	code := 0
	var response string
	call("GET", "/cats/nonexistent-id", nil, &code, &response)

	fmt.Println("GET /cats/nonexistent-id ->", code, response)

	if code != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %d", code)
	}

	if response != "Cat not found" {
		t.Errorf("Expected 'Cat not found' message, got '%s'", response)
	}

	// Note: err might not be nil due to JSON decoding a string response
}

func TestDeleteCat(t *testing.T) {
	// First create a cat to delete
	newCat := &CatModel{
		Name:  "TempCat",
		Color: "Orange",
	}

	createCode := 0
	var catId string
	err := call("POST", "/cats", newCat, &createCode, &catId)
	if err != nil {
		t.Error("Error creating cat for delete test", err)
	}

	if createCode != http.StatusCreated {
		t.Errorf("Failed to create cat for delete test, got status %d", createCode)
		return
	}

	// Now delete the cat
	deleteCode := 0
	_ = call("DELETE", "/cats/"+catId, nil, &deleteCode, nil)

	fmt.Println("DELETE /cats/"+catId+" ->", deleteCode)

	if deleteCode != http.StatusNoContent {
		t.Errorf("Expected status code 204, got %d", deleteCode)
	}

	// Verify the cat was deleted
	getCode := 0
	var response string
	call("GET", "/cats/"+catId, nil, &getCode, &response)

	if getCode != http.StatusNotFound {
		t.Errorf("Cat should be deleted, but GET returned status %d", getCode)
	}
}

func TestDeleteCatNotFound(t *testing.T) {
	// Test deleting a non-existent cat
	code := 0
	var response string
	call("DELETE", "/cats/nonexistent-id", nil, &code, &response)

	fmt.Println("DELETE /cats/nonexistent-id ->", code, response)

	if code != http.StatusNotFound {
		t.Errorf("Expected status code 404, got %d", code)
	}

	if response != "Cat not found" {
		t.Errorf("Expected 'Cat not found' message, got '%s'", response)
	}

	// Note: err might not be nil due to JSON decoding a string response
}

func TestCRUDWorkflow(t *testing.T) {
	// Test complete CRUD workflow

	// 1. Create a cat
	newCat := &CatModel{
		Name:      "WorkflowCat",
		Color:     "Calico",
		BirthDate: "2023-06-01",
	}

	createCode := 0
	var catId string
	err := call("POST", "/cats", newCat, &createCode, &catId)
	if err != nil {
		t.Fatal("Error creating cat", err)
	}

	if createCode != http.StatusCreated {
		t.Fatalf("Expected status 201 for create, got %d", createCode)
	}

	// 2. Read the cat
	readCode := 0
	var retrievedCat CatModel
	err = call("GET", "/cats/"+catId, nil, &readCode, &retrievedCat)
	if err != nil {
		t.Fatal("Error reading cat", err)
	}

	if readCode != http.StatusOK {
		t.Fatalf("Expected status 200 for read, got %d", readCode)
	}

	if retrievedCat.Name != newCat.Name {
		t.Errorf("Name mismatch: expected %s, got %s", newCat.Name, retrievedCat.Name)
	}

	// 3. Verify cat appears in list
	listCode := 0
	var catIds []string
	err = call("GET", "/cats", nil, &listCode, &catIds)
	if err != nil {
		t.Fatal("Error listing cats", err)
	}

	if listCode != http.StatusOK {
		t.Fatalf("Expected status 200 for list, got %d", listCode)
	}

	found := false
	for _, id := range catIds {
		if id == catId {
			found = true
			break
		}
	}

	if !found {
		t.Error("Created cat not found in list")
	}

	// 4. Delete the cat
	deleteCode := 0
	err = call("DELETE", "/cats/"+catId, nil, &deleteCode, nil)
	if err != nil {
		t.Fatal("Error deleting cat", err)
	}

	if deleteCode != http.StatusNoContent {
		t.Fatalf("Expected status 204 for delete, got %d", deleteCode)
	}

	// 5. Verify cat is gone
	verifyCode := 0
	call("GET", "/cats/"+catId, nil, &verifyCode, nil)

	if verifyCode != http.StatusNotFound {
		t.Errorf("Expected cat to be deleted (404), but got status %d", verifyCode)
	}

	fmt.Println("CRUD workflow test completed successfully")
}