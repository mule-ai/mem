package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jbutlerdev/mem/internal/models"
	"github.com/jbutlerdev/mem/internal/storage"
)

// TestIntegration_StoreQueryDelete tests a full workflow
func TestIntegration_StoreQueryDelete(t *testing.T) {
	// Create temporary directory for test data
	tmpDir := t.TempDir()
	dataPath := filepath.Join(tmpDir, "data")

	// Initialize storage
	store, err := storage.NewChromeMStorageWithDim(dataPath, 1024)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Create test memory
	now := time.Now()
	memory := &models.Memory{
		ID:        "test-mem-1",
		Content:   "User prefers dark mode and vim keybindings",
		Embedding: make([]float32, 1024), // Mock embedding
		Namespace: "default",
		Tags:      []string{"preferences", "editor"},
		Metadata: map[string]interface{}{
			"source": "test",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Initialize mock embedding with values
	for i := range memory.Embedding {
		memory.Embedding[i] = 0.01
	}

	// Store memory
	t.Run("Store memory", func(t *testing.T) {
		err := store.Store(memory)
		if err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}

		// Verify memory was stored
		retrieved, err := store.Get(memory.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve stored memory: %v", err)
		}

		if retrieved.ID != memory.ID {
			t.Errorf("Expected ID %s, got %s", memory.ID, retrieved.ID)
		}
		if retrieved.Content != memory.Content {
			t.Errorf("Expected content %s, got %s", memory.Content, retrieved.Content)
		}
		if retrieved.Namespace != memory.Namespace {
			t.Errorf("Expected namespace %s, got %s", memory.Namespace, retrieved.Namespace)
		}
	})

	// Query memories
	t.Run("Query memories", func(t *testing.T) {
		queryEmbedding := make([]float32, 1024)
		for i := range queryEmbedding {
			queryEmbedding[i] = 0.01
		}

		opts := &storage.QueryOptions{
			Namespace: "default",
			Limit:     1,  // Only request what we have
			Threshold: 0.5,
		}

		results, err := store.Query(queryEmbedding, opts)
		if err != nil {
			t.Fatalf("Failed to query memories: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least one result")
		} else {
			if results[0].Memory.ID != memory.ID {
				t.Errorf("Expected memory ID %s, got %s", memory.ID, results[0].Memory.ID)
			}
		}
	})

	// List memories
	t.Run("List memories", func(t *testing.T) {
		memories, err := store.List("default", "", 0)
		if err != nil {
			t.Fatalf("Failed to list memories: %v", err)
		}

		if len(memories) != 1 {
			t.Errorf("Expected 1 memory, got %d", len(memories))
		}

		if memories[0].ID != memory.ID {
			t.Errorf("Expected memory ID %s, got %s", memory.ID, memories[0].ID)
		}
	})

	// Update memory
	t.Run("Update memory", func(t *testing.T) {
		updatedMemory := *memory
		updatedMemory.Content = "Updated: User prefers light mode but still uses vim"
		updatedMemory.UpdatedAt = time.Now()

		err := store.Update(&updatedMemory)
		if err != nil {
			t.Fatalf("Failed to update memory: %v", err)
		}

		retrieved, err := store.Get(memory.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve updated memory: %v", err)
		}

		if retrieved.Content != updatedMemory.Content {
			t.Errorf("Expected updated content %s, got %s", updatedMemory.Content, retrieved.Content)
		}
	})

	// Delete memory
	t.Run("Delete memory", func(t *testing.T) {
		err := store.Delete(memory.ID)
		if err != nil {
			t.Fatalf("Failed to delete memory: %v", err)
		}

		_, err = store.Get(memory.ID)
		if err == nil {
			t.Error("Expected error when retrieving deleted memory")
		}
	})
}

// TestIntegration_Namespaces tests namespace isolation
func TestIntegration_Namespaces(t *testing.T) {
	tmpDir := t.TempDir()
	dataPath := filepath.Join(tmpDir, "data")

	store, err := storage.NewChromeMStorageWithDim(dataPath, 1024)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	now := time.Now()
	
	// Create memories in different namespaces
	memories := []*models.Memory{
		{
			ID:        "mem-default-1",
			Content:   "Memory in default namespace",
			Embedding: make([]float32, 1024),
			Namespace: "default",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "mem-work-1",
			Content:   "Memory in work namespace",
			Embedding: make([]float32, 1024),
			Namespace: "work",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "mem-work-2",
			Content:   "Another memory in work namespace",
			Embedding: make([]float32, 1024),
			Namespace: "work",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// Initialize embeddings
	for _, mem := range memories {
		for i := range mem.Embedding {
			mem.Embedding[i] = 0.01
		}
	}

	// Store all memories
	for _, mem := range memories {
		err := store.Store(mem)
		if err != nil {
			t.Fatalf("Failed to store memory %s: %v", mem.ID, err)
		}
	}

	// Query default namespace
	t.Run("Query default namespace", func(t *testing.T) {
		queryEmbedding := make([]float32, 1024)
		for i := range queryEmbedding {
			queryEmbedding[i] = 0.01
		}

		opts := &storage.QueryOptions{
			Namespace: "default",
			Limit:     1,
		}

		results, err := store.Query(queryEmbedding, opts)
		if err != nil {
			t.Fatalf("Failed to query: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result in default namespace, got %d", len(results))
		}
	})

	// Query work namespace
	t.Run("Query work namespace", func(t *testing.T) {
		queryEmbedding := make([]float32, 1024)
		for i := range queryEmbedding {
			queryEmbedding[i] = 0.01
		}

		opts := &storage.QueryOptions{
			Namespace: "work",
			Limit:     2,
		}

		results, err := store.Query(queryEmbedding, opts)
		if err != nil {
			t.Fatalf("Failed to query: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results in work namespace, got %d", len(results))
		}
	})

	// List by namespace
	t.Run("List by namespace", func(t *testing.T) {
		defaultMems, err := store.List("default", "", 0)
		if err != nil {
			t.Fatalf("Failed to list default: %v", err)
		}
		if len(defaultMems) != 1 {
			t.Errorf("Expected 1 memory in default, got %d", len(defaultMems))
		}

		workMems, err := store.List("work", "", 0)
		if err != nil {
			t.Fatalf("Failed to list work: %v", err)
		}
		if len(workMems) != 2 {
			t.Errorf("Expected 2 memories in work, got %d", len(workMems))
		}
	})

	// Count by namespace
	t.Run("Count by namespace", func(t *testing.T) {
		count, err := store.Count("default")
		if err != nil {
			t.Fatalf("Failed to count default: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected count 1 for default, got %d", count)
		}

		count, err = store.Count("work")
		if err != nil {
			t.Fatalf("Failed to count work: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected count 2 for work, got %d", count)
		}
	})
}

// TestIntegration_Tags tests tag filtering
func TestIntegration_Tags(t *testing.T) {
	tmpDir := t.TempDir()
	dataPath := filepath.Join(tmpDir, "data")

	store, err := storage.NewChromeMStorageWithDim(dataPath, 1024)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	now := time.Now()

	memories := []*models.Memory{
		{
			ID:        "mem-1",
			Content:   "Memory with preferences tag",
			Embedding: make([]float32, 1024),
			Namespace: "default",
			Tags:      []string{"preferences", "user"},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "mem-2",
			Content:   "Memory with work tag",
			Embedding: make([]float32, 1024),
			Namespace: "default",
			Tags:      []string{"work", "project"},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "mem-3",
			Content:   "Memory with both preferences and work tags",
			Embedding: make([]float32, 1024),
			Namespace: "default",
			Tags:      []string{"preferences", "work"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, mem := range memories {
		for i := range mem.Embedding {
			mem.Embedding[i] = 0.01
		}
		err := store.Store(mem)
		if err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}
	}

	// Get by tag
	t.Run("Get by tag", func(t *testing.T) {
		prefsMems, err := store.GetByTag("preferences")
		if err != nil {
			t.Fatalf("Failed to get by tag: %v", err)
		}
		if len(prefsMems) != 2 {
			t.Errorf("Expected 2 memories with preferences tag, got %d", len(prefsMems))
		}

		workMems, err := store.GetByTag("work")
		if err != nil {
			t.Fatalf("Failed to get by tag: %v", err)
		}
		if len(workMems) != 2 {
			t.Errorf("Expected 2 memories with work tag, got %d", len(workMems))
		}
	})

	// List with tag filter
	t.Run("List with tag filter", func(t *testing.T) {
		taggedMems, err := store.List("", "preferences", 0)
		if err != nil {
			t.Fatalf("Failed to list by tag: %v", err)
		}
		if len(taggedMems) != 2 {
			t.Errorf("Expected 2 memories when filtering by preferences tag, got %d", len(taggedMems))
		}
	})
}

// TestIntegration_ExportImport tests export/import round-trip
func TestIntegration_ExportImport(t *testing.T) {
	tmpDir := t.TempDir()
	dataPath := filepath.Join(tmpDir, "data")
	exportPath := filepath.Join(tmpDir, "export.json")

	store1, err := storage.NewChromeMStorageWithDim(dataPath, 1024)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	now := time.Now()

	memories := []*models.Memory{
		{
			ID:        "mem-1",
			Content:   "First memory",
			Embedding: make([]float32, 1024),
			Namespace: "default",
			Tags:      []string{"tag1"},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "mem-2",
			Content:   "Second memory",
			Embedding: make([]float32, 1024),
			Namespace: "work",
			Tags:      []string{"tag2"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	for _, mem := range memories {
		for i := range mem.Embedding {
			mem.Embedding[i] = 0.01
		}
		err := store1.Store(mem)
		if err != nil {
			t.Fatalf("Failed to store memory: %v", err)
		}
	}

	// Export memories
	t.Run("Export memories", func(t *testing.T) {
		allMems, err := store1.List("", "", 0)
		if err != nil {
			t.Fatalf("Failed to list memories: %v", err)
		}

		data, err := json.MarshalIndent(allMems, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal memories: %v", err)
		}

		err = os.WriteFile(exportPath, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write export file: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(exportPath); os.IsNotExist(err) {
			t.Error("Export file was not created")
		}
	})

	store1.Close()

	// Import memories
	t.Run("Import memories", func(t *testing.T) {
		data, err := os.ReadFile(exportPath)
		if err != nil {
			t.Fatalf("Failed to read export file: %v", err)
		}

		var imported []*models.Memory
		err = json.Unmarshal(data, &imported)
		if err != nil {
			t.Fatalf("Failed to unmarshal memories: %v", err)
		}

		if len(imported) != len(memories) {
			t.Errorf("Expected %d memories, got %d", len(memories), len(imported))
		}

		// Create a map for easy lookup
		memMap := make(map[string]*models.Memory)
		for _, mem := range memories {
			memMap[mem.ID] = mem
		}

		// Verify content by matching IDs
		for _, importedMem := range imported {
			original, exists := memMap[importedMem.ID]
			if !exists {
				t.Errorf("Imported memory %s not found in original", importedMem.ID)
				continue
			}
			if importedMem.Content != original.Content {
				t.Errorf("Memory %s: content mismatch", importedMem.ID)
			}
			if importedMem.Namespace != original.Namespace {
				t.Errorf("Memory %s: namespace mismatch", importedMem.ID)
			}
		}
	})
}

// TestIntegration_NamespaceManagement tests namespace operations
func TestIntegration_NamespaceManagement(t *testing.T) {
	tmpDir := t.TempDir()
	dataPath := filepath.Join(tmpDir, "data")

	store, err := storage.NewChromeMStorageWithDim(dataPath, 1024)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer store.Close()

	// Add memories to different namespaces
	now := time.Now()
	mem1 := &models.Memory{
		ID:        "test-mem-1",
		Content:   "Test memory 1",
		Embedding: make([]float32, 1024),
		Namespace: "test-ns-1",
		CreatedAt: now,
		UpdatedAt: now,
	}
	mem2 := &models.Memory{
		ID:        "test-mem-2",
		Content:   "Test memory 2",
		Embedding: make([]float32, 1024),
		Namespace: "test-ns-2",
		CreatedAt: now,
		UpdatedAt: now,
	}

	for i := range mem1.Embedding {
		mem1.Embedding[i] = 0.01
		mem2.Embedding[i] = 0.01
	}

	err = store.Store(mem1)
	if err != nil {
		t.Fatalf("Failed to store test memory 1: %v", err)
	}
	err = store.Store(mem2)
	if err != nil {
		t.Fatalf("Failed to store test memory 2: %v", err)
	}

	// Count by namespace
	t.Run("Count by namespace", func(t *testing.T) {
		// Namespaces in chromem-go are implicit - they exist when memories are stored there
		count1, err := store.Count("test-ns-1")
		if err != nil {
			t.Fatalf("Failed to count namespace 1: %v", err)
		}
		if count1 != 1 {
			t.Errorf("Expected count 1 for namespace 1, got %d", count1)
		}

		count2, err := store.Count("test-ns-2")
		if err != nil {
			t.Fatalf("Failed to count namespace 2: %v", err)
		}
		if count2 != 1 {
			t.Errorf("Expected count 1 for namespace 2, got %d", count2)
		}
	})

	// Delete namespace
	t.Run("Delete namespace", func(t *testing.T) {
		err := store.DeleteNamespace("test-ns-1")
		if err != nil {
			t.Fatalf("Failed to delete namespace: %v", err)
		}

		// Verify memories in other namespace still exist
		count2, err := store.Count("test-ns-2")
		if err != nil {
			t.Fatalf("Failed to count namespace 2 after delete: %v", err)
		}
		if count2 != 1 {
			t.Errorf("Expected count 1 for namespace 2 after deleting namespace 1, got %d", count2)
		}
	})
}