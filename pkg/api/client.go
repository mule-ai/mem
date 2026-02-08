package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultMaxRetries = 3
	defaultRetryDelay = 500 * time.Millisecond
)

// Client is an HTTP client for making API requests with retry logic
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	maxRetries int
	retryDelay time.Duration
}

// ClientConfig holds configuration for the API client
type ClientConfig struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

// NewClient creates a new API client with the given configuration
func NewClient(config ClientConfig) *Client {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = defaultMaxRetries
	}

	retryDelay := config.RetryDelay
	if retryDelay == 0 {
		retryDelay = defaultRetryDelay
	}

	return &Client{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

// Do performs an HTTP request with retry logic
func (c *Client) Do(ctx context.Context, method, path string, body, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(c.retryDelay * time.Duration(attempt)):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		url := c.baseURL + path
		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")
		if c.apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+c.apiKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			continue
		}

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Check status code
		if resp.StatusCode >= 400 {
			var errResp struct {
				Error struct {
					Message string `json:"message"`
					Type    string `json:"type"`
					Code    string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(respBody, &errResp); err == nil && errResp.Error.Message != "" {
				lastErr = fmt.Errorf("API error (status %d): %s", resp.StatusCode, errResp.Error.Message)
			} else {
				lastErr = fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
			}
			
			// Don't retry client errors (4xx)
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return lastErr
			}
			continue
		}

		// Decode successful response
		if result != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, result); err != nil {
				return fmt.Errorf("failed to decode response: %w", err)
			}
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.Do(ctx, http.MethodGet, path, nil, result)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body, result interface{}) error {
	return c.Do(ctx, http.MethodPost, path, body, result)
}
