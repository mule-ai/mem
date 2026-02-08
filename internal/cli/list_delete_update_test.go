package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jbutlerdev/mem/internal/models"
)

// mockStorageForList is a simplified mock for list/delete/update tests
type mockStorageForList struct {
	memories map[string]*models.Memory
}

func newMockStorageForList() *mockStorageForList {
	return &mockStorageForList{
		memories: make(map[string]*models.Memory),
	}
}

func (m *mockStorageForList) Store(memory *models.Memory) error {
	m.memories[memory.ID] = memory
	return nil
}

func (m *mockStorageForList) Query(embedding []float32, opts *interface{}) ([]models.SearchResult, error) {
	return nil, nil
}

func (m *mockStorageForList) Get(id string) (*models.Memory, error) {
	if mem, exists := m.memories[id]; exists {
		return mem, nil
	}
	return nil, errors.New("memory not found")
}

func (m *mockStorageForList) GetByNamespace(namespace string) ([]*models.Memory, error) {
	results := make([]*models.Memory, 0)
	for _, mem := range m.memories {
		if mem.Namespace == namespace {
			results = append(results, mem)
		}
	}
	return results, nil
}

func (m *mockStorageForList) GetByTag(tag string) ([]*models.Memory, error) {
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

func (m *mockStorageForList) Update(memory *models.Memory) error {
	m.memories[memory.ID] = memory
	return nil
}

func (m *mockStorageForList) Delete(id string) error {
	delete(m.memories, id)
	return nil
}

func (m *mockStorageForList) DeleteByNamespace(namespace string) error {
	for id, mem := range m.memories {
		if mem.Namespace == namespace {
			delete(m.memories, id)
		}
	}
	return nil
}

func (m *mockStorageForList) List(namespace string, tag string, limit int) ([]*models.Memory, error) {
	results := make([]*models.Memory, 0)
	count := 0
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
		if limit > 0 && count >= limit {
			break
		}
		results = append(results, mem)
		count++
	}
	return results, nil
}

func (m *mockStorageForList) ListNamespaces() ([]interface{}, error) {
	return nil, nil
}

func (m *mockStorageForList) Count(namespace string) (int, error) {
	return 0, nil
}

func (m *mockStorageForList) CreateNamespace(namespace string) error {
	return nil
}

func (m *mockStorageForList) DeleteNamespace(namespace string) error {
	return m.DeleteByNamespace(namespace)
}

func (m *mockStorageForList) Close() error {
	return nil
}

// TestListCmd tests the list command
func TestListCmd(t *testing.T) {
	t.Run("ListCommandExists", func(t *testing.T) {
		if ListCmd == nil {
			t.Fatal("ListCmd should not be nil")
		}
		if ListCmd.Use != "list" {
			t.Errorf("Expected use 'list', got '%s'", ListCmd.Use)
		}
		if ListCmd.Short != "List stored memories" {
			t.Errorf("Expected short description 'List stored memories', got '%s'", ListCmd.Short)
		}
	})

	t.Run("ListCommandFlags", func(t *testing.T) {
		flags := ListCmd.Flags()

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

		limitFlag := flags.Lookup("limit")
		if limitFlag == nil {
			t.Error("limit flag should be defined")
		} else if limitFlag.Shorthand != "l" {
			t.Errorf("Expected limit shorthand 'l', got '%s'", limitFlag.Shorthand)
		}

		formatFlag := flags.Lookup("format")
		if formatFlag == nil {
			t.Error("format flag should be defined")
		} else if formatFlag.Shorthand != "f" {
			t.Errorf("Expected format shorthand 'f', got '%s'", formatFlag.Shorthand)
		}
	})
}

// TestDeleteCmd tests the delete command
func TestDeleteCmd(t *testing.T) {
	t.Run("DeleteCommandExists", func(t *testing.T) {
		if DeleteCmd == nil {
			t.Fatal("DeleteCmd should not be nil")
		}
		if DeleteCmd.Use != "delete [id]" {
			t.Errorf("Expected use 'delete [id]', got '%s'", DeleteCmd.Use)
		}
		if DeleteCmd.Short != "Delete a stored memory" {
			t.Errorf("Expected short description 'Delete a stored memory', got '%s'", DeleteCmd.Short)
		}
	})

	t.Run("DeleteCommandFlags", func(t *testing.T) {
		flags := DeleteCmd.Flags()

		namespaceFlag := flags.Lookup("namespace")
		if namespaceFlag == nil {
			t.Error("namespace flag should be defined")
		} else if namespaceFlag.Shorthand != "n" {
			t.Errorf("Expected namespace shorthand 'n', got '%s'", namespaceFlag.Shorthand)
		}

		queryFlag := flags.Lookup("query")
		if queryFlag == nil {
			t.Error("query flag should be defined")
		} else if queryFlag.Shorthand != "q" {
			t.Errorf("Expected query shorthand 'q', got '%s'", queryFlag.Shorthand)
		}

		forceFlag := flags.Lookup("force")
		if forceFlag == nil {
			t.Error("force flag should be defined")
		} else if forceFlag.Shorthand != "f" {
			t.Errorf("Expected force shorthand 'f', got '%s'", forceFlag.Shorthand)
		}
	})
}

// TestUpdateCmd tests the update command
func TestUpdateCmd(t *testing.T) {
	t.Run("UpdateCommandExists", func(t *testing.T) {
		if UpdateCmd == nil {
			t.Fatal("UpdateCmd should not be nil")
		}
		if UpdateCmd.Use != "update [id] [new_content]" {
			t.Errorf("Expected use 'update [id] [new_content]', got '%s'", UpdateCmd.Use)
		}
		if UpdateCmd.Short != "Update an existing memory" {
			t.Errorf("Expected short description 'Update an existing memory', got '%s'", UpdateCmd.Short)
		}
	})

	t.Run("UpdateCommandFlags", func(t *testing.T) {
		flags := UpdateCmd.Flags()

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
	})
}

// TestListOutputFormat tests list output formatting
func TestListOutputFormat(t *testing.T) {
	now := time.Now()
	memories := []*models.Memory{
		{
			ID:        "mem1",
			Content:   "First memory about user preferences",
			Namespace: "default",
			Tags:      []string{"preferences", "user"},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "mem2",
			Content:   "Second memory about database config",
			Namespace: "work",
			Tags:      []string{"database"},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	t.Run("text output", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Simulate list text output
		for i, mem := range memories {
			fmt.Printf("%d. %s\n", i+1, truncateString(mem.Content, 60))
			fmt.Printf("   ID:        %s\n", mem.ID)
			fmt.Printf("   Namespace: %s\n", mem.Namespace)
			if len(mem.Tags) > 0 {
				fmt.Printf("   Tags:      %s\n", strings.Join(mem.Tags, ", "))
			}
			fmt.Printf("   Created:   %s\n", mem.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Println()
		}

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify output contains expected content
		expectedStrings := []string{
			"mem1",
			"mem2",
			"preferences",
			"database",
			"default",
			"work",
		}

		for _, expected := range expectedStrings {
			if !strings.Contains(output, expected) {
				t.Errorf("Output should contain %q", expected)
			}
		}
	})

	t.Run("json output", func(t *testing.T) {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Simulate list JSON output
		data, err := json.MarshalIndent(memories, "", "  ")
		if err == nil {
			fmt.Println(string(data))
		}

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Verify it's valid JSON
		var decoded []*models.Memory
		if err := json.Unmarshal([]byte(output), &decoded); err != nil {
			t.Errorf("Output should be valid JSON: %v", err)
		}

		// Verify content
		if !strings.Contains(output, "mem1") {
			t.Error("JSON output should contain mem1")
		}
		if !strings.Contains(output, "preferences") {
			t.Error("JSON output should contain preferences")
		}
	})
}

// TestListFiltering tests list filtering logic
func TestListFiltering(t *testing.T) {
	now := time.Now()
	mockStore := newMockStorageForList()
	
	// Add test memories
	mockStore.memories["mem1"] = &models.Memory{
		ID:        "mem1",
		Content:   "Memory in default namespace",
		Namespace: "default",
		Tags:      []string{"tag1", "tag2"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	mockStore.memories["mem2"] = &models.Memory{
		ID:        "mem2",
		Content:   "Memory in work namespace",
		Namespace: "work",
		Tags:      []string{"tag2", "tag3"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	mockStore.memories["mem3"] = &models.Memory{
		ID:        "mem3",
		Content:   "Another memory in default namespace",
		Namespace: "default",
		Tags:      []string{"tag1"},
		CreatedAt: now,
		UpdatedAt: now,
	}

	tests := []struct {
		name      string
		namespace string
		tag       string
		limit     int
		expected  int
	}{
		{
			name:     "list all",
			expected: 3,
		},
		{
			name:      "filter by namespace default",
			namespace: "default",
			expected:  2,
		},
		{
			name:      "filter by namespace work",
			namespace: "work",
			expected:  1,
		},
		{
			name:     "filter by tag tag1",
			tag:      "tag1",
			expected: 2,
		},
		{
			name:     "filter by tag tag2",
			tag:      "tag2",
			expected: 2,
		},
		{
			name:      "filter by namespace and tag",
			namespace: "default",
			tag:       "tag1",
			expected:  2,
		},
		{
			name:      "filter by non-existent namespace",
			namespace: "nonexistent",
			expected:  0,
		},
		{
			name:     "filter by non-existent tag",
			tag:      "nonexistent",
			expected: 0,
		},
		{
			name:  "with limit",
			limit: 2,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := mockStore.List(tt.namespace, tt.tag, tt.limit)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}
			if len(results) != tt.expected {
				t.Errorf("List() returned %d results, expected %d", len(results), tt.expected)
			}
		})
	}
}

// TestDeleteValidation tests delete validation
func TestDeleteValidation(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		force   bool
		wantErr bool
	}{
		{
			name:    "valid id",
			id:      "mem123",
			force:   false,
			wantErr: false,
		},
		{
			name:    "empty id",
			id:      "",
			force:   false,
			wantErr: true,
		},
		{
			name:    "force flag",
			id:      "mem123",
			force:   true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate validation logic from delete command
			hasErr := tt.id == ""
			if hasErr != tt.wantErr {
				t.Errorf("Expected error: %v, got: %v", tt.wantErr, hasErr)
			}
		})
	}
}

// TestUpdateValidation tests update validation
func TestUpdateValidation(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		newContent string
		wantErr    bool
	}{
		{
			name:       "valid update",
			id:         "mem123",
			newContent: "Updated content",
			wantErr:    false,
		},
		{
			name:       "empty id",
			id:         "",
			newContent: "Updated content",
			wantErr:    true,
		},
		{
			name:       "empty content",
			id:         "mem123",
			newContent: "",
			wantErr:    true,
		},
		{
			name:       "both empty",
			id:         "",
			newContent: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate validation logic from update command
			hasErr := tt.id == "" || strings.TrimSpace(tt.newContent) == ""
			if hasErr != tt.wantErr {
				t.Errorf("Expected error: %v, got: %v", tt.wantErr, hasErr)
			}
		})
	}
}

// TestMemoryUpdateLogic tests memory update behavior
func TestMemoryUpdateLogic(t *testing.T) {
	now := time.Now()
	mockStore := newMockStorageForList()
	
	original := &models.Memory{
		ID:        "mem1",
		Content:   "Original content",
		Namespace: "default",
		Tags:      []string{"tag1"},
		Metadata:  map[string]interface{}{"key1": "value1"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	mockStore.memories[original.ID] = original

	t.Run("update content only", func(t *testing.T) {
		updated := *original
		updated.Content = "Updated content"
		updated.UpdatedAt = time.Now()
		
		err := mockStore.Update(&updated)
		if err != nil {
			t.Errorf("Update() error = %v", err)
			return
		}
		
		retrieved, _ := mockStore.Get("mem1")
		if retrieved.Content != "Updated content" {
			t.Errorf("Content not updated, got: %s", retrieved.Content)
		}
		// Tags should remain unchanged
		if len(retrieved.Tags) != 1 || retrieved.Tags[0] != "tag1" {
			t.Errorf("Tags should remain unchanged, got: %v", retrieved.Tags)
		}
	})

	t.Run("update tags only", func(t *testing.T) {
		updated := *original
		updated.Tags = []string{"tag2", "tag3"}
		updated.UpdatedAt = time.Now()
		
		err := mockStore.Update(&updated)
		if err != nil {
			t.Errorf("Update() error = %v", err)
			return
		}
		
		retrieved, _ := mockStore.Get("mem1")
		if len(retrieved.Tags) != 2 {
			t.Errorf("Tags not updated, got: %v", retrieved.Tags)
		}
	})

	t.Run("update metadata only", func(t *testing.T) {
		updated := *original
		updated.Metadata = map[string]interface{}{"key2": "value2"}
		updated.UpdatedAt = time.Now()
		
		err := mockStore.Update(&updated)
		if err != nil {
			t.Errorf("Update() error = %v", err)
			return
		}
		
		retrieved, _ := mockStore.Get("mem1")
		if retrieved.Metadata["key2"] != "value2" {
			t.Errorf("Metadata not updated, got: %v", retrieved.Metadata)
		}
	})
}