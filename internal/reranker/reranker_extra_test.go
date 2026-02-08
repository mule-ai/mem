package reranker

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewClientExtended tests the creation of a new reranker client
func TestNewClientExtended(t *testing.T) {
	t.Run("DefaultConfiguration", func(t *testing.T) {
		config := Config{}
		client := NewClient(config)
		
		if client.config.BaseURL != defaultEndpoint {
			t.Errorf("Expected BaseURL '%s', got '%s'", defaultEndpoint, client.config.BaseURL)
		}
		if client.config.Model != defaultModel {
			t.Errorf("Expected Model '%s', got '%s'", defaultModel, client.config.Model)
		}
		if client.config.TopK != 10 {
			t.Errorf("Expected TopK 10, got %d", client.config.TopK)
		}
	})

	t.Run("CustomConfiguration", func(t *testing.T) {
		config := Config{
			BaseURL:   "http://custom-endpoint",
			Model:     "custom-model",
			APIKey:    "test-key",
			Enabled:   true,
			TopK:      20,
			Threshold: 0.7,
		}
		client := NewClient(config)
		
		if client.config.BaseURL != "http://custom-endpoint" {
			t.Errorf("Expected BaseURL 'http://custom-endpoint', got '%s'", client.config.BaseURL)
		}
		if client.config.Model != "custom-model" {
			t.Errorf("Expected Model 'custom-model', got '%s'", client.config.Model)
		}
		if client.config.APIKey != "test-key" {
			t.Errorf("Expected APIKey 'test-key', got '%s'", client.config.APIKey)
		}
		if !client.config.Enabled {
			t.Error("Expected Enabled to be true")
		}
		if client.config.TopK != 20 {
			t.Errorf("Expected TopK 20, got %d", client.config.TopK)
		}
		if client.config.Threshold != 0.7 {
			t.Errorf("Expected Threshold 0.7, got %f", client.config.Threshold)
		}
	})
}

// TestNoopRerankExtended tests the noop reranking (when disabled)
func TestNoopRerankExtended(t *testing.T) {
	config := Config{Enabled: false}
	client := NewClient(config)
	
	documents := []DocumentToRerank{
		{ID: "doc1", Content: "Content 1", Score: 0.8},
		{ID: "doc2", Content: "Content 2", Score: 0.7},
		{ID: "doc3", Content: "Content 3", Score: 0.6},
	}
	
	result, err := client.Rerank(context.Background(), "test query", documents)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(result) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(result))
	}
	
	for i, doc := range result {
		if doc.Rank != i+1 {
			t.Errorf("Expected rank %d, got %d", i+1, doc.Rank)
		}
		if doc.FinalScore != documents[i].Score {
			t.Errorf("Expected FinalScore %f, got %f", documents[i].Score, doc.FinalScore)
		}
		if doc.RerankScore != float64(documents[i].Score) {
			t.Errorf("Expected RerankScore %f, got %f", float64(documents[i].Score), doc.RerankScore)
		}
	}
}

// TestRerankWithEmptyDocuments tests reranking with empty document list
func TestRerankWithEmptyDocuments(t *testing.T) {
	config := Config{Enabled: true}
	client := NewClient(config)
	
	result, err := client.Rerank(context.Background(), "test query", []DocumentToRerank{})
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(result) != 0 {
		t.Errorf("Expected 0 results, got %d", len(result))
	}
}

// TestRerankWithRealAPI tests reranking with a mock API server
func TestRerankWithRealAPI(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/rerank" {
			t.Errorf("Expected path '/v1/rerank', got '%s'", r.URL.Path)
		}
		
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		
		// Return a mock rerank response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"object": "list",
			"model": "test-model",
			"results": [
				{
					"index": 1,
					"relevance_score": 0.95,
					"document": {
						"id": "doc2",
						"text": "Content 2"
					}
				},
				{
					"index": 0,
					"relevance_score": 0.75,
					"document": {
						"id": "doc1",
						"text": "Content 1"
					}
				}
			],
			"usage": {
				"total_tokens": 100
			}
		}`))
	}))
	defer server.Close()
	
	config := Config{
		BaseURL: server.URL,
		Enabled: true,
		TopK:    10,
	}
	client := NewClient(config)
	
	documents := []DocumentToRerank{
		{ID: "doc1", Content: "Content 1", Score: 0.8},
		{ID: "doc2", Content: "Content 2", Score: 0.7},
	}
	
	result, err := client.Rerank(context.Background(), "test query", documents)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(result) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(result))
	}
	
	// Check that results are reranked correctly
	if result[0].ID != "doc2" {
		t.Errorf("Expected first result to be 'doc2', got '%s'", result[0].ID)
	}
	if result[0].Rank != 1 {
		t.Errorf("Expected rank 1 for first result, got %d", result[0].Rank)
	}
	if result[0].RerankScore != 0.95 {
		t.Errorf("Expected rerank score 0.95, got %f", result[0].RerankScore)
	}
}

// TestRerankGracefulDegradation tests that reranking gracefully degrades on error
func TestRerankGracefulDegradation(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()
	
	config := Config{
		BaseURL: server.URL,
		Enabled: true,
	}
	client := NewClient(config)
	
	documents := []DocumentToRerank{
		{ID: "doc1", Content: "Content 1", Score: 0.8},
		{ID: "doc2", Content: "Content 2", Score: 0.7},
	}
	
	// Should not return an error, but fall back to original scores
	result, err := client.Rerank(context.Background(), "test query", documents)
	
	if err != nil {
		t.Fatalf("Unexpected error (should degrade gracefully): %v", err)
	}
	
	if len(result) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(result))
	}
	
	// Should return original scores (no reranking)
	for i, doc := range result {
		if doc.FinalScore != documents[i].Score {
			t.Errorf("Expected original score %f, got %f", documents[i].Score, doc.FinalScore)
		}
	}
}

// TestRerankWithThreshold tests filtering by threshold
func TestRerankWithThreshold(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"results": [
				{
					"index": 0,
					"relevance_score": 0.9,
					"document": {"id": "doc1", "text": "Content 1"}
				},
				{
					"index": 1,
					"relevance_score": 0.4,
					"document": {"id": "doc2", "text": "Content 2"}
				}
			]
		}`))
	}))
	defer server.Close()
	
	config := Config{
		BaseURL:   server.URL,
		Enabled:   true,
		Threshold: 0.5, // Filter out results below 0.5
	}
	client := NewClient(config)
	
	documents := []DocumentToRerank{
		{ID: "doc1", Content: "Content 1", Score: 0.8},
		{ID: "doc2", Content: "Content 2", Score: 0.7},
	}
	
	result, err := client.Rerank(context.Background(), "test query", documents)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	// Only doc1 should be returned (score >= 0.5)
	if len(result) != 1 {
		t.Fatalf("Expected 1 result (filtered by threshold), got %d", len(result))
	}
	
	if result[0].ID != "doc1" {
		t.Errorf("Expected 'doc1', got '%s'", result[0].ID)
	}
}

// TestIsEnabled tests the IsEnabled method
func TestIsEnabled(t *testing.T) {
	t.Run("Enabled", func(t *testing.T) {
		config := Config{Enabled: true}
		client := NewClient(config)
		
		if !client.IsEnabled() {
			t.Error("Expected IsEnabled to return true")
		}
	})

	t.Run("Disabled", func(t *testing.T) {
		config := Config{Enabled: false}
		client := NewClient(config)
		
		if client.IsEnabled() {
			t.Error("Expected IsEnabled to return false")
		}
	})
}

// TestGetTopK tests the GetTopK method
func TestGetTopK(t *testing.T) {
	config := Config{TopK: 25}
	client := NewClient(config)
	
	if client.GetTopK() != 25 {
		t.Errorf("Expected GetTopK to return 25, got %d", client.GetTopK())
	}
}

// TestRerankWithEmptyResponse tests handling of empty API responses
func TestRerankWithEmptyResponse(t *testing.T) {
	// Create a mock server that returns an empty response (LM Studio behavior)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()
	
	config := Config{
		BaseURL: server.URL,
		Enabled: true,
	}
	client := NewClient(config)
	
	documents := []DocumentToRerank{
		{ID: "doc1", Content: "Content 1", Score: 0.8},
	}
	
	// Should fall back to noop rerank
	result, err := client.Rerank(context.Background(), "test query", documents)
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
	
	if result[0].FinalScore != 0.8 {
		t.Errorf("Expected original score 0.8, got %f", result[0].FinalScore)
	}
}
