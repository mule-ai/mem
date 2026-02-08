package cli

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jbutlerdev/mem/internal/config"
	"github.com/jbutlerdev/mem/internal/models"
	"github.com/jbutlerdev/mem/internal/storage"
	"github.com/spf13/viper"
)

// ErrMemoryNotFound is returned when a memory is not found
var ErrMemoryNotFound = errors.New("memory not found")

// mockStorage is a mock implementation of the Storage interface for testing
type mockStorage struct {
	memories      map[string]*models.Memory
	storeCalled   bool
	queryCalled   bool
	getCalled     bool
	updateCalled  bool
	deleteCalled  bool
	listCalled    bool
	lastQueryEmb  []float32
	lastQueryOpts *storage.QueryOptions
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		memories: make(map[string]*models.Memory),
	}
}

func (m *mockStorage) Store(memory *models.Memory) error {
	m.storeCalled = true
	m.memories[memory.ID] = memory
	return nil
}

func (m *mockStorage) Query(embedding []float32, opts *storage.QueryOptions) ([]models.SearchResult, error) {
	m.queryCalled = true
	m.lastQueryEmb = embedding
	m.lastQueryOpts = opts
	
	results := make([]models.SearchResult, 0)
	for _, mem := range m.memories {
		if opts.Namespace != "" && mem.Namespace != opts.Namespace {
			continue
		}
		// Return a mock result with a fixed score
		results = append(results, models.SearchResult{
			Memory: *mem,
			Score:  0.85,
			Rank:   1,
		})
	}
	return results, nil
}

func (m *mockStorage) Get(id string) (*models.Memory, error) {
	m.getCalled = true
	if mem, exists := m.memories[id]; exists {
		return mem, nil
	}
	return nil, ErrMemoryNotFound
}

func (m *mockStorage) Update(memory *models.Memory) error {
	m.updateCalled = true
	m.memories[memory.ID] = memory
	return nil
}

func (m *mockStorage) Delete(id string) error {
	m.deleteCalled = true
	delete(m.memories, id)
	return nil
}

func (m *mockStorage) List(namespace string, tag string, limit int) ([]*models.Memory, error) {
	m.listCalled = true
	results := make([]*models.Memory, 0)
	for _, mem := range m.memories {
		if namespace != "" && mem.Namespace != namespace {
			continue
		}
		if tag != "" && !hasTag(mem, tag) {
			continue
		}
		if limit > 0 && len(results) >= limit {
			break
		}
		results = append(results, mem)
	}
	return results, nil
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
		if hasTag(mem, tag) {
			results = append(results, mem)
		}
	}
	return results, nil
}

func (m *mockStorage) DeleteByNamespace(namespace string) error {
	for id, mem := range m.memories {
		if mem.Namespace == namespace {
			delete(m.memories, id)
		}
	}
	return nil
}

func (m *mockStorage) ListNamespaces() ([]storage.NamespaceInfo, error) {
	namespaceSet := make(map[string]bool)
	for _, mem := range m.memories {
		namespaceSet[mem.Namespace] = true
	}
	
	results := make([]storage.NamespaceInfo, 0, len(namespaceSet))
	for ns := range namespaceSet {
		results = append(results, storage.NamespaceInfo{Name: ns})
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
	// Mock implementation - just check if namespace exists
	for _, mem := range m.memories {
		if mem.Namespace == namespace {
			return nil // Already exists
		}
	}
	return nil
}

func (m *mockStorage) DeleteNamespace(namespace string) error {
	return m.DeleteByNamespace(namespace)
}

func (m *mockStorage) Close() error {
	return nil
}

func hasTag(mem *models.Memory, tag string) bool {
	for _, t := range mem.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

func hasAnyTag(mem *models.Memory, tags []string) bool {
	tagMap := make(map[string]bool)
	for _, t := range mem.Tags {
		tagMap[t] = true
	}
	for _, t := range tags {
		if tagMap[t] {
			return true
		}
	}
	return false
}

// TestStoreCmdBasic tests basic store command functionality
func TestStoreCmdBasic(t *testing.T) {
	// Create a temporary directory for test config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create a test config
	testConfig := config.DefaultConfig()
	testConfig.Memory.Path = tmpDir
	testConfig.Memory.Backend = "chromem"
	testConfig.Embeddings.BaseURL = "http://localhost:8080"
	testConfig.Embeddings.Model = "test-model"
	testConfig.CLI.Verbose = false

	// Write config file
	v := viper.New()
	v.Set("config", configPath)

	// Note: This test would require mocking the embedding client
	// For now, we test the command structure and flag parsing
	t.Run("StoreCommandExists", func(t *testing.T) {
		if StoreCmd == nil {
			t.Fatal("StoreCmd should not be nil")
		}
		if StoreCmd.Use != "store [content]" {
			t.Errorf("Expected use 'store [content]', got '%s'", StoreCmd.Use)
		}
		if StoreCmd.Short != "Store a new memory in semantic memory" {
			t.Errorf("Expected short description 'Store a new memory in semantic memory', got '%s'", StoreCmd.Short)
		}
	})

	t.Run("StoreCommandFlags", func(t *testing.T) {
		// Check that flags are defined
		flags := StoreCmd.Flags()
		
		namespaceFlag := flags.Lookup("namespace")
		if namespaceFlag == nil {
			t.Error("namespace flag should be defined")
		} else if namespaceFlag.Shorthand != "n" {
			t.Errorf("Expected namespace shorthand 'n', got '%s'", namespaceFlag.Shorthand)
		}

		tagFlag := flags.Lookup("tag")
		if tagFlag == nil {
			t.Error("tag flag should be defined")
		} else if tagFlag.Shorthand != "t" {
			t.Errorf("Expected tag shorthand 't', got '%s'", tagFlag.Shorthand)
		}

		metadataFlag := flags.Lookup("metadata")
		if metadataFlag == nil {
			t.Error("metadata flag should be defined")
		} else if metadataFlag.Shorthand != "m" {
			t.Errorf("Expected metadata shorthand 'm', got '%s'", metadataFlag.Shorthand)
		}

		stdinFlag := flags.Lookup("stdin")
		if stdinFlag == nil {
			t.Error("stdin flag should be defined")
		}
	})
}

// TestGenerateID tests the ID generation function
func TestGenerateID(t *testing.T) {
	ids := make(map[string]bool)
	
	// Generate multiple IDs and check for uniqueness
	for i := 0; i < 100; i++ {
		id := generateID()
		if ids[id] {
			t.Errorf("Generated duplicate ID: %s", id)
		}
		ids[id] = true
	}
}

// TestTruncateString tests the string truncation function
func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "long string",
			input:    "hello world",
			maxLen:   8,
			expected: "hello...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestOutputText tests the text output function
func TestOutputText(t *testing.T) {
	mem := &models.Memory{
		ID:        "test123",
		Content:   "This is a test memory with some content",
		Namespace: "test",
		Tags:      []string{"tag1", "tag2"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	outputText(mem)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Check that output contains expected content
	if !contains(output, "✓ Memory stored successfully") {
		t.Error("Output should contain success message")
	}
	if !contains(output, "test123") {
		t.Error("Output should contain memory ID")
	}
	if !contains(output, "test") {
		t.Error("Output should contain namespace")
	}
	if !contains(output, "tag1") {
		t.Error("Output should contain tags")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		indexOf(s, substr) >= 0)
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// TestValidateMemory tests memory validation
func TestValidateMemory(t *testing.T) {
	tests := []struct {
		name    string
		memory  *models.Memory
		wantErr bool
	}{
		{
			name: "valid memory",
			memory: &models.Memory{
				ID:        "test123",
				Content:   "Valid content",
				Embedding: []float32{0.1, 0.2, 0.3},
				Namespace: "default",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			memory: &models.Memory{
				ID:        "",
				Content:   "Valid content",
				Embedding: []float32{0.1, 0.2, 0.3},
				Namespace: "default",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty content",
			memory: &models.Memory{
				ID:        "test123",
				Content:   "",
				Embedding: []float32{0.1, 0.2, 0.3},
				Namespace: "default",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "empty namespace",
			memory: &models.Memory{
				ID:        "test123",
				Content:   "Valid content",
				Embedding: []float32{0.1, 0.2, 0.3},
				Namespace: "",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.memory.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Memory.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}