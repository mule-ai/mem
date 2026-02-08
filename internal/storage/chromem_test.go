package storage

import (
	"os"
	"testing"

	"github.com/jbutlerdev/mem/internal/models"
)

func TestChromemStorage(t *testing.T) {
	// Create a temporary directory for test data
	tempDir, err := os.MkdirTemp("", "mem-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create storage
	config := &ChromemConfig{
		DataPath:        tempDir,
		CollectionName:  "test-memories",
		DefaultNamespace: "test",
		EmbeddingDim:    1024,
		EmbeddingFunc:   nil, // We'll provide embeddings manually
	}

	storage, err := NewChromemStorage(config)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	t.Run("Store memory", func(t *testing.T) {
		// Create a test memory with mock embedding
		embedding := make([]float32, 1024)
		for i := range embedding {
			embedding[i] = 0.1
		}

		memory, err := models.NewMemory(
			"test-id-1",
			"Test memory content",
			"test",
			embedding,
			[]string{"test", "unit"},
		)
		if err != nil {
			t.Fatalf("Failed to create memory: %v", err)
		}

		// Store the memory
		err = storage.Store(memory)
		if err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}

		// Verify we can retrieve it
		retrieved, err := storage.Get("test-id-1")
		if err != nil {
			t.Fatalf("Failed to retrieve memory: %v", err)
		}

		if retrieved.ID != "test-id-1" {
			t.Errorf("Expected ID test-id-1, got %s", retrieved.ID)
		}
		if retrieved.Content != "Test memory content" {
			t.Errorf("Expected content 'Test memory content', got %s", retrieved.Content)
		}
		if retrieved.Namespace != "test" {
			t.Errorf("Expected namespace 'test', got %s", retrieved.Namespace)
		}
	})

	t.Run("Query memories", func(t *testing.T) {
		// Create a query embedding (similar to the stored memory)
		queryEmbedding := make([]float32, 1024)
		for i := range queryEmbedding {
			queryEmbedding[i] = 0.1
		}

		// Query with default options - reduce limit to 1 since we only have 1 document
		opts := &QueryOptions{
			Limit:     1,
			Threshold: 0.0, // Set low threshold to ensure we get results
		}

		results, err := storage.Query(queryEmbedding, opts)
		if err != nil {
			t.Fatalf("Failed to query: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result, got none")
		} else {
			if results[0].Memory.ID != "test-id-1" {
				t.Errorf("Expected first result ID test-id-1, got %s", results[0].Memory.ID)
			}
		}
	})

	t.Run("Query with namespace filter", func(t *testing.T) {
		// First, let's add more documents to have enough to query
		// Create another memory in a different namespace
		embedding := make([]float32, 1024)
		for i := range embedding {
			embedding[i] = 0.15
		}

		memory2, err := models.NewMemory(
			"test-id-2",
			"Another test memory",
			"other-namespace",
			embedding,
			[]string{"other"},
		)
		if err != nil {
			t.Fatalf("Failed to create memory: %v", err)
		}

		err = storage.Store(memory2)
		if err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}

		// Add one more to the same namespace so we have 2 documents
		memory3, err := models.NewMemory(
			"test-id-3",
			"Yet another memory",
			"other-namespace",
			embedding,
			[]string{"other"},
		)
		if err != nil {
			t.Fatalf("Failed to create memory: %v", err)
		}

		err = storage.Store(memory3)
		if err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}

		// Query with namespace filter
		queryEmbedding := make([]float32, 1024)
		for i := range queryEmbedding {
			queryEmbedding[i] = 0.1
		}

		opts := &QueryOptions{
			Namespace: "other-namespace",
			Limit:     2,
			TopK:      2,
			Threshold: 0.0,
		}

		results, err := storage.Query(queryEmbedding, opts)
		if err != nil {
			t.Fatalf("Failed to query: %v", err)
		}

		// Should only get results from other-namespace
		for _, result := range results {
			if result.Memory.Namespace != "other-namespace" {
				t.Errorf("Expected namespace 'other-namespace', got %s", result.Memory.Namespace)
			}
		}
	})

	t.Run("Update memory", func(t *testing.T) {
		// Get existing memory
		memory, err := storage.Get("test-id-1")
		if err != nil {
			t.Fatalf("Failed to get memory: %v", err)
		}

		// Update content
		memory.Content = "Updated test memory content"
		memory.Touch()

		// Update in storage
		err = storage.Update(memory)
		if err != nil {
			t.Fatalf("Failed to update memory: %v", err)
		}

		// Verify update
		updated, err := storage.Get("test-id-1")
		if err != nil {
			t.Fatalf("Failed to retrieve updated memory: %v", err)
		}

		if updated.Content != "Updated test memory content" {
			t.Errorf("Expected content 'Updated test memory content', got %s", updated.Content)
		}
	})

	t.Run("Delete memory", func(t *testing.T) {
		// Create a memory to delete
		embedding := make([]float32, 1024)
		for i := range embedding {
			embedding[i] = 0.2
		}

		memory, err := models.NewMemory(
			"test-id-delete",
			"Memory to delete",
			"test",
			embedding,
			nil,
		)
		if err != nil {
			t.Fatalf("Failed to create memory: %v", err)
		}

		err = storage.Store(memory)
		if err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}

		// Verify it exists
		_, err = storage.Get("test-id-delete")
		if err != nil {
			t.Fatalf("Failed to retrieve memory before delete: %v", err)
		}

		// Delete it
		err = storage.Delete("test-id-delete")
		if err != nil {
			t.Fatalf("Failed to delete memory: %v", err)
		}

		// Verify it's gone
		_, err = storage.Get("test-id-delete")
		if err == nil {
			t.Error("Expected error when retrieving deleted memory, got nil")
		}
	})

	t.Run("Delete by namespace", func(t *testing.T) {
		// Create memories in a namespace to delete
		embedding := make([]float32, 1024)
		for i := range embedding {
			embedding[i] = 0.3
		}

		memory1, _ := models.NewMemory("delete-ns-1", "Memory 1", "delete-ns", embedding, nil)
		memory2, _ := models.NewMemory("delete-ns-2", "Memory 2", "delete-ns", embedding, nil)

		storage.Store(memory1)
		storage.Store(memory2)

		// Delete entire namespace
		err := storage.DeleteByNamespace("delete-ns")
		if err != nil {
			t.Fatalf("Failed to delete namespace: %v", err)
		}

		// Verify memories are gone
		_, err = storage.Get("delete-ns-1")
		if err == nil {
			t.Error("Expected error when retrieving deleted memory 1, got nil")
		}
	})
}

func TestChromemStorageWithPersistence(t *testing.T) {
	// Create a temporary directory for test data
	tempDir, err := os.MkdirTemp("", "mem-test-persist-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// First storage instance
	config1 := &ChromemConfig{
		DataPath:        tempDir,
		CollectionName:  "persist-test",
		DefaultNamespace: "test",
		EmbeddingFunc:   nil,
	}

	storage1, err := NewChromemStorage(config1)
	if err != nil {
		t.Fatalf("Failed to create first storage: %v", err)
	}

	// Store a memory
	embedding := make([]float32, 1024)
	for i := range embedding {
		embedding[i] = 0.5
	}

	memory, _ := models.NewMemory("persist-1", "Persistent memory", "test", embedding, nil)
	err = storage1.Store(memory)
	if err != nil {
		t.Fatalf("Failed to store memory: %v", err)
	}

	storage1.Close()

	// Create second storage instance with same config
	// Should load existing data
	config2 := &ChromemConfig{
		DataPath:        tempDir,
		CollectionName:  "persist-test",
		DefaultNamespace: "test",
		EmbeddingFunc:   nil,
	}

	storage2, err := NewChromemStorage(config2)
	if err != nil {
		t.Fatalf("Failed to create second storage: %v", err)
	}
	defer storage2.Close()

	// Verify we can retrieve the stored memory
	retrieved, err := storage2.Get("persist-1")
	if err != nil {
		t.Fatalf("Failed to retrieve persisted memory: %v", err)
	}

	if retrieved.Content != "Persistent memory" {
		t.Errorf("Expected content 'Persistent memory', got %s", retrieved.Content)
	}
}

func TestDefaultQueryOptions(t *testing.T) {
	opts := DefaultQueryOptions()

	if opts.Limit != 5 {
		t.Errorf("Expected Limit 5, got %d", opts.Limit)
	}
	if opts.Threshold != 0.6 {
		t.Errorf("Expected Threshold 0.6, got %f", opts.Threshold)
	}
	if opts.TopK != 10 {
		t.Errorf("Expected TopK 10, got %d", opts.TopK)
	}
}
