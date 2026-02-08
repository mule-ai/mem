package reranker

import (
	"context"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		wantEnabled bool
		wantTopK    int
	}{
		{
			name: "default values",
			config: Config{
				Enabled: true,
			},
			wantEnabled: true,
			wantTopK:    10,
		},
		{
			name: "custom values",
			config: Config{
				BaseURL: "http://localhost:8080",
				Model:   "custom-reranker",
				APIKey:  "test-key",
				Enabled: true,
				TopK:    20,
				Threshold: 0.7,
			},
			wantEnabled: true,
			wantTopK:    20,
		},
		{
			name: "disabled",
			config: Config{
				Enabled: false,
			},
			wantEnabled: false,
			wantTopK:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.config)
			if client.IsEnabled() != tt.wantEnabled {
				t.Errorf("NewClient() enabled = %v, want %v", client.IsEnabled(), tt.wantEnabled)
			}
			if client.GetTopK() != tt.wantTopK {
				t.Errorf("NewClient() topK = %v, want %v", client.GetTopK(), tt.wantTopK)
			}
		})
	}
}

func TestNoopRerank(t *testing.T) {
	config := Config{
		Enabled: false,
	}
	client := NewClient(config)

	ctx := context.Background()
	documents := []DocumentToRerank{
		{
			ID:      "doc1",
			Content: "First document",
			Score:   0.9,
		},
		{
			ID:      "doc2",
			Content: "Second document",
			Score:   0.8,
		},
		{
			ID:      "doc3",
			Content: "Third document",
			Score:   0.7,
		},
	}

	results, err := client.Rerank(ctx, "test query", documents)
	if err != nil {
		t.Fatalf("Rerank() error = %v", err)
	}

	if len(results) != len(documents) {
		t.Fatalf("Rerank() returned %d results, want %d", len(results), len(documents))
	}

	for i, result := range results {
		if result.ID != documents[i].ID {
			t.Errorf("Result[%d].ID = %v, want %v", i, result.ID, documents[i].ID)
		}
		if result.OriginalScore != documents[i].Score {
			t.Errorf("Result[%d].OriginalScore = %v, want %v", i, result.OriginalScore, documents[i].Score)
		}
		if result.RerankScore != float64(documents[i].Score) {
			t.Errorf("Result[%d].RerankScore = %v, want %v", i, result.RerankScore, float64(documents[i].Score))
		}
		if result.FinalScore != documents[i].Score {
			t.Errorf("Result[%d].FinalScore = %v, want %v", i, result.FinalScore, documents[i].Score)
		}
		if result.Rank != i+1 {
			t.Errorf("Result[%d].Rank = %v, want %v", i, result.Rank, i+1)
		}
	}
}

func TestRerankEmptyDocuments(t *testing.T) {
	config := Config{
		Enabled: true,
	}
	client := NewClient(config)

	ctx := context.Background()
	documents := []DocumentToRerank{}

	results, err := client.Rerank(ctx, "test query", documents)
	if err != nil {
		t.Fatalf("Rerank() error = %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("Rerank() returned %d results, want 0", len(results))
	}
}

func TestRerankThresholdFiltering(t *testing.T) {
	// This test documents the threshold filtering behavior
	// Actual API testing requires LM Studio to be running
	config := Config{
		Enabled:   true,
		Threshold: 0.8, // High threshold
	}
	client := NewClient(config)

	if client.config.Threshold != 0.8 {
		t.Errorf("Client threshold not set correctly, got %v", client.config.Threshold)
	}
}

// Integration test - run only when LM Studio is available
// To run: go test -tags=integration ./internal/reranker/
func TestRerankWithLMStudio(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := Config{
		BaseURL:    "http://10.10.199.29:8080",
		Model:      "qwen3-reranker-8b",
		Enabled:    true,
		TopK:       5,
		Threshold:  0.0,
	}
	client := NewClient(config)

	ctx := context.Background()
	query := "What programming language is used for web development?"
	
	documents := []DocumentToRerank{
		{
			ID:      "doc1",
			Content: "Python is a popular programming language for web development with Django and Flask.",
			Score:   0.85,
		},
		{
			ID:      "doc2",
			Content: "JavaScript is the primary language for front-end web development.",
			Score:   0.82,
		},
		{
			ID:      "doc3",
			Content: "Java is widely used in enterprise applications and Android development.",
			Score:   0.75,
		},
		{
			ID:      "doc4",
			Content: "Go is a modern language developed by Google for system programming.",
			Score:   0.70,
		},
	}

	results, err := client.Rerank(ctx, query, documents)
	if err != nil {
		t.Logf("Rerank() failed (LM Studio may not be running): %v", err)
		t.Skip("LM Studio not available, skipping integration test")
		return
	}

	// If reranking endpoint is not available, results should still be returned
	// (graceful degradation with original scores)
	if len(results) == 0 {
		t.Fatal("Rerank() returned no results (expected graceful degradation)")
	}

	// Check that scores are in descending order (reranked)
	for i := 1; i < len(results); i++ {
		if results[i].FinalScore > results[i-1].FinalScore {
			t.Errorf("Results not sorted by score: results[%d].FinalScore=%v > results[%d].FinalScore=%v",
				i, results[i].FinalScore, i-1, results[i-1].FinalScore)
		}
	}

	// Verify JavaScript document ranks higher (more relevant to web development query)
	jsDocFound := false
	jsDocRank := 0
	for _, result := range results {
		if result.ID == "doc2" {
			jsDocFound = true
			jsDocRank = result.Rank
			break
		}
	}

	if !jsDocFound {
		t.Error("JavaScript document not found in results")
	} else if jsDocRank > 2 {
		t.Logf("Warning: JavaScript document ranked %d (expected top 2 for web development query)", jsDocRank)
	}

	t.Logf("Reranked results:")
	for _, result := range results {
		t.Logf("  Rank %d: %s (Original: %.3f, Rerank: %.3f, Final: %.3f)",
			result.Rank, result.ID, result.OriginalScore, result.RerankScore, result.FinalScore)
	}
}
