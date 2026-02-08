package reranker

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// getTestClient returns a client configured for testing
func getTestClient() *Client {
	baseURL := os.Getenv("MEM_BASE_URL")
	if baseURL == "" {
		baseURL = "http://10.10.199.29:8080"
	}

	model := os.Getenv("MEM_RERANKER_MODEL")
	if model == "" {
		model = "qwen3-reranker-8b"
	}

	return NewClient(Config{
		BaseURL:   baseURL,
		Model:     model,
		APIKey:    os.Getenv("MEM_API_KEY"),
		Enabled:   true,
		TopK:      10,
		Threshold: 0.0,
	})
}

// generateTestResults creates test search results
func generateTestResults(count int) []DocumentToRerank {
	results := make([]DocumentToRerank, count)
	for i := 0; i < count; i++ {
		results[i] = DocumentToRerank{
			ID:      fmt.Sprintf("mem-%d", i),
			Content: fmt.Sprintf("Test memory content number %d for reranking benchmark", i),
			Score:   float32(1.0 - float64(i)*0.05), // Decreasing scores
		}
	}
	return results
}

// generateContent creates content of approximately the specified length
func generateContent(length int) string {
	words := []string{
		"test", "content", "memory", "search", "result", "reranking", "benchmark",
		"performance", "latency", "throughput", "semantic", "similarity", "score",
		"relevance", "ranking", "query", "document", "retrieval", "evaluation",
	}

	result := ""
	currentLen := 0
	for currentLen < length {
		for _, word := range words {
			if currentLen >= length {
				break
			}
			result += word + " "
			currentLen += len(word) + 1
		}
	}

	return result[:length]
}

// BenchmarkRerankSmall benchmarks reranking with small result sets
func BenchmarkRerankSmall(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()
	query := "test memory query"

	sizes := []int{5, 10}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Results_%d", size), func(b *testing.B) {
			results := generateTestResults(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.Rerank(ctx, query, results)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
			}
		})
	}
}

// BenchmarkRerankMedium benchmarks reranking with medium result sets
func BenchmarkRerankMedium(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()
	query := "search query for medium sized result set"

	sizes := []int{20, 50}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Results_%d", size), func(b *testing.B) {
			results := generateTestResults(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.Rerank(ctx, query, results)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
			}
		})
	}
}

// BenchmarkRerankLarge benchmarks reranking with large result sets
func BenchmarkRerankLarge(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()
	query := "search query for large result set reranking performance"

	sizes := []int{100, 200}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Results_%d", size), func(b *testing.B) {
			results := generateTestResults(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.Rerank(ctx, query, results)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
			}
		})
	}
}

// BenchmarkRerankQueryLength benchmarks reranking with different query lengths
func BenchmarkRerankQueryLength(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	queries := map[string]string{
		"Short":   "test query",
		"Medium":  "This is a medium length query for testing reranker performance",
		"Long":    "This is a much longer query that contains more detail and context for the reranker to process when evaluating the relevance of search results",
		"Complex": "User preferences: dark mode enabled, vim keybindings active, font size 14px. Search for related configuration settings and similar user preferences",
	}

	for name, query := range queries {
		b.Run(name, func(b *testing.B) {
			results := generateTestResults(20)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.Rerank(ctx, query, results)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
			}
		})
	}
}

// BenchmarkRerankContentLength benchmarks reranking with different content lengths
func BenchmarkRerankContentLength(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()
	query := "test query"

	lengths := []int{50, 200, 500, 1000}

	for _, length := range lengths {
		b.Run(fmt.Sprintf("Content_%d", length), func(b *testing.B) {
			results := make([]DocumentToRerank, 20)
			for i := 0; i < 20; i++ {
				content := generateContent(length)
				results[i] = DocumentToRerank{
					ID:      fmt.Sprintf("mem-%d", i),
					Content: content,
					Score:   float32(1.0 - float64(i)*0.05),
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.Rerank(ctx, query, results)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
			}
		})
	}
}

// BenchmarkRerankRealWorld benchmarks realistic reranking scenarios
func BenchmarkRerankRealWorld(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	scenarios := []struct {
		Name     string
		Query    string
		Contents []string
	}{
		{
			Name:  "UserPreferences",
			Query: "user interface preferences",
			Contents: []string{
				"User prefers dark mode with vim keybindings",
				"Database configuration: PostgreSQL 14 with pgvector",
				"Font settings: monospace, 14pt size",
				"API endpoint configuration for production",
				"Keyboard shortcuts and editor settings",
			},
		},
		{
			Name:  "CodeConfig",
			Query: "TypeScript configuration options",
			Contents: []string{
				"Enable strict mode in tsconfig.json",
				"Set target to ES2020 for modern JavaScript",
				"Configure path aliases for imports",
				"Use MongoDB for document storage",
				"Enable decorators and experimental features",
			},
		},
		{
			Name:  "MeetingNotes",
			Query: "project deadline and milestones",
			Contents: []string{
				"Q1 roadmap review completed",
				"Project alpha deadline: March 15th",
				"Team standup at 10am daily",
				"Bug fixes priority list updated",
				"Sprint planning for next iteration",
			},
		},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.Name, func(b *testing.B) {
			results := make([]DocumentToRerank, len(scenario.Contents))
			for i, content := range scenario.Contents {
				results[i] = DocumentToRerank{
					ID:      fmt.Sprintf("mem-%d", i),
					Content: content,
					Score:   float32(1.0 - float64(i)*0.1),
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.Rerank(ctx, scenario.Query, results)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
			}
		})
	}
}

// BenchmarkRerankConcurrent benchmarks concurrent reranking requests
func BenchmarkRerankConcurrent(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	b.Run("ConcurrentRerank", func(b *testing.B) {
		results := generateTestResults(20)
		query := "test query for concurrent reranking"

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				_, err := client.Rerank(ctx, query, results)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
				i++
			}
		})
	})
}

// BenchmarkRerankMemoryAllocation benchmarks memory allocation during reranking
func BenchmarkRerankMemoryAllocation(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()
	query := "test query for memory allocation"
	results := generateTestResults(20)

	b.Run("AllocPerRerank", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := client.Rerank(ctx, query, results)
			if err != nil {
				b.Skipf("Skipping benchmark: API not available or error: %v", err)
				return
			}
		}
	})
}

// BenchmarkRerankTopK benchmarks reranking with different top_k values
func BenchmarkRerankTopK(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()
	query := "test query for top k benchmarking"

	topKValues := []int{5, 10, 20}

	for _, topK := range topKValues {
		b.Run(fmt.Sprintf("TopK_%d", topK), func(b *testing.B) {
			results := generateTestResults(50)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				reranked, err := client.Rerank(ctx, query, results)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
				// Apply top_k filter
				if len(reranked) > topK {
					reranked = reranked[:topK]
				}
			}
		})
	}
}

// BenchmarkRerankWithThreshold benchmarks reranking with threshold filtering
func BenchmarkRerankWithThreshold(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()
	query := "test query with threshold"

	thresholds := []float32{0.3, 0.5, 0.7}

	for _, threshold := range thresholds {
		b.Run(fmt.Sprintf("Threshold_%.2f", threshold), func(b *testing.B) {
			results := generateTestResults(30)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				reranked, err := client.Rerank(ctx, query, results)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
				// Apply threshold filter
				filtered := make([]RerankedDocument, 0)
				for _, r := range reranked {
					if r.FinalScore >= threshold {
						filtered = append(filtered, r)
					}
				}
				_ = filtered
			}
		})
	}
}

// BenchmarkEndToEnd benchmarks the full pipeline: query -> search -> rerank
func BenchmarkEndToEnd(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	queries := []string{
		"user preferences",
		"database configuration",
		"project deadlines",
	}

	for _, query := range queries {
		b.Run(query, func(b *testing.B) {
			// Simulate vector search results
			results := generateTestResults(20)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Rerank the results
				reranked, err := client.Rerank(ctx, query, results)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
				// Take top 5
				if len(reranked) > 5 {
					reranked = reranked[:5]
				}
				_ = reranked
			}
		})
	}
}
