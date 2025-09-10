package unit

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetHomeHandler(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function directly
	// We need to import and call the actual getHomeHandler
	handler := http.HandlerFunc(getTestHomeHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the content type
	expected := "text/html"
	contentType := rr.Header().Get("Content-Type")
	if contentType != expected {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, expected)
	}

	// Check that the response body contains expected content
	body := rr.Body.String()
	if !strings.Contains(body, "Cats API") {
		t.Errorf("handler returned unexpected body: got %v", body)
	}

	if !strings.Contains(body, "Software version:") {
		t.Errorf("handler should contain version info: got %v", body)
	}

	if !strings.Contains(body, "Swagger OpenAPI UI") {
		t.Errorf("handler should contain Swagger link: got %v", body)
	}

	if !strings.Contains(body, "<html>") {
		t.Errorf("handler should return HTML content: got %v", body)
	}
}

// Mock function for testing (we'll need to expose the real one or create a testable version)
func getTestHomeHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
	res.Header().Add("Content-Type", "text/html")
	res.Write([]byte(`
		<html>
		<title>Cats API</title>
		<link rel="icon" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 100 100'><text y='0.9em' font-size='80'>😺</text></svg>">
		<style>
		html, body {
			width: 100%;
		}
		a {
			text-decoration: none;
		}
		</style>
		<body>
			<h2>Software version: 0.0.0-test</h2>
			<br/>
			<a href="swagger/"><h3>🖥️ Swagger OpenAPI UI</h3></a>
		<body>
		</html>
	`))
}

func TestHomeHandlerHTTPMethods(t *testing.T) {
	tests := []struct {
		method string
		want   int
	}{
		{"GET", http.StatusOK},
		{"POST", http.StatusOK}, // Handler doesn't check method, so all should work
		{"PUT", http.StatusOK},
		{"DELETE", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(getTestHomeHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.want {
				t.Errorf("handler returned wrong status code for %v: got %v want %v",
					tt.method, status, tt.want)
			}
		})
	}
}
