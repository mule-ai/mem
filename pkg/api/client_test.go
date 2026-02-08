package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestNewClient tests the creation of a new API client
func TestNewClient(t *testing.T) {
	t.Run("DefaultConfiguration", func(t *testing.T) {
		config := ClientConfig{
			BaseURL: "http://localhost:8080",
		}
		
		client := NewClient(config)
		
		if client.baseURL != "http://localhost:8080" {
			t.Errorf("Expected baseURL 'http://localhost:8080', got '%s'", client.baseURL)
		}
		if client.httpClient.Timeout != defaultTimeout {
			t.Errorf("Expected timeout %v, got %v", defaultTimeout, client.httpClient.Timeout)
		}
		if client.maxRetries != defaultMaxRetries {
			t.Errorf("Expected maxRetries %d, got %d", defaultMaxRetries, client.maxRetries)
		}
		if client.retryDelay != defaultRetryDelay {
			t.Errorf("Expected retryDelay %v, got %v", defaultRetryDelay, client.retryDelay)
		}
	})

	t.Run("CustomConfiguration", func(t *testing.T) {
		config := ClientConfig{
			BaseURL:    "http://example.com",
			APIKey:     "test-key",
			Timeout:    10 * time.Second,
			MaxRetries: 5,
			RetryDelay: 1 * time.Second,
		}
		
		client := NewClient(config)
		
		if client.baseURL != "http://example.com" {
			t.Errorf("Expected baseURL 'http://example.com', got '%s'", client.baseURL)
		}
		if client.apiKey != "test-key" {
			t.Errorf("Expected apiKey 'test-key', got '%s'", client.apiKey)
		}
		if client.httpClient.Timeout != 10*time.Second {
			t.Errorf("Expected timeout 10s, got %v", client.httpClient.Timeout)
		}
		if client.maxRetries != 5 {
			t.Errorf("Expected maxRetries 5, got %d", client.maxRetries)
		}
		if client.retryDelay != 1*time.Second {
			t.Errorf("Expected retryDelay 1s, got %v", client.retryDelay)
		}
	})
}

// TestClientDo tests the Do method for making HTTP requests
func TestClientDo(t *testing.T) {
	t.Run("SuccessfulGETRequest", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}
			
			if r.URL.Path != "/test" {
				t.Errorf("Expected path /test, got %s", r.URL.Path)
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "success"})
		}))
		defer server.Close()
		
		client := NewClient(ClientConfig{BaseURL: server.URL})
		
		var result map[string]string
		err := client.Get(context.Background(), "/test", &result)
		
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if result["message"] != "success" {
			t.Errorf("Expected message 'success', got '%s'", result["message"])
		}
	})

	t.Run("SuccessfulPOSTRequest", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"status": "created"})
		}))
		defer server.Close()
		
		client := NewClient(ClientConfig{BaseURL: server.URL})
		
		body := map[string]string{"key": "value"}
		var result map[string]string
		err := client.Post(context.Background(), "/create", body, &result)
		
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if result["status"] != "created" {
			t.Errorf("Expected status 'created', got '%s'", result["status"])
		}
	})

	t.Run("RequestWithAPIKey", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth != "Bearer test-api-key" {
				t.Errorf("Expected Authorization 'Bearer test-api-key', got '%s'", auth)
			}
			
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]bool{"authenticated": true})
		}))
		defer server.Close()
		
		client := NewClient(ClientConfig{
			BaseURL: server.URL,
			APIKey:  "test-api-key",
		})
		
		var result map[string]bool
		err := client.Get(context.Background(), "/auth", &result)
		
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	t.Run("HTTPErrorNotRetried", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "bad request"})
		}))
		defer server.Close()
		
		client := NewClient(ClientConfig{
			BaseURL:    server.URL,
			MaxRetries: 3,
		})
		
		var result map[string]string
		err := client.Get(context.Background(), "/test", &result)
		
		if err == nil {
			t.Fatal("Expected error for bad request, got nil")
		}
		
		if requestCount != 1 {
			t.Errorf("Expected 1 request for 4xx error, got %d", requestCount)
		}
	})

	t.Run("ServerErrorRetried", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal error"})
		}))
		defer server.Close()
		
		client := NewClient(ClientConfig{
			BaseURL:    server.URL,
			MaxRetries: 2,
			RetryDelay: 10 * time.Millisecond,
		})
		
		var result map[string]string
		err := client.Get(context.Background(), "/test", &result)
		
		if err == nil {
			t.Fatal("Expected error for server error, got nil")
		}
		
		// Should attempt 3 times total (initial + 2 retries)
		if requestCount != 3 {
			t.Errorf("Expected 3 requests for 5xx error with maxRetries=2, got %d", requestCount)
		}
	})

	t.Run("NetworkErrorRetried", func(t *testing.T) {
		client := NewClient(ClientConfig{
			BaseURL:    "http://invalid.example.com:99999",
			MaxRetries: 2,
			RetryDelay: 10 * time.Millisecond,
			Timeout:    100 * time.Millisecond,
		})
		
		var result map[string]string
		err := client.Get(context.Background(), "/test", &result)
		
		if err == nil {
			t.Fatal("Expected error for network failure, got nil")
		}
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		requestCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount++
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		
		client := NewClient(ClientConfig{
			BaseURL:    server.URL,
			MaxRetries: 5,
			RetryDelay: 50 * time.Millisecond,
		})
		
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		
		var result map[string]string
		err := client.Get(ctx, "/test", &result)
		
		if err == nil {
			t.Fatal("Expected error for context cancellation, got nil")
		}
	})
}

// TestGetAndPostConvenienceMethods tests Get and Post methods
func TestGetAndPostConvenienceMethods(t *testing.T) {
	t.Run("GetMethod", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET, got %s", r.Method)
			}
			json.NewEncoder(w).Encode(map[string]string{"method": "GET"})
		}))
		defer server.Close()
		
		client := NewClient(ClientConfig{BaseURL: server.URL})
		
		var result map[string]string
		err := client.Get(context.Background(), "/test", &result)
		
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if result["method"] != "GET" {
			t.Errorf("Expected method 'GET', got '%s'", result["method"])
		}
	})

	t.Run("PostMethod", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST, got %s", r.Method)
			}
			json.NewEncoder(w).Encode(map[string]string{"method": "POST"})
		}))
		defer server.Close()
		
		client := NewClient(ClientConfig{BaseURL: server.URL})
		
		var result map[string]string
		err := client.Post(context.Background(), "/test", nil, &result)
		
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		
		if result["method"] != "POST" {
			t.Errorf("Expected method 'POST', got '%s'", result["method"])
		}
	})
}
