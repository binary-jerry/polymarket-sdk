package common

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	// Test with nil config
	client := NewHTTPClient(nil)
	if client == nil {
		t.Error("NewHTTPClient(nil) should not return nil")
	}

	// Test with custom config
	config := &HTTPClientConfig{
		BaseURL:      "https://example.com",
		Timeout:      10 * time.Second,
		MaxRetries:   5,
		RetryDelayMs: 500,
	}
	client = NewHTTPClient(config)
	if client.baseURL != "https://example.com" {
		t.Errorf("BaseURL = %s, expected https://example.com", client.baseURL)
	}
	if client.maxRetries != 5 {
		t.Errorf("MaxRetries = %d, expected 5", client.maxRetries)
	}
}

func TestHTTPClientGet(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check query params
		if r.URL.Query().Get("key") != "value" {
			t.Errorf("Expected query param key=value, got %s", r.URL.Query().Get("key"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	})

	type Params struct {
		Key string `url:"key"`
	}
	type Result struct {
		Result string `json:"result"`
	}

	var result Result
	err := client.Get(context.Background(), "/test", &Params{Key: "value"}, &result)
	if err != nil {
		t.Errorf("Get() error: %v", err)
	}
	if result.Result != "success" {
		t.Errorf("Result = %s, expected success", result.Result)
	}
}

func TestHTTPClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["name"] != "test" {
			t.Errorf("Expected body.name = test, got %s", body["name"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	})

	type Result struct {
		ID string `json:"id"`
	}

	var result Result
	err := client.Post(context.Background(), "/test", map[string]string{"name": "test"}, &result)
	if err != nil {
		t.Errorf("Post() error: %v", err)
	}
	if result.ID != "123" {
		t.Errorf("Result.ID = %s, expected 123", result.ID)
	}
}

func TestHTTPClientDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	})

	err := client.Delete(context.Background(), "/test/123", nil)
	if err != nil {
		t.Errorf("Delete() error: %v", err)
	}
}

func TestHTTPClientErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error":   "BAD_REQUEST",
			"message": "Invalid input",
		})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 0, // No retries
	})

	var result map[string]string
	err := client.Get(context.Background(), "/test", nil, &result)
	if err == nil {
		t.Error("Expected error for 400 response")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Errorf("Expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, expected 400", apiErr.StatusCode)
	}
}

func TestHTTPClientWithAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token123" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	})

	authHeaders := map[string]string{
		"Authorization": "Bearer token123",
	}

	var result map[string]string
	err := client.DoWithAuth(context.Background(), "GET", "/protected", nil, authHeaders, &result)
	if err != nil {
		t.Errorf("DoWithAuth() error: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("Result.status = %s, expected ok", result["status"])
	}
}

func TestHTTPClientSetDefaultHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Errorf("Expected X-Custom-Header = custom-value, got %s", r.Header.Get("X-Custom-Header"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	})
	client.SetDefaultHeader("X-Custom-Header", "custom-value")

	err := client.Get(context.Background(), "/test", nil, nil)
	if err != nil {
		t.Errorf("Get() error: %v", err)
	}
}

func TestHTTPClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL:    server.URL,
		Timeout:    5 * time.Second,
		MaxRetries: 0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := client.Get(ctx, "/slow", nil, nil)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
}

func TestStructToQueryString(t *testing.T) {
	type TestParams struct {
		Name     string `url:"name"`
		Age      int    `url:"age,omitempty"`
		Active   *bool  `url:"active,omitempty"`
		Ignored  string `url:"-"`
		NoTag    string
	}

	active := true
	params := TestParams{
		Name:    "test",
		Age:     25,
		Active:  &active,
		Ignored: "ignored",
		NoTag:   "notag",
	}

	result := structToQueryString(params)

	// Check that required params are present
	if result == "" {
		t.Error("Expected non-empty query string")
	}
}

func TestGetBaseURL(t *testing.T) {
	client := NewHTTPClient(&HTTPClientConfig{
		BaseURL: "https://api.example.com/v1",
	})

	if client.GetBaseURL() != "https://api.example.com/v1" {
		t.Errorf("GetBaseURL() = %s, expected https://api.example.com/v1", client.GetBaseURL())
	}
}
