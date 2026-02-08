package embeddings

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

	return NewClient(Config{
		BaseURL:    baseURL,
		Model:      defaultModel,
		Dimensions: defaultDimensions,
		BatchSize:  10,
	})
}

// BenchmarkEmbedSingle benchmarks embedding a single text
func BenchmarkEmbedSingle(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	testTexts := []string{
		"Short text",
		"This is a medium length text for embedding generation benchmark testing",
		"This is a much longer text that contains more information and should take more time to process through the embedding model. It includes multiple sentences and various details to test performance.",
	}

	for _, text := range testTexts {
		b.Run(fmt.Sprintf("Length_%d", len(text)), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.Embed(ctx, text)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
			}
		})
	}
}

// BenchmarkEmbedBatch benchmarks embedding multiple texts in a single batch
func BenchmarkEmbedBatch(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	batchSizes := []int{1, 5, 10, 20, 50}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(b *testing.B) {
			// Prepare batch texts
			texts := make([]string, batchSize)
			for i := 0; i < batchSize; i++ {
				texts[i] = fmt.Sprintf("This is test text number %d for batch embedding benchmarking with medium length content", i)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.EmbedBatch(ctx, texts)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
			}
		})
	}
}

// BenchmarkEmbedBatchLarge benchmarks embedding larger batches
func BenchmarkEmbedBatchLarge(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	// Test with the configured batch size
	b.Run("ConfiguredBatchSize", func(b *testing.B) {
		batchSize := client.config.BatchSize
		texts := make([]string, batchSize)
		for i := 0; i < batchSize; i++ {
			texts[i] = fmt.Sprintf("Test text %d for batch embedding with larger content size to test throughput and performance characteristics", i)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := client.EmbedBatch(ctx, texts)
			if err != nil {
				b.Skipf("Skipping benchmark: API not available or error: %v", err)
				return
			}
		}
	})
}

// BenchmarkEmbedThroughput benchmarks the throughput of embedding operations
func BenchmarkEmbedThroughput(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	b.Run("SequentialSingle", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			text := fmt.Sprintf("Throughput test text number %d", i)
			_, err := client.Embed(ctx, text)
			if err != nil {
				b.Skipf("Skipping benchmark: API not available or error: %v", err)
				return
			}
		}
	})

	b.Run("SequentialBatch", func(b *testing.B) {
		batchSize := 10
		b.ResetTimer()
		for i := 0; i < b.N/batchSize; i++ {
			texts := make([]string, batchSize)
			for j := 0; j < batchSize; j++ {
				texts[j] = fmt.Sprintf("Throughput test text %d", i*batchSize+j)
			}
			_, err := client.EmbedBatch(ctx, texts)
			if err != nil {
				b.Skipf("Skipping benchmark: API not available or error: %v", err)
				return
			}
		}
	})
}

// BenchmarkEmbedVariety benchmarks embedding texts of varying content types
func BenchmarkEmbedVariety(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	contentTypes := map[string]string{
		"Code":         `func main() { fmt.Println("Hello, World!") }`,
		"JSON":         `{"name": "test", "value": 123, "nested": {"key": "value"}}`,
		"URL":          `https://example.com/path/to/resource?param=value&other=123`,
		"Email":        `test@example.com, another@test.org, user@company.co.uk`,
		"Numbers":      `1234567890 3.14159 98.6 -42 1000000 0.001`,
		"Mixed":        `User john_doe logged in at 2024-02-07T10:30:00Z from IP 192.168.1.1. Session ID: abc123xyz456`,
	}

	for name, content := range contentTypes {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.Embed(ctx, content)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
			}
		})
	}
}

// BenchmarkEmbedLongText benchmarks embedding long texts
func BenchmarkEmbedLongText(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	// Generate texts of different lengths
	lengths := []int{100, 500, 1000, 2000, 4000}

	for _, length := range lengths {
		b.Run(fmt.Sprintf("Chars_%d", length), func(b *testing.B) {
			// Generate text of specified length
			text := generateText(length)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.Embed(ctx, text)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
			}
		})
	}
}

// generateText creates a text of approximately the specified length
func generateText(length int) string {
	words := []string{
		"the", "quick", "brown", "fox", "jumps", "over", "lazy", "dog",
		"benchmark", "test", "embedding", "vector", "semantic", "memory",
		"search", "retrieval", "performance", "latency", "throughput",
		"chromem", "storage", "database", "query", "similarity", "score",
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

// BenchmarkEmbedConcurrent benchmarks concurrent embedding requests
func BenchmarkEmbedConcurrent(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	b.Run("ConcurrentSingle", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				text := fmt.Sprintf("Concurrent test text %d", i)
				_, err := client.Embed(ctx, text)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
				i++
			}
		})
	})

	b.Run("ConcurrentBatch", func(b *testing.B) {
		batchSize := 10
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				texts := make([]string, batchSize)
				for j := 0; j < batchSize; j++ {
					texts[j] = fmt.Sprintf("Concurrent batch test %d-%d", i, j)
				}
				_, err := client.EmbedBatch(ctx, texts)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
				i++
			}
		})
	})
}

// BenchmarkEmbedRealWorld benchmarks realistic use cases
func BenchmarkEmbedRealWorld(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	scenarios := map[string]string{
		"UserPreference":    "User prefers dark mode, uses vim keybindings, likes monospace fonts, and sets font size to 14pt",
		"CodeSnippet":       "function calculateHash(data) { return crypto.createHash('sha256').update(data).digest('hex'); }",
		"MeetingNote":       "Meeting with engineering team: Discussed Q1 roadmap, decided to prioritize GraphQL API, assigned tasks to Sarah and Mike, follow-up next Tuesday at 2pm",
		"Documentation":     "The API endpoint /api/v1/memories accepts POST requests with a JSON body containing content, namespace, tags, and optional metadata fields. Returns the created memory object with generated ID and embedding",
		"ErrorLog":          "ERROR: Failed to connect to database at localhost:5432/mem: connection refused. Retrying in 5 seconds... Attempt 3/10",
		"ConfigSnippet":     "memory:\n  path: ~/.mem/data\n  default_namespace: default\nembeddings:\n  base_url: http://10.10.199.29:8080\n  model: text-embedding-qwen3-embedding-8b",
	}

	for name, text := range scenarios {
		b.Run(name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := client.Embed(ctx, text)
				if err != nil {
					b.Skipf("Skipping benchmark: API not available or error: %v", err)
					return
				}
			}
		})
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation during embedding
func BenchmarkMemoryAllocation(b *testing.B) {
	client := getTestClient()
	ctx := context.Background()

	b.Run("AllocPerEmbed", func(b *testing.B) {
		text := "Test text for memory allocation benchmarking"
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := client.Embed(ctx, text)
			if err != nil {
				b.Skipf("Skipping benchmark: API not available or error: %v", err)
				return
			}
		}
	})
}
