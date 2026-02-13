package reranker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

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

// ChatMessage represents a message in a chat completion request
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents a request to the chat/completions endpoint
type ChatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// ChatCompletionResponse represents the response from the chat/completions API
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Document represents a document to be reranked
type Document struct {
	ID      string `json:"id"`
	Content string `json:"text"`
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

// Rerank reranks the given documents based on the query using chat/completions
func (c *Client) Rerank(ctx context.Context, query string, documents []DocumentToRerank) ([]RerankedDocument, error) {
	if !c.config.Enabled {
		// Reranking disabled, return original scores
		return c.noopRerank(documents), nil
	}

	if len(documents) == 0 {
		return []RerankedDocument{}, nil
	}

	// Build the reranking prompt
	// The model should output a JSON array of objects with id and score
	prompt := c.buildRerankPrompt(query, documents)

	req := ChatCompletionRequest{
		Model: c.config.Model,
		Messages: []ChatMessage{
			{Role: "user", Content: prompt},
		},
	}

	var resp ChatCompletionResponse
	err := c.apiClient.Post(ctx, "/v1/chat/completions", req, &resp)
	if err != nil {
		log.Printf("Warning: reranking API call failed: %v", err)
		return c.noopRerank(documents), nil
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		log.Printf("Warning: reranking API returned empty response")
		return c.noopRerank(documents), nil
	}

	// Parse the model's response to extract rankings
	reranked, err := c.parseRerankResponse(resp.Choices[0].Message.Content, documents)
	if err != nil {
		log.Printf("Warning: failed to parse rerank response: %v", err)
		return c.noopRerank(documents), nil
	}

	return reranked, nil
}

// buildRerankPrompt creates a prompt for the reranker model
func (c *Client) buildRerankPrompt(query string, documents []DocumentToRerank) string {
	var sb strings.Builder
	sb.WriteString("Given the following query and documents, rate each document's relevance to the query on a scale of 0-1.\n")
	sb.WriteString("Output ONLY a valid JSON array with objects containing 'id' and 'score' fields.\n\n")
	sb.WriteString("Query: " + query + "\n\n")
	sb.WriteString("Documents:\n")
	for i, doc := range documents {
		sb.WriteString(fmt.Sprintf("%d. [id: %s] %s\n", i+1, doc.ID, doc.Content))
	}
	sb.WriteString("\nRespond with a JSON array like: [{\"id\": \"doc1\", \"score\": 0.95}, {\"id\": \"doc2\", \"score\": 0.80}, ...]\n")
	sb.WriteString("Only include documents that are at least somewhat relevant (score > 0.1).")
	return sb.String()
}

// parseRerankResponse parses the model's response to extract reranking scores
func (c *Client) parseRerankResponse(response string, documents []DocumentToRerank) ([]RerankedDocument, error) {
	// Clean up the response
	response = strings.TrimSpace(response)

	// Remove markdown code blocks if present
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")

	// Try to find JSON array in response more carefully
	// Look for the first '[' and last ']'
	startIdx := -1
	endIdx := -1
	
	for i, ch := range response {
		if ch == '[' {
			startIdx = i
			break
		}
	}
	
	for i := len(response) - 1; i >= 0; i-- {
		if response[i] == ']' {
			endIdx = i
			break
		}
	}
	
	if startIdx == -1 || endIdx == -1 || startIdx >= endIdx {
		// Last resort: try to find any {...} or [...] pattern
		return nil, fmt.Errorf("no JSON array found in response: %s", response[:min(len(response), 100)])
	}

	jsonStr := response[startIdx : endIdx+1]

	// Parse the JSON response
	var scores []struct {
		ID    string  `json:"id"`
		Score float64 `json:"score"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &scores); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Build a map of original documents
	docMap := make(map[string]DocumentToRerank)
	for _, doc := range documents {
		docMap[doc.ID] = doc
	}

	// Build reranked results
	scoreMap := make(map[string]float64)
	for _, s := range scores {
		scoreMap[s.ID] = s.Score
	}

	reranked := make([]RerankedDocument, 0, len(scores))

	// Process scores in order
	for _, s := range scores {
		doc, exists := docMap[s.ID]
		if !exists {
			log.Printf("Warning: reranker returned unknown document ID: %s", s.ID)
			continue
		}

		// Filter by threshold if set
		if c.config.Threshold > 0 && s.Score < c.config.Threshold {
			continue
		}

		// Calculate final score (blend of original and rerank scores)
		// Using a weighted average: 30% original, 70% rerank
		finalScore := float32(doc.Score)*0.3 + float32(s.Score)*0.7

		reranked = append(reranked, RerankedDocument{
			ID:            doc.ID,
			Content:       doc.Content,
			OriginalScore: doc.Score,
			RerankScore:   s.Score,
			FinalScore:    finalScore,
			Rank:          len(reranked) + 1,
		})
	}

	// Sort by rerank score descending
	for i := 0; i < len(reranked)-1; i++ {
		for j := i + 1; j < len(reranked); j++ {
			if reranked[j].RerankScore > reranked[i].RerankScore {
				reranked[i], reranked[j] = reranked[j], reranked[i]
			}
		}
	}

	// Update ranks after sorting
	for i := range reranked {
		reranked[i].Rank = i + 1
	}

	// If no valid scores, fall back to noop
	if len(reranked) == 0 {
		return c.noopRerank(documents), nil
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