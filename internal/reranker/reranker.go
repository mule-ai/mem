package reranker

import (
	"context"
	"log"

	"github.com/jbutlerdev/mem/pkg/api"
)

const (
	defaultEndpoint = "http://10.10.199.29:8080"
	defaultModel    = "qwen3-reranker-8b"
)

// Config holds configuration for the reranker client
type Config struct {
	BaseURL    string
	Model      string
	APIKey     string
	Enabled    bool
	TopK       int
	Threshold  float64
}

// Client is a reranker API client
type Client struct {
	apiClient *api.Client
	config    Config
}

// RerankRequest represents a request to rerank results
type RerankRequest struct {
	Model    string       `json:"model"`
	Query    string       `json:"query"`
	Documents []Document  `json:"documents"`
	TopN     int          `json:"top_n,omitempty"`
}

// Document represents a document to be reranked
type Document struct {
	ID      string `json:"id"`
	Content string `json:"text"`
}

// RerankResponse represents the response from the reranker API
type RerankResponse struct {
	Object string `json:"object"`
	Model  string `json:"model"`
	Results []Result `json:"results"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// Result represents a single reranked result
type Result struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
	Document       Document `json:"document"`
}

// DocumentToRerank represents a document with its metadata before reranking
type DocumentToRerank struct {
	ID      string
	Content string
	Score   float32 // Original similarity score
}

// RerankedDocument represents a document after reranking
type RerankedDocument struct {
	ID              string
	Content         string
	OriginalScore   float32
	RerankScore     float64
	FinalScore      float32
	Rank            int
}

// NewClient creates a new reranker client
func NewClient(config Config) *Client {
	if config.BaseURL == "" {
		config.BaseURL = defaultEndpoint
	}
	if config.Model == "" {
		config.Model = defaultModel
	}
	if config.TopK == 0 {
		config.TopK = 10
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

// Rerank reranks the given documents based on the query
func (c *Client) Rerank(ctx context.Context, query string, documents []DocumentToRerank) ([]RerankedDocument, error) {
	if !c.config.Enabled {
		// Reranking disabled, return original scores
		return c.noopRerank(documents), nil
	}

	if len(documents) == 0 {
		return []RerankedDocument{}, nil
	}

	// Prepare request
	docs := make([]Document, len(documents))
	for i, doc := range documents {
		docs[i] = Document{
			ID:      doc.ID,
			Content: doc.Content,
		}
	}

	req := RerankRequest{
		Model:     c.config.Model,
		Query:     query,
		Documents: docs,
		TopN:      len(docs), // Request all results back
	}

	var resp RerankResponse
	err := c.apiClient.Post(ctx, "/v1/rerank", req, &resp)
	if err != nil {
		// Graceful degradation: return original scores if reranking fails
		// This handles cases where:
		// - Reranking endpoint is not available (e.g., LM Studio without rerank support)
		// - API is temporarily unreachable
		// - Model is not loaded
		log.Printf("Warning: reranking API call failed (endpoint may not be supported): %v", err)
		return c.noopRerank(documents), nil
	}

	// Some APIs (like LM Studio) return HTTP 200 with an error in the body
	// Check if the response indicates an error (empty results with no success indication)
	if len(resp.Results) == 0 && resp.Object == "" {
		log.Printf("Warning: reranking API returned empty response (endpoint may not be supported)")
		return c.noopRerank(documents), nil
	}

	// Build reranked results
	resultMap := make(map[string]DocumentToRerank)
	for _, doc := range documents {
		resultMap[doc.ID] = doc
	}

	reranked := make([]RerankedDocument, 0, len(resp.Results))
	for rank, result := range resp.Results {
		originalDoc, exists := resultMap[result.Document.ID]
		if !exists {
			log.Printf("Warning: reranker returned unknown document ID: %s", result.Document.ID)
			continue
		}

		// Filter by threshold if set
		if c.config.Threshold > 0 && result.RelevanceScore < c.config.Threshold {
			continue
		}

		// Calculate final score (blend of original and rerank scores)
		// Using a weighted average: 30% original, 70% rerank
		finalScore := float32(originalDoc.Score)*0.3 + float32(result.RelevanceScore)*0.7

		reranked = append(reranked, RerankedDocument{
			ID:            result.Document.ID,
			Content:       result.Document.Content,
			OriginalScore: originalDoc.Score,
			RerankScore:   result.RelevanceScore,
			FinalScore:    finalScore,
			Rank:          rank + 1,
		})
	}

	return reranked, nil
}

// noopRerank returns documents without reranking (when disabled or on error)
func (c *Client) noopRerank(documents []DocumentToRerank) []RerankedDocument {
	reranked := make([]RerankedDocument, len(documents))
	for i, doc := range documents {
		reranked[i] = RerankedDocument{
			ID:            doc.ID,
			Content:       doc.Content,
			OriginalScore: doc.Score,
			RerankScore:   float64(doc.Score),
			FinalScore:    doc.Score,
			Rank:          i + 1,
		}
	}
	return reranked
}

// IsEnabled returns whether reranking is enabled
func (c *Client) IsEnabled() bool {
	return c.config.Enabled
}

// GetTopK returns the configured top K value for initial retrieval
func (c *Client) GetTopK() int {
	return c.config.TopK
}