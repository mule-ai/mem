package storage

import (
	"testing"
	"time"

	"github.com/jbutlerdev/mem/internal/models"
)

// TestNamespaceInfo tests the NamespaceInfo structure
func TestNamespaceInfo(t *testing.T) {
	t.Run("CreateNamespaceInfo", func(t *testing.T) {
		now := time.Now()
		info := NamespaceInfo{
			Name:      "test-namespace",
			CreatedAt: now,
		}
		
		if info.Name != "test-namespace" {
			t.Errorf("Expected name 'test-namespace', got '%s'", info.Name)
		}
		if info.CreatedAt != now {
			t.Error("CreatedAt should match the provided time")
		}
	})

	t.Run("NamespaceInfoWithStringCreatedAt", func(t *testing.T) {
		info := NamespaceInfo{
			Name:      "test-namespace",
			CreatedAt: "2024-01-01T00:00:00Z",
		}
		
		if info.Name != "test-namespace" {
			t.Errorf("Expected name 'test-namespace', got '%s'", info.Name)
		}
		if info.CreatedAt != "2024-01-01T00:00:00Z" {
			t.Error("CreatedAt should match the provided string")
		}
	})
}

// TestMemoryValidationExtended tests additional validation scenarios
func TestMemoryValidationExtended(t *testing.T) {
	t.Run("ValidMemory", func(t *testing.T) {
		mem := &models.Memory{
			ID:        "test123",
			Content:   "Test content",
			Embedding: []float32{0.1, 0.2, 0.3},
			Namespace: "default",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		if err := mem.Validate(); err != nil {
			t.Errorf("Expected valid memory, got error: %v", err)
		}
	})

	t.Run("MemoryWithEmptyID", func(t *testing.T) {
		mem := &models.Memory{
			ID:        "",
			Content:   "Test content",
			Embedding: []float32{0.1, 0.2, 0.3},
			Namespace: "default",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		if err := mem.Validate(); err == nil {
			t.Error("Expected error for memory with empty ID")
		}
	})

	t.Run("MemoryWithEmptyContent", func(t *testing.T) {
		mem := &models.Memory{
			ID:        "test123",
			Content:   "",
			Embedding: []float32{0.1, 0.2, 0.3},
			Namespace: "default",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		if err := mem.Validate(); err == nil {
			t.Error("Expected error for memory with empty content")
		}
	})

	t.Run("MemoryWithEmptyNamespace", func(t *testing.T) {
		mem := &models.Memory{
			ID:        "test123",
			Content:   "Test content",
			Embedding: []float32{0.1, 0.2, 0.3},
			Namespace: "",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		if err := mem.Validate(); err == nil {
			t.Error("Expected error for memory with empty namespace")
		}
	})
}

// TestSearchResult tests the SearchResult structure
func TestSearchResult(t *testing.T) {
	t.Run("CreateSearchResult", func(t *testing.T) {
		mem := &models.Memory{
			ID:        "test123",
			Content:   "Test content",
			Namespace: "default",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		result := models.SearchResult{
			Memory:   *mem,
			Score:    0.85,
			Reranked: true,
			Rank:     1,
		}
		
		if result.Memory.ID != "test123" {
			t.Errorf("Expected memory ID 'test123', got '%s'", result.Memory.ID)
		}
		if result.Score != 0.85 {
			t.Errorf("Expected score 0.85, got %f", result.Score)
		}
		if !result.Reranked {
			t.Error("Expected reranked to be true")
		}
		if result.Rank != 1 {
			t.Errorf("Expected rank 1, got %d", result.Rank)
		}
	})
}

// Mock storage for testing storage interface compliance
type mockStorage struct {
	memories map[string]*models.Memory
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		memories: make(map[string]*models.Memory),
	}
}

func (m *mockStorage) Store(memory *models.Memory) error {
	m.memories[memory.ID] = memory
	return nil
}

func (m *mockStorage) Query(queryEmbedding []float32, opts *QueryOptions) ([]models.SearchResult, error) {
	results := make([]models.SearchResult, 0)
	for _, mem := range m.memories {
		if opts.Namespace != "" && mem.Namespace != opts.Namespace {
			continue
		}
		results = append(results, models.SearchResult{
			Memory: *mem,
			Score:  0.85,
			Rank:   len(results) + 1,
		})
		if opts.Limit > 0 && len(results) >= opts.Limit {
			break
		}
	}
	return results, nil
}

func (m *mockStorage) Get(id string) (*models.Memory, error) {
	if mem, exists := m.memories[id]; exists {
		return mem, nil
	}
	return nil, nil
}

func (m *mockStorage) GetByNamespace(namespace string) ([]*models.Memory, error) {
	results := make([]*models.Memory, 0)
	for _, mem := range m.memories {
		if mem.Namespace == namespace {
			results = append(results, mem)
		}
	}
	return results, nil
}

func (m *mockStorage) GetByTag(tag string) ([]*models.Memory, error) {
	results := make([]*models.Memory, 0)
	for _, mem := range m.memories {
		for _, t := range mem.Tags {
			if t == tag {
				results = append(results, mem)
				break
			}
		}
	}
	return results, nil
}

func (m *mockStorage) Update(memory *models.Memory) error {
	m.memories[memory.ID] = memory
	return nil
}

func (m *mockStorage) Delete(id string) error {
	delete(m.memories, id)
	return nil
}

func (m *mockStorage) DeleteByNamespace(namespace string) error {
	for id, mem := range m.memories {
		if mem.Namespace == namespace {
			delete(m.memories, id)
		}
	}
	return nil
}

func (m *mockStorage) List(namespace string, tag string, limit int) ([]*models.Memory, error) {
	results := make([]*models.Memory, 0)
	for _, mem := range m.memories {
		if namespace != "" && mem.Namespace != namespace {
			continue
		}
		if tag != "" {
			found := false
			for _, t := range mem.Tags {
				if t == tag {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		results = append(results, mem)
		if limit > 0 && len(results) >= limit {
			break
		}
	}
	return results, nil
}

func (m *mockStorage) ListNamespaces() ([]NamespaceInfo, error) {
	namespaceSet := make(map[string]bool)
	for _, mem := range m.memories {
		namespaceSet[mem.Namespace] = true
	}
	
	results := make([]NamespaceInfo, 0, len(namespaceSet))
	for ns := range namespaceSet {
		results = append(results, NamespaceInfo{Name: ns})
	}
	return results, nil
}

func (m *mockStorage) Count(namespace string) (int, error) {
	count := 0
	for _, mem := range m.memories {
		if namespace == "" || mem.Namespace == namespace {
			count++
		}
	}
	return count, nil
}

func (m *mockStorage) CreateNamespace(namespace string) error {
	return nil
}

func (m *mockStorage) DeleteNamespace(namespace string) error {
	return m.DeleteByNamespace(namespace)
}

func (m *mockStorage) Close() error {
	return nil
}

// TestMockStorageExtended tests the mock storage implementation
func TestMockStorageExtended(t *testing.T) {
	t.Run("StoreAndGet", func(t *testing.T) {
		storage := newMockStorage()
		mem := &models.Memory{
			ID:        "test1",
			Content:   "Test content",
			Namespace: "default",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		err := storage.Store(mem)
		if err != nil {
			t.Errorf("Failed to store: %v", err)
		}
		
		retrieved, err := storage.Get("test1")
		if err != nil {
			t.Errorf("Failed to get: %v", err)
		}
		if retrieved == nil {
			t.Fatal("Retrieved memory is nil")
		}
		if retrieved.ID != "test1" {
			t.Errorf("Expected ID 'test1', got '%s'", retrieved.ID)
		}
	})

	t.Run("QueryWithNamespace", func(t *testing.T) {
		storage := newMockStorage()
		mem1 := &models.Memory{
			ID:        "test1",
			Content:   "Content 1",
			Namespace: "ns1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mem2 := &models.Memory{
			ID:        "test2",
			Content:   "Content 2",
			Namespace: "ns2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		storage.Store(mem1)
		storage.Store(mem2)
		
		opts := &QueryOptions{
			Namespace: "ns1",
			Limit:     10,
		}
		
		results, err := storage.Query([]float32{0.1}, opts)
		if err != nil {
			t.Errorf("Failed to query: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
		if len(results) > 0 && results[0].Memory.Namespace != "ns1" {
			t.Errorf("Expected namespace 'ns1', got '%s'", results[0].Memory.Namespace)
		}
	})

	t.Run("ListNamespaces", func(t *testing.T) {
		storage := newMockStorage()
		
		mem1 := &models.Memory{
			ID:        "test1",
			Content:   "Content 1",
			Namespace: "ns1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mem2 := &models.Memory{
			ID:        "test2",
			Content:   "Content 2",
			Namespace: "ns2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		storage.Store(mem1)
		storage.Store(mem2)
		
		namespaces, err := storage.ListNamespaces()
		if err != nil {
			t.Errorf("Failed to list namespaces: %v", err)
		}
		if len(namespaces) != 2 {
			t.Errorf("Expected 2 namespaces, got %d", len(namespaces))
		}
	})

	t.Run("Delete", func(t *testing.T) {
		storage := newMockStorage()
		mem := &models.Memory{
			ID:        "test1",
			Content:   "Content",
			Namespace: "default",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		storage.Store(mem)
		
		err := storage.Delete("test1")
		if err != nil {
			t.Errorf("Failed to delete: %v", err)
		}
		
		retrieved, _ := storage.Get("test1")
		if retrieved != nil {
			t.Error("Memory should be deleted")
		}
	})
}