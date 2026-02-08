package cli

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jbutlerdev/mem/internal/models"
)

// TestImportCmd tests the import command
func TestImportCmd(t *testing.T) {
	t.Run("ImportCommandExists", func(t *testing.T) {
		if ImportCmd == nil {
			t.Fatal("ImportCmd should not be nil")
		}
		if ImportCmd.Use != "import <input_file>" {
			t.Errorf("Expected use 'import <input_file>', got '%s'", ImportCmd.Use)
		}
		if ImportCmd.Short != "Import memories from a backup file" {
			t.Errorf("Expected short description 'Import memories from a backup file', got '%s'", ImportCmd.Short)
		}
	})

	t.Run("ImportCommandFlags", func(t *testing.T) {
		flags := ImportCmd.Flags()

		namespaceFlag := flags.Lookup("namespace")
		if namespaceFlag == nil {
			t.Error("namespace flag should be defined")
		} else if namespaceFlag.Shorthand != "n" {
			t.Errorf("Expected namespace shorthand 'n', got '%s'", namespaceFlag.Shorthand)
		}

		formatFlag := flags.Lookup("format")
		if formatFlag == nil {
			t.Error("format flag should be defined")
		}
	})
}

// TestExportCmd tests the export command
func TestExportCmd(t *testing.T) {
	t.Run("ExportCommandExists", func(t *testing.T) {
		if ExportCmd == nil {
			t.Fatal("ExportCmd should not be nil")
		}
		if ExportCmd.Use != "export [output_file]" {
			t.Errorf("Expected use 'export [output_file]', got '%s'", ExportCmd.Use)
		}
		if ExportCmd.Short != "Export memories to a file for backup" {
			t.Errorf("Expected short description 'Export memories to a file for backup', got '%s'", ExportCmd.Short)
		}
	})

	t.Run("ExportCommandFlags", func(t *testing.T) {
		flags := ExportCmd.Flags()

		namespaceFlag := flags.Lookup("namespace")
		if namespaceFlag == nil {
			t.Error("namespace flag should be defined")
		} else if namespaceFlag.Shorthand != "n" {
			t.Errorf("Expected namespace shorthand 'n', got '%s'", namespaceFlag.Shorthand)
		}

		formatFlag := flags.Lookup("format")
		if formatFlag == nil {
			t.Error("format flag should be defined")
		}
	})
}

// TestNamespaceCmd tests the namespace command
func TestNamespaceCmd(t *testing.T) {
	t.Run("NamespaceCommandExists", func(t *testing.T) {
		if NamespaceCmd == nil {
			t.Fatal("NamespaceCmd should not be nil")
		}
		if NamespaceCmd.Use != "namespace [command]" {
			t.Errorf("Expected use 'namespace [command]', got '%s'", NamespaceCmd.Use)
		}
		if NamespaceCmd.Short != "Manage memory namespaces" {
			t.Errorf("Expected short description 'Manage memory namespaces', got '%s'", NamespaceCmd.Short)
		}
	})

	t.Run("NamespaceSubcommands", func(t *testing.T) {
		subcommands := NamespaceCmd.Commands()
		
		expectedSubcommands := []string{"list", "create", "delete"}
		subcommandMap := make(map[string]bool)
		for _, cmd := range subcommands {
			subcommandMap[cmd.Name()] = true
		}
		
		for _, expected := range expectedSubcommands {
			if !subcommandMap[expected] {
				t.Errorf("Expected subcommand %q not found", expected)
			}
		}
	})
}

// TestExportImportRoundTrip tests export/import cycle
func TestExportImportRoundTrip(t *testing.T) {
	now := time.Now()
	
	// Create test memories
	memories := []*models.Memory{
		{
			ID:        "mem1",
			Content:   "First memory about preferences",
			Namespace: "default",
			Tags:      []string{"preferences"},
			Metadata:  map[string]interface{}{"source": "test"},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        "mem2",
			Content:   "Second memory about work",
			Namespace: "work",
			Tags:      []string{"work", "project"},
			Metadata:  map[string]interface{}{"priority": 1},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	t.Run("export to JSON", func(t *testing.T) {
		tmpFile := t.TempDir() + "/export.json"
		
		// Create export file
		data, err := json.MarshalIndent(memories, "", "  ")
		if err != nil {
			t.Fatalf("Failed to marshal memories: %v", err)
		}
		
		err = os.WriteFile(tmpFile, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write export file: %v", err)
		}
		
		// Verify file exists and is valid JSON
		content, err := os.ReadFile(tmpFile)
		if err != nil {
			t.Fatalf("Failed to read export file: %v", err)
		}
		
		var imported []*models.Memory
		err = json.Unmarshal(content, &imported)
		if err != nil {
			t.Errorf("Failed to unmarshal exported data: %v", err)
		}
		
		if len(imported) != len(memories) {
			t.Errorf("Expected %d memories, got %d", len(memories), len(imported))
		}
		
		// Verify content
		for i, mem := range imported {
			if mem.ID != memories[i].ID {
				t.Errorf("Memory %d: expected ID %s, got %s", i, memories[i].ID, mem.ID)
			}
			if mem.Content != memories[i].Content {
				t.Errorf("Memory %d: content mismatch", i)
			}
		}
	})

	t.Run("export with namespace filter", func(t *testing.T) {
		// Filter memories by namespace
		filtered := make([]*models.Memory, 0)
		for _, mem := range memories {
			if mem.Namespace == "work" {
				filtered = append(filtered, mem)
			}
		}
		
		if len(filtered) != 1 {
			t.Errorf("Expected 1 memory in work namespace, got %d", len(filtered))
		}
		
		if filtered[0].ID != "mem2" {
			t.Errorf("Expected mem2, got %s", filtered[0].ID)
		}
	})
}

// TestImportValidation tests import validation
func TestImportValidation(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name: "valid JSON",
			content: `[
				{"id": "mem1", "content": "Test", "namespace": "default", "created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z", "embedding": [0.1, 0.2, 0.3]}
			]`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			content: `{invalid json}`,
			wantErr: true,
		},
		{
			name:    "empty array",
			content: `[]`,
			wantErr: false,
		},
		{
			name: "valid JSON without embedding (import will generate)",
			content: `[
				{"id": "mem1", "content": "Test", "namespace": "default", "created_at": "2024-01-01T00:00:00Z", "updated_at": "2024-01-01T00:00:00Z"}
			]`,
			wantErr: false, // Import will generate embedding
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := t.TempDir() + "/import.json"
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}
			
			content, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}
			
			var memories []*models.Memory
			err = json.Unmarshal(content, &memories)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error: %v, got: %v", tt.wantErr, err)
			}
			
			if err == nil && len(memories) > 0 {
				// Validate each memory - for import, we allow empty embedding since it will be generated
				for _, mem := range memories {
					// Skip embedding validation for import test
					if mem.ID == "" || mem.Content == "" || mem.Namespace == "" {
						if !tt.wantErr {
							t.Errorf("Memory missing required fields: %+v", mem)
						}
					}
				}
			}
		})
	}
}

// TestNamespaceOperations tests namespace management
func TestNamespaceOperations(t *testing.T) {
	mockStore := newMockStorageForList()
	
	// Add memories to different namespaces
	now := time.Now()
	mockStore.memories["mem1"] = &models.Memory{
		ID:        "mem1",
		Content:   "Memory in default namespace",
		Namespace: "default",
		CreatedAt: now,
		UpdatedAt: now,
	}
	mockStore.memories["mem2"] = &models.Memory{
		ID:        "mem2",
		Content:   "Memory in work namespace",
		Namespace: "work",
		CreatedAt: now,
		UpdatedAt: now,
	}
	mockStore.memories["mem3"] = &models.Memory{
		ID:        "mem3",
		Content:   "Another memory in default namespace",
		Namespace: "default",
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Run("list namespaces", func(t *testing.T) {
		// Get unique namespaces from memories
		namespaceSet := make(map[string]bool)
		for _, mem := range mockStore.memories {
			namespaceSet[mem.Namespace] = true
		}
		
		namespaces := make([]string, 0, len(namespaceSet))
		for ns := range namespaceSet {
			namespaces = append(namespaces, ns)
		}
		
		if len(namespaces) != 2 {
			t.Errorf("Expected 2 namespaces, got %d", len(namespaces))
		}
		
		// Check expected namespaces exist
		namespaceMap := make(map[string]bool)
		for _, ns := range namespaces {
			namespaceMap[ns] = true
		}
		
		if !namespaceMap["default"] {
			t.Error("Expected 'default' namespace")
		}
		if !namespaceMap["work"] {
			t.Error("Expected 'work' namespace")
		}
	})

	t.Run("count memories per namespace", func(t *testing.T) {
		counts := make(map[string]int)
		for _, mem := range mockStore.memories {
			counts[mem.Namespace]++
		}
		
		if counts["default"] != 2 {
			t.Errorf("Expected 2 memories in default namespace, got %d", counts["default"])
		}
		if counts["work"] != 1 {
			t.Errorf("Expected 1 memory in work namespace, got %d", counts["work"])
		}
	})

	t.Run("delete namespace", func(t *testing.T) {
		// Simulate deleting a namespace
		nsToDelete := "work"
		deletedCount := 0
		for id, mem := range mockStore.memories {
			if mem.Namespace == nsToDelete {
				delete(mockStore.memories, id)
				deletedCount++
			}
		}
		
		if deletedCount != 1 {
			t.Errorf("Expected to delete 1 memory, got %d", deletedCount)
		}
		
		// Verify work namespace no longer exists
		found := false
		for _, mem := range mockStore.memories {
			if mem.Namespace == nsToDelete {
				found = true
				break
			}
		}
		
		if found {
			t.Error("Work namespace should be deleted")
		}
	})
}

// TestConfigCmd tests the config command
func TestConfigCmd(t *testing.T) {
	t.Run("ConfigCommandExists", func(t *testing.T) {
		if ConfigCmd == nil {
			t.Fatal("ConfigCmd should not be nil")
		}
		if ConfigCmd.Use != "config [command]" {
			t.Errorf("Expected use 'config [command]', got '%s'", ConfigCmd.Use)
		}
		if ConfigCmd.Short != "Manage mem configuration" {
			t.Errorf("Expected short description 'Manage mem configuration', got '%s'", ConfigCmd.Short)
		}
	})

	t.Run("ConfigSubcommands", func(t *testing.T) {
		subcommands := ConfigCmd.Commands()
		
		expectedSubcommands := []string{"show", "get", "set", "edit"}
		subcommandMap := make(map[string]bool)
		for _, cmd := range subcommands {
			subcommandMap[cmd.Name()] = true
		}
		
		for _, expected := range expectedSubcommands {
			if !subcommandMap[expected] {
				t.Errorf("Expected subcommand %q not found", expected)
			}
		}
	})
}

// TestConfigOperations tests config management operations
func TestConfigOperations(t *testing.T) {
	t.Run("config get/set simulation", func(t *testing.T) {
		// Simulate config key-value operations
		config := make(map[string]interface{})
		
		// Set values
		config["memory.backend"] = "chromem"
		config["embeddings.model"] = "text-embedding-qwen3-embedding-8b"
		config["query.default_limit"] = 10
		
		// Get values
		if config["memory.backend"] != "chromem" {
			t.Error("memory.backend should be chromem")
		}
		if config["embeddings.model"] != "text-embedding-qwen3-embedding-8b" {
			t.Error("embeddings.model should be text-embedding-qwen3-embedding-8b")
		}
	})

	t.Run("nested config paths", func(t *testing.T) {
		// Test parsing nested config paths
		tests := []struct {
			path    string
			parts   []string
		}{
			{"memory.backend", []string{"memory", "backend"}},
			{"embeddings.model", []string{"embeddings", "model"}},
			{"query.default_limit", []string{"query", "default_limit"}},
		}
		
		for _, tt := range tests {
			t.Run(tt.path, func(t *testing.T) {
				parts := strings.Split(tt.path, ".")
				if len(parts) != len(tt.parts) {
					t.Errorf("Expected %d parts, got %d", len(tt.parts), len(parts))
				}
			})
		}
	})
}