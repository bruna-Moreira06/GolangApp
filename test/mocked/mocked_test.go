package mocked

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

type Cat struct {
	Name  string `json:"name"`
	ID    string `json:"id,omitempty"`
	Color string `json:"color,omitempty"`
}

type MockRepo struct {
	cats    map[string]*Cat
	counter int
}

func NewMockRepo() *MockRepo {
	return &MockRepo{
		cats:    make(map[string]*Cat),
		counter: 0,
	}
}

func (r *MockRepo) Create(cat *Cat) (*Cat, error) {
	r.counter++
	id := "mock-" + strconv.Itoa(r.counter)
	cat.ID = id
	r.cats[id] = cat
	return cat, nil
}

func (r *MockRepo) GetAll() []Cat {
	cats := make([]Cat, 0)
	for _, cat := range r.cats {
		cats = append(cats, *cat)
	}
	return cats
}

func TestBasicMockOperations(t *testing.T) {
	repo := NewMockRepo()

	// Test create
	cat := &Cat{Name: "TestCat", Color: "Black"}
	created, err := repo.Create(cat)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if created.ID == "" {
		t.Error("Expected ID to be set")
	}

	// Test get all
	cats := repo.GetAll()
	if len(cats) != 1 {
		t.Errorf("Expected 1 cat, got %d", len(cats))
	}
}

func TestMockHTTPOperations(t *testing.T) {
	repo := NewMockRepo()

	// Create a cat first
	cat := &Cat{Name: "HTTPCat", Color: "White"}
	repo.Create(cat)

	// Test HTTP list endpoint
	req := httptest.NewRequest("GET", "/cats", nil)
	w := httptest.NewRecorder()

	// Simple handler
	handler := func(w http.ResponseWriter, r *http.Request) {
		cats := repo.GetAll()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cats)
	}

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var cats []Cat
	err := json.Unmarshal(w.Body.Bytes(), &cats)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if len(cats) != 1 {
		t.Errorf("Expected 1 cat in response, got %d", len(cats))
	}
}
