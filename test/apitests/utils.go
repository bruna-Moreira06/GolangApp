//go:build integration

package apitests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type CatModel struct {
	Name      string `json:"name"`
	ID        string `json:"id,omitempty"`
	BirthDate string `json:"birthDate,omitempty"`
	Color     string `json:"color,omitempty"`
}

var baseUrl = "http://localhost:8080/api"

// Global client with a proper timeout
var client = &http.Client{Timeout: 10 * time.Second}

// Wrapper to HTTP API calls, does the error handling and JSON decoding
func call(method, path string, reqBody any, code *int, result any) error {

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, baseUrl+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	// Set appropriate headers
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// send the request
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if code != nil {
		*code = res.StatusCode
	}

	if result != nil {
		err = json.NewDecoder(res.Body).Decode(result)
		// Don't treat JSON decode errors as fatal for API tests
		// Sometimes we get plain text responses for error cases
	}

	return err
}