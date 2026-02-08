package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jbutlerdev/mem/internal/models"
)

// cleanupTestDB removes the test database directory
func cleanupTestDB(dataPath string) {
	os.RemoveAll(dataPath)
}

// BenchmarkStoreSingle benchmarks storing a single memory
func BenchmarkStoreSingle(b *testing.B) {
	dataPath := filepath.Join(os.TempDir(), fmt.Sprintf("mem-bench-store-%d", time.Now().UnixNano()))
	defer cleanupTestDB(dataPath)

	config := &ChromemConfig{
		DataPath:        dataPath,
		CollectionName:  "bench_memories",
		DefaultNamespace: "default",
		EmbeddingDim:    1024,
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memory := &models.Memory{
			ID:        fmt.Sprintf("bench-%d", i),
			Content:   "This is a test memory for benchmarking store performance",
			Embedding: make([]float32, 1024),
			Namespace: "default",
			Tags:      []string{"benchmark"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := storage.Store(memory); err != nil {
			b.Fatalf("Failed to store memory: %v", err)
		}
	}
}

// BenchmarkStoreBatch benchmarks storing multiple memories in sequence
func BenchmarkStoreBatch(b *testing.B) {
	dataPath := filepath.Join(os.TempDir(), fmt.Sprintf("mem-bench-batch-%d", time.Now().UnixNano()))
	defer cleanupTestDB(dataPath)

	config := &ChromemConfig{
		DataPath:        dataPath,
		CollectionName:  "bench_memories",
		DefaultNamespace: "default",
		EmbeddingDim:    1024,
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	batchSizes := []int{10, 50, 100}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < batchSize; j++ {
					memory := &models.Memory{
						ID:        fmt.Sprintf("bench-%d-%d", i, j),
						Content:   fmt.Sprintf("Test memory %d in batch %d for benchmarking", j, i),
						Embedding: make([]float32, 1024),
						Namespace: "default",
						Tags:      []string{"benchmark", fmt.Sprintf("batch-%d", batchSize)},
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}
					if err := storage.Store(memory); err != nil {
						b.Fatalf("Failed to store memory: %v", err)
					}
				}
			}
		})
	}
}

// BenchmarkQuery benchmarks querying memories
func BenchmarkQuery(b *testing.B) {
	dataPath := filepath.Join(os.TempDir(), fmt.Sprintf("mem-bench-query-%d", time.Now().UnixNano()))
	defer cleanupTestDB(dataPath)

	config := &ChromemConfig{
		DataPath:        dataPath,
		CollectionName:  "bench_memories",
		DefaultNamespace: "default",
		EmbeddingDim:    1024,
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Pre-populate with memories
	numMemories := 1000
	for i := 0; i < numMemories; i++ {
		memory := &models.Memory{
			ID:        fmt.Sprintf("memory-%d", i),
			Content:   fmt.Sprintf("This is test memory number %d with some content for querying", i),
			Embedding: make([]float32, 1024),
			Namespace: "default",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		// Initialize embedding with some values
		for j := range memory.Embedding {
			memory.Embedding[j] = float32(i%100) / 100.0
		}
		if err := storage.Store(memory); err != nil {
			b.Fatalf("Failed to store memory: %v", err)
		}
	}

	// Benchmark query with different limits
	limits := []int{5, 10, 20, 50}
	for _, limit := range limits {
		b.Run(fmt.Sprintf("Limit_%d", limit), func(b *testing.B) {
			queryEmbedding := make([]float32, 1024)
			for i := range queryEmbedding {
				queryEmbedding[i] = 0.5
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				opts := &QueryOptions{
					Namespace: "default",
					Limit:     limit,
					Threshold: 0.0,
				}
				_, err := storage.Query(queryEmbedding, opts)
				if err != nil {
					b.Fatalf("Failed to query: %v", err)
				}
			}
		})
	}
}

// BenchmarkGet benchmarks retrieving a memory by ID
func BenchmarkGet(b *testing.B) {
	dataPath := filepath.Join(os.TempDir(), fmt.Sprintf("mem-bench-get-%d", time.Now().UnixNano()))
	defer cleanupTestDB(dataPath)

	config := &ChromemConfig{
		DataPath:        dataPath,
		CollectionName:  "bench_memories",
		DefaultNamespace: "default",
		EmbeddingDim:    1024,
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Store a test memory
	testID := "test-memory-123"
	memory := &models.Memory{
		ID:        testID,
		Content:   "Test memory for Get benchmark",
		Embedding: make([]float32, 1024),
		Namespace: "default",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := storage.Store(memory); err != nil {
		b.Fatalf("Failed to store memory: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := storage.Get(testID)
		if err != nil {
			b.Fatalf("Failed to get memory: %v", err)
		}
	}
}

// BenchmarkList benchmarks listing memories
func BenchmarkList(b *testing.B) {
	dataPath := filepath.Join(os.TempDir(), fmt.Sprintf("mem-bench-list-%d", time.Now().UnixNano()))
	defer cleanupTestDB(dataPath)

	config := &ChromemConfig{
		DataPath:        dataPath,
		CollectionName:  "bench_memories",
		DefaultNamespace: "default",
		EmbeddingDim:    1024,
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Pre-populate with memories
	sizes := []int{100, 500, 1000}
	for _, size := range sizes {
		// Clean up between iterations
		cleanupTestDB(dataPath)
		os.MkdirAll(dataPath, 0755)

		storage, err = NewChromemStorage(config)
		if err != nil {
			b.Fatalf("Failed to create storage: %v", err)
		}

		for i := 0; i < size; i++ {
			memory := &models.Memory{
				ID:        fmt.Sprintf("memory-%d", i),
				Content:   fmt.Sprintf("Test memory %d for listing", i),
				Embedding: make([]float32, 1024),
				Namespace: "default",
				Tags:      []string{"tag1", "tag2", "tag3"},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := storage.Store(memory); err != nil {
				b.Fatalf("Failed to store memory: %v", err)
			}
		}

		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := storage.List("default", "", 0)
				if err != nil {
					b.Fatalf("Failed to list: %v", err)
				}
			}
		})
	}
}

// BenchmarkUpdate benchmarks updating a memory
func BenchmarkUpdate(b *testing.B) {
	dataPath := filepath.Join(os.TempDir(), fmt.Sprintf("mem-bench-update-%d", time.Now().UnixNano()))
	defer cleanupTestDB(dataPath)

	config := &ChromemConfig{
		DataPath:        dataPath,
		CollectionName:  "bench_memories",
		DefaultNamespace: "default",
		EmbeddingDim:    1024,
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Store a test memory
	testID := "test-memory-update"
	memory := &models.Memory{
		ID:        testID,
		Content:   "Original content",
		Embedding: make([]float32, 1024),
		Namespace: "default",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := storage.Store(memory); err != nil {
		b.Fatalf("Failed to store memory: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		memory.Content = fmt.Sprintf("Updated content iteration %d", i)
		memory.UpdatedAt = time.Now()
		if err := storage.Update(memory); err != nil {
			b.Fatalf("Failed to update memory: %v", err)
		}
	}
}

// BenchmarkDelete benchmarks deleting a memory
func BenchmarkDelete(b *testing.B) {
	dataPath := filepath.Join(os.TempDir(), fmt.Sprintf("mem-bench-delete-%d", time.Now().UnixNano()))
	defer cleanupTestDB(dataPath)

	config := &ChromemConfig{
		DataPath:        dataPath,
		CollectionName:  "bench_memories",
		DefaultNamespace: "default",
		EmbeddingDim:    1024,
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Pre-populate with memories
	numMemories := b.N
	memories := make([]*models.Memory, numMemories)
	for i := 0; i < numMemories; i++ {
		memories[i] = &models.Memory{
			ID:        fmt.Sprintf("delete-test-%d", i),
			Content:   "Memory to be deleted",
			Embedding: make([]float32, 1024),
			Namespace: "default",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := storage.Store(memories[i]); err != nil {
			b.Fatalf("Failed to store memory: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := storage.Delete(memories[i].ID); err != nil {
			b.Fatalf("Failed to delete memory: %v", err)
		}
	}
}

// BenchmarkNamespaceOperations benchmarks namespace-related operations
func BenchmarkNamespaceOperations(b *testing.B) {
	dataPath := filepath.Join(os.TempDir(), fmt.Sprintf("mem-bench-namespace-%d", time.Now().UnixNano()))
	defer cleanupTestDB(dataPath)

	config := &ChromemConfig{
		DataPath:        dataPath,
		CollectionName:  "bench_memories",
		DefaultNamespace: "default",
		EmbeddingDim:    1024,
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Pre-populate with multiple namespaces
	numNamespaces := 10
	memoriesPerNamespace := 50
	for ns := 0; ns < numNamespaces; ns++ {
		namespace := fmt.Sprintf("ns-%d", ns)
		for i := 0; i < memoriesPerNamespace; i++ {
			memory := &models.Memory{
				ID:        fmt.Sprintf("%s-memory-%d", namespace, i),
				Content:   fmt.Sprintf("Memory in namespace %s", namespace),
				Embedding: make([]float32, 1024),
				Namespace: namespace,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if err := storage.Store(memory); err != nil {
				b.Fatalf("Failed to store memory: %v", err)
			}
		}
	}

	b.Run("ListNamespaces", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := storage.ListNamespaces()
			if err != nil {
				b.Fatalf("Failed to list namespaces: %v", err)
			}
		}
	})

	b.Run("GetByNamespace", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			namespace := fmt.Sprintf("ns-%d", i%numNamespaces)
			_, err := storage.GetByNamespace(namespace)
			if err != nil {
				b.Fatalf("Failed to get by namespace: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentOperations benchmarks concurrent access to storage
func BenchmarkConcurrentOperations(b *testing.B) {
	dataPath := filepath.Join(os.TempDir(), fmt.Sprintf("mem-bench-concurrent-%d", time.Now().UnixNano()))
	defer cleanupTestDB(dataPath)

	config := &ChromemConfig{
		DataPath:        dataPath,
		CollectionName:  "bench_memories",
		DefaultNamespace: "default",
		EmbeddingDim:    1024,
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	b.Run("ConcurrentStore", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				memory := &models.Memory{
					ID:        fmt.Sprintf("concurrent-%d", i),
					Content:   "Concurrent store test",
					Embedding: make([]float32, 1024),
					Namespace: "default",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				if err := storage.Store(memory); err != nil {
					b.Errorf("Failed to store: %v", err)
				}
				i++
			}
		})
	})

	b.Run("ConcurrentQuery", func(b *testing.B) {
		// Pre-populate
		for i := 0; i < 100; i++ {
			memory := &models.Memory{
				ID:        fmt.Sprintf("query-test-%d", i),
				Content:   "Query test memory",
				Embedding: make([]float32, 1024),
				Namespace: "default",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			storage.Store(memory)
		}

		queryEmbedding := make([]float32, 1024)
		for i := range queryEmbedding {
			queryEmbedding[i] = 0.5
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				opts := &QueryOptions{
					Namespace: "default",
					Limit:     10,
				}
				_, err := storage.Query(queryEmbedding, opts)
				if err != nil {
					b.Errorf("Failed to query: %v", err)
				}
			}
		})
	})
}
