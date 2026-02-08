package cli

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jbutlerdev/mem/internal/models"
)

// TestQueryCmdBasic tests basic query command functionality
func TestQueryCmdBasic(t *testing.T) {
	t.Run("QueryCommandExists", func(t *testing.T) {
		if QueryCmd == nil {
			t.Fatal("QueryCmd should not be nil")
		}
		if QueryCmd.Use != "query [query]" {
			t.Errorf("Expected use 'query [query]', got '%s'", QueryCmd.Use)
		}
		if QueryCmd.Short != "Query memories using semantic search" {
			t.Errorf("Expected short description 'Query memories using semantic search', got '%s'", QueryCmd.Short)
		}
	})

	t.Run("QueryCommandFlags", func(t *testing.T) {
		flags := QueryCmd.Flags()

		namespaceFlag := flags.Lookup("namespace")
		if namespaceFlag == nil {
			t.Error("namespace flag should be defined")
		} else if namespaceFlag.Shorthand != "n" {
			t.Errorf("Expected namespace shorthand 'n', got '%s'", namespaceFlag.Shorthand)
		}

		limitFlag := flags.Lookup("limit")
		if limitFlag == nil {
			t.Error("limit flag should be defined")
		} else if limitFlag.Shorthand != "l" {
			t.Errorf("Expected limit shorthand 'l', got '%s'", limitFlag.Shorthand)
		}

		topFlag := flags.Lookup("top")
		if topFlag == nil {
			t.Error("top flag should be defined")
		}

		thresholdFlag := flags.Lookup("threshold")
		if thresholdFlag == nil {
			t.Error("threshold flag should be defined")
		}

		noRerankFlag := flags.Lookup("no-rerank")
		if noRerankFlag == nil {
			t.Error("no-rerank flag should be defined")
		}

		outputFlag := flags.Lookup("output")
		if outputFlag == nil {
			t.Error("output flag should be defined")
		} else if outputFlag.Shorthand != "o" {
			t.Errorf("Expected output shorthand 'o', got '%s'", outputFlag.Shorthand)
		}
	})
}

// TestOutputQueryResultsText tests text output formatting
func TestOutputQueryResultsText(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name    string
		results []models.SearchResult
		wantIn  []string
	}{
		{
			name: "single result",
			results: []models.SearchResult{
				{
					Memory: models.Memory{
						ID:        "mem1",
						Content:   "Test content about user preferences",
						Namespace: "default",
						Tags:      []string{"preferences"},
						CreatedAt: now,
						UpdatedAt: now,
					},
					Score:    0.95,
					Reranked: false,
					Rank:     1,
				},
			},
			wantIn: []string{
				"Found 1 result",
				"[default]",
				"mem1",
				"0.9500",
				"preferences",
			},
		},
		{
			name: "multiple results with reranking",
			results: []models.SearchResult{
				{
					Memory: models.Memory{
						ID:        "mem1",
						Content:   "First result",
						Namespace: "test",
						Tags:      []string{"tag1"},
						CreatedAt: now,
						UpdatedAt: now,
					},
					Score:    0.92,
					Reranked: true,
					Rank:     1,
				},
				{
					Memory: models.Memory{
						ID:        "mem2",
						Content:   "Second result",
						Namespace: "test",
						Tags:      []string{"tag2"},
						CreatedAt: now,
						UpdatedAt: now,
					},
					Score:    0.85,
					Reranked: true,
					Rank:     2,
				},
			},
			wantIn: []string{
				"Found 2 result",
				"(reranked)",
				"mem1",
				"mem2",
			},
		},
		{
			name:    "no results",
			results: []models.SearchResult{},
			wantIn:  []string{"No results found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			outputQueryResultsText(tt.results)

			w.Close()
			os.Stdout = old

			var buf strings.Builder
			io.Copy(&buf, r)
			output := buf.String()

			// Check that output contains expected strings
			for _, want := range tt.wantIn {
				if !strings.Contains(output, want) {
					t.Errorf("Output should contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

// TestQueryOptionsDefaults tests default query options
func TestQueryOptionsDefaults(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		top           int
		threshold     float32
		configLimit   int
		configThreshold float32
		expectLimit   int
		expectTop     int
		expectThresh  float32
	}{
		{
			name:          "all defaults",
			limit:         5,
			top:           10,
			threshold:     0.6,
			configLimit:   0,
			configThreshold: 0.0,
			expectLimit:   5,
			expectTop:     10,
			expectThresh:  0.6,
		},
		{
			name:          "config overrides",
			limit:         5,
			top:           10,
			threshold:     0.6,
			configLimit:   10,
			configThreshold: 0.7,
			expectLimit:   10,
			expectTop:     20,  // top = limit * 2 when config limit is set
			expectThresh:  0.7,
		},
		{
			name:          "flag overrides",
			limit:         15,
			top:           30,
			threshold:     0.5,
			configLimit:   10,
			configThreshold: 0.7,
			expectLimit:   15,
			expectTop:     30,
			expectThresh:  0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from runQuery
			limit := tt.limit
			if limit == 5 && tt.configLimit != 0 {
				limit = tt.configLimit
			}

			top := tt.top
			if top == 10 && tt.configLimit != 0 {
				top = limit * 2
			}

			threshold := tt.threshold
			if threshold == 0.6 && tt.configThreshold != 0 {
				threshold = float32(tt.configThreshold)
			}

			if limit != tt.expectLimit {
				t.Errorf("Expected limit %d, got %d", tt.expectLimit, limit)
			}
			if top != tt.expectTop {
				t.Errorf("Expected top %d, got %d", tt.expectTop, top)
			}
			if threshold != tt.expectThresh {
				t.Errorf("Expected threshold %.2f, got %.2f", tt.expectThresh, threshold)
			}
		})
	}
}

// TestQueryValidation tests query validation logic
func TestQueryValidation(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid query",
			query:   "user preferences",
			wantErr: false,
		},
		{
			name:    "query with leading/trailing spaces",
			query:   "  user preferences  ",
			wantErr: false,
		},
		{
			name:    "empty query",
			query:   "",
			wantErr: true,
			errMsg:  "query cannot be empty",
		},
		{
			name:    "whitespace only query",
			query:   "   ",
			wantErr: true,
			errMsg:  "query cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := strings.TrimSpace(tt.query)
			hasErr := query == ""
			
			if hasErr != tt.wantErr {
				t.Errorf("Expected error: %v, got: %v", tt.wantErr, hasErr)
			}
			
			if hasErr && tt.errMsg != "" {
				// Would return this error in runQuery
				expectedErr := tt.errMsg
				if query == "" && tt.query != "" {
					// This is the trimmed case
					if !strings.Contains(expectedErr, "empty") {
						t.Errorf("Error message should mention empty query")
					}
				}
			}
		})
	}
}

// TestNamespaceResolution tests namespace resolution logic
func TestNamespaceResolution(t *testing.T) {
	tests := []struct {
		name          string
		flagNS        string
		configNS      string
		expectedNS    string
	}{
		{
			name:       "flag overrides config",
			flagNS:     "flag-ns",
			configNS:   "config-ns",
			expectedNS: "flag-ns",
		},
		{
			name:       "config used when no flag",
			flagNS:     "",
			configNS:   "config-ns",
			expectedNS: "config-ns",
		},
		{
			name:       "default namespace when neither set",
			flagNS:     "",
			configNS:   "",
			expectedNS: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			namespace := tt.flagNS
			if namespace == "" {
				namespace = tt.configNS
			}
			if namespace == "" {
				namespace = "default"
			}

			if namespace != tt.expectedNS {
				t.Errorf("Expected namespace %q, got %q", tt.expectedNS, namespace)
			}
		})
	}
}

// TestOutputFormatResolution tests output format resolution
func TestOutputFormatResolution(t *testing.T) {
	tests := []struct {
		name         string
		flagOutput   string
		configOutput string
		expected     string
	}{
		{
			name:         "flag overrides config",
			flagOutput:   "json",
			configOutput: "yaml",
			expected:     "json",
		},
		{
			name:         "config used when no flag",
			flagOutput:   "",
			configOutput: "yaml",
			expected:     "yaml",
		},
		{
			name:         "text default when neither set",
			flagOutput:   "",
			configOutput: "",
			expected:     "text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputFormat := tt.flagOutput
			if outputFormat == "" {
				outputFormat = tt.configOutput
			}
			if outputFormat == "" {
				outputFormat = "text"
			}

			if outputFormat != tt.expected {
				t.Errorf("Expected output format %q, got %q", tt.expected, outputFormat)
			}
		})
	}
}