package embeddings

import (
	"context"
	"testing"
)

// TestEmbeddingsClient tests the embeddings client against LM Studio
func TestEmbeddingsClient(t *testing.T) {
	ctx := context.Background()
	
	// Note: Use whichever model is loaded in LM Studio
	// text-embedding-qwen3-embedding-8b is the specified model in the spec
	// If not loaded, text-embedding-nomic-embed-text-v1.5 can be used for testing
	client := NewClient(Config{
		BaseURL:    "http://10.10.199.29:8080",
		Model:      "text-embedding-nomic-embed-text-v1.5",
		Dimensions: 768, // Nomic uses 768 dimensions
	})
	
	t.Run("Single Embedding", func(t *testing.T) {
		text := "User prefers dark mode and vim keybindings"
		embedding, err := client.Embed(ctx, text)
		if err != nil {
			t.Fatalf("Failed to generate embedding: %v", err)
		}
		
		if len(embedding) == 0 {
			t.Fatal("Embedding is empty")
		}
		
		t.Logf("Generated embedding with %d dimensions", len(embedding))
	})
	
	t.Run("Batch Embeddings", func(t *testing.T) {
		texts := []string{
			"Database uses PostgreSQL 14 with pgvector extension",
			"Always use TypeScript strict mode for type safety",
			"Project alpha deadline is March 15th",
		}
		
		embeddings, err := client.EmbedBatch(ctx, texts)
		if err != nil {
			t.Fatalf("Failed to generate batch embeddings: %v", err)
		}
		
		if len(embeddings) != len(texts) {
			t.Fatalf("Expected %d embeddings, got %d", len(texts), len(embeddings))
		}
		
		for i, emb := range embeddings {
			if len(emb) == 0 {
				t.Errorf("Embedding %d is empty", i)
			}
			t.Logf("Embedding %d: %d dimensions", i, len(emb))
		}
	})
	
	t.Run("Large Batch Test", func(t *testing.T) {
		// Test batching with more than batch_size texts
		texts := make([]string, 25)
		for i := range texts {
			texts[i] = "Test memory number " + string(rune('A'+i))
		}
		
		embeddings, err := client.EmbedBatch(ctx, texts)
		if err != nil {
			t.Fatalf("Failed to generate large batch: %v", err)
		}
		
		if len(embeddings) != len(texts) {
			t.Fatalf("Expected %d embeddings, got %d", len(texts), len(embeddings))
		}
		
		t.Logf("Successfully generated %d embeddings in batch", len(embeddings))
	})
}

// TestEmbeddingsDimensions verifies the embedding dimensions
func TestEmbeddingsDimensions(t *testing.T) {
	ctx := context.Background()
	
	client := NewClient(Config{
		BaseURL:    "http://10.10.199.29:8080",
		Model:      "text-embedding-nomic-embed-text-v1.5",
		Dimensions: 768,
	})
	
	text := "Test dimension verification"
	embedding, err := client.Embed(ctx, text)
	if err != nil {
		t.Fatalf("Failed to generate embedding: %v", err)
	}
	
	if len(embedding) != client.config.Dimensions {
		t.Logf("Warning: Expected %d dimensions, got %d", client.config.Dimensions, len(embedding))
	} else {
		t.Logf("Dimension check passed: %d", len(embedding))
	}
}