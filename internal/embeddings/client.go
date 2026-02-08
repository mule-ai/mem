package embeddings

import (
	"context"
	"fmt"
	"log"

	"github.com/jbutlerdev/mem/pkg/api"
)

const (
	defaultEndpoint = "http://10.10.199.29:8080"
	defaultModel    = "text-embedding-qwen3-embedding-8b"
	defaultDimensions = 1024
)

// Config holds configuration for the embeddings client
type Config struct {
	BaseURL    string
	Model      string
	APIKey     string
	Dimensions int
	BatchSize  int
}

// Client is an embeddings API client
type Client struct {
	apiClient *api.Client
	config    Config
}

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// EmbeddingResponse represents the response from the embeddings API
type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewClient creates a new embeddings client
func NewClient(config Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = defaultEndpoint
	}
	if config.Model == "" {
		config.Model = defaultModel
	}
	if config.Dimensions == 0 {
		config.Dimensions = defaultDimensions
	}
	if config.BatchSize == 0 {
		config.BatchSize = 10
	}

	apiClient := api.NewClient(api.ClientConfig{
		BaseURL: config.BaseURL,
		APIKey:  config.APIKey,
	})

	return &Client{
		apiClient: apiClient,
		config:    config,
	}
}

// Embed generates embeddings for a single text
func (c *Client) Embed(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := c.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return embeddings[0], nil
}

// EmbedBatch generates embeddings for multiple texts
func (c *Client) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// Process in batches if needed
	var allEmbeddings [][]float32
	
	for i := 0; i < len(texts); i += c.config.BatchSize {
		end := i + c.config.BatchSize
		if end > len(texts) {
			end = len(texts)
		}
		
		batch := texts[i:end]
		
		req := EmbeddingRequest{
			Input: batch,
			Model: c.config.Model,
		}
		
		var resp EmbeddingResponse
		err := c.apiClient.Post(ctx, "/v1/embeddings", req, &resp)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embeddings for batch %d-%d: %w", i, end-1, err)
		}
		
		if len(resp.Data) != len(batch) {
			return nil, fmt.Errorf("expected %d embeddings, got %d", len(batch), len(resp.Data))
		}
		
		// Sort by index to ensure correct order
		embeddings := make([][]float32, len(resp.Data))
		for _, item := range resp.Data {
			if item.Index < 0 || item.Index >= len(embeddings) {
				return nil, fmt.Errorf("invalid embedding index: %d", item.Index)
			}
			embeddings[item.Index] = item.Embedding
		}
		
		// Validate dimensions
		for _, emb := range embeddings {
			if len(emb) != c.config.Dimensions {
				log.Printf("Warning: embedding dimension mismatch (expected %d, got %d)", c.config.Dimensions, len(emb))
			}
		}
		
		allEmbeddings = append(allEmbeddings, embeddings...)
	}
	
	return allEmbeddings, nil
}
