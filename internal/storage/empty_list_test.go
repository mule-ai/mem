package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jbutlerdev/mem/internal/models"
)

// TestEmptyCollectionList tests listing from an empty collection
func TestEmptyCollectionList(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mem-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataPath := filepath.Join(tmpDir, "data")

	// Create storage
	store, err := NewChromeMStorageWithDim(dataPath, 1024)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Try to list from empty collection
	t.Run("List empty collection", func(t *testing.T) {
		memories, err := store.List("default", "", 0)
		if err != nil {
			t.Errorf("Failed to list from empty collection: %v", err)
		}

		if len(memories) != 0 {
			t.Errorf("Expected 0 memories from empty collection, got %d", len(memories))
		}
	})

	// Try with a limit
	t.Run("List empty collection with limit", func(t *testing.T) {
		memories, err := store.List("default", "", 10)
		if err != nil {
			t.Errorf("Failed to list with limit from empty collection: %v", err)
		}

		if len(memories) != 0 {
			t.Errorf("Expected 0 memories with limit=10 from empty collection, got %d", len(memories))
		}
	})
}

// TestListWithOneItemRequestingFive tests the specific case: 1 item in collection, limit=5
func TestListWithOneItemRequestingFive(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mem-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataPath := filepath.Join(tmpDir, "data")

	// Create storage
	store, err := NewChromeMStorageWithDim(dataPath, 1024)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Add exactly 1 memory
	now := time.Now()
	memory := &models.Memory{
		ID:        "test-single",
		Content:   "Single test memory",
		Embedding: make([]float32, 1024),
		Namespace: "default",
		Tags:      []string{"test"},
		Metadata:  nil,
		CreatedAt: now,
		UpdatedAt: now,
	}
	// Fill embedding with dummy values
	for j := range memory.Embedding {
		memory.Embedding[j] = 0.1
	}
	if err := store.Store(memory); err != nil {
		t.Fatalf("Failed to store memory: %v", err)
	}

	// Request 5 results when only 1 exists - this should NOT error
	t.Run("List with limit=5, collection has 1 item", func(t *testing.T) {
		memories, err := store.List("default", "", 5)
		if err != nil {
			t.Errorf("Failed to list with limit=5 on collection with 1 item: %v", err)
		}

		if len(memories) != 1 {
			t.Errorf("Expected 1 memory when requesting 5, got %d", len(memories))
		}
	})
}

// TestFewerResultsThanLimit tests listing when there are fewer results than the requested limit
func TestFewerResultsThanLimit(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mem-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataPath := filepath.Join(tmpDir, "data")

	// Create storage
	store, err := NewChromeMStorageWithDim(dataPath, 1024)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Add 3 memories
	now := time.Now()
	for i := 1; i <= 3; i++ {
		memory := &models.Memory{
			ID:        generateTestID(),
			Content:   "Test memory " + string(rune('0'+i)),
			Embedding: make([]float32, 1024),
			Namespace: "default",
			Tags:      []string{"test"},
			Metadata:  nil,
			CreatedAt: now,
			UpdatedAt: now,
		}
		// Fill embedding with dummy values
		for j := range memory.Embedding {
			memory.Embedding[j] = 0.1
		}
		if err := store.Store(memory); err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}
	}

	// Request more results than available
	t.Run("Request more results than available", func(t *testing.T) {
		memories, err := store.List("default", "", 10)
		if err != nil {
			t.Errorf("Failed to list with limit higher than available: %v", err)
		}

		if len(memories) != 3 {
			t.Errorf("Expected 3 memories when requesting 10, got %d", len(memories))
		}
	})

	// Request exact number of results
	t.Run("Request exact number of results", func(t *testing.T) {
		memories, err := store.List("default", "", 3)
		if err != nil {
			t.Errorf("Failed to list with exact limit: %v", err)
		}

		if len(memories) != 3 {
			t.Errorf("Expected 3 memories when requesting 3, got %d", len(memories))
		}
	})

	// Request fewer results than available
	t.Run("Request fewer results than available", func(t *testing.T) {
		memories, err := store.List("default", "", 2)
		if err != nil {
			t.Errorf("Failed to list with lower limit: %v", err)
		}

		if len(memories) != 2 {
			t.Errorf("Expected 2 memories when requesting 2, got %d", len(memories))
		}
	})
}

// Helper function to generate test IDs
func generateTestID() string {
	return "test-" + time.Now().Format("20060102150405.000000000")
}

// TestEmptyCollectionQuery tests querying from an empty collection
func TestEmptyCollectionQuery(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mem-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataPath := filepath.Join(tmpDir, "data")

	// Create storage
	store, err := NewChromeMStorageWithDim(dataPath, 1024)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Create a query embedding
	queryEmbedding := make([]float32, 1024)
	for i := range queryEmbedding {
		queryEmbedding[i] = 0.1
	}

	// Try to query empty collection
	t.Run("Query empty collection", func(t *testing.T) {
		results, err := store.Query(queryEmbedding, &QueryOptions{
			Namespace: "default",
			Limit:     5,
			Threshold: 0.6,
		})
		if err != nil {
			t.Errorf("Failed to query empty collection: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("Expected 0 results from empty collection, got %d", len(results))
		}
	})
}

// TestQueryWithOneItemRequestingFive tests the specific case: 1 item in collection, n_results=5
func TestQueryWithOneItemRequestingFive(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mem-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataPath := filepath.Join(tmpDir, "data")

	// Create storage
	store, err := NewChromeMStorageWithDim(dataPath, 1024)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Add exactly 1 memory
	now := time.Now()
	memory := &models.Memory{
		ID:        "test-single",
		Content:   "Single test memory",
		Embedding: make([]float32, 1024),
		Namespace: "default",
		Tags:      []string{"test"},
		Metadata:  nil,
		CreatedAt: now,
		UpdatedAt: now,
	}
	// Fill embedding with dummy values
	for j := range memory.Embedding {
		memory.Embedding[j] = 0.1
	}
	if err := store.Store(memory); err != nil {
		t.Fatalf("Failed to store memory: %v", err)
	}

	// Create a query embedding
	queryEmbedding := make([]float32, 1024)
	for i := range queryEmbedding {
		queryEmbedding[i] = 0.1
	}

	// Request 5 results when only 1 exists - this should NOT error
	t.Run("Query with n_results=5, collection has 1 item", func(t *testing.T) {
		results, err := store.Query(queryEmbedding, &QueryOptions{
			Namespace: "default",
			Limit:     5,
			Threshold: 0.0,
		})
		if err != nil {
			t.Errorf("Failed to query with n_results=5 on collection with 1 item: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result when requesting 5, got %d", len(results))
		}
	})
}

// TestQueryFewerResultsThanLimit tests querying when there are fewer results than requested
func TestQueryFewerResultsThanLimit(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "mem-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dataPath := filepath.Join(tmpDir, "data")

	// Create storage
	store, err := NewChromeMStorageWithDim(dataPath, 1024)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Add 3 memories
	now := time.Now()
	for i := 1; i <= 3; i++ {
		memory := &models.Memory{
			ID:        generateTestID(),
			Content:   "Test memory " + string(rune('0'+i)),
			Embedding: make([]float32, 1024),
			Namespace: "default",
			Tags:      []string{"test"},
			Metadata:  nil,
			CreatedAt: now,
			UpdatedAt: now,
		}
		// Fill embedding with dummy values
		for j := range memory.Embedding {
			memory.Embedding[j] = 0.1
		}
		if err := store.Store(memory); err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}
	}

	// Create a query embedding
	queryEmbedding := make([]float32, 1024)
	for i := range queryEmbedding {
		queryEmbedding[i] = 0.1
	}

	// Request more results than available
	t.Run("Request more query results than available", func(t *testing.T) {
		results, err := store.Query(queryEmbedding, &QueryOptions{
			Namespace: "default",
			Limit:     10,
			Threshold: 0.0, // Set low threshold to get all results
		})
		if err != nil {
			t.Errorf("Failed to query with limit higher than available: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results when requesting 10, got %d", len(results))
		}
	})

	// Request exact number of results
	t.Run("Request exact number of query results", func(t *testing.T) {
		results, err := store.Query(queryEmbedding, &QueryOptions{
			Namespace: "default",
			Limit:     3,
			Threshold: 0.0,
		})
		if err != nil {
			t.Errorf("Failed to query with exact limit: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results when requesting 3, got %d", len(results))
		}
	})

	// Request fewer results than available
	t.Run("Request fewer query results than available", func(t *testing.T) {
		results, err := store.Query(queryEmbedding, &QueryOptions{
			Namespace: "default",
			Limit:     2,
			Threshold: 0.0,
		})
		if err != nil {
			t.Errorf("Failed to query with lower limit: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results when requesting 2, got %d", len(results))
		}
	})
}
