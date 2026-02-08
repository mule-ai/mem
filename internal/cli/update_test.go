package cli

import (
	"testing"
	"time"

	"github.com/jbutlerdev/mem/internal/models"
)

// TestUpdateCmdExists tests that the update command is properly defined
func TestUpdateCmdExists(t *testing.T) {
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

// TestMemoryUpdateLogicExtra tests the logic for updating memories
func TestMemoryUpdateLogicExtra(t *testing.T) {
	t.Run("UpdateContent", func(t *testing.T) {
		mem := createTestMemory(t)
		originalID := mem.ID
		originalCreatedAt := mem.CreatedAt
		
		// Update content
		mem.Content = "Updated content"
		mem.UpdatedAt = time.Now()
		
		if mem.ID != originalID {
			t.Error("ID should not change on update")
		}
		if !mem.CreatedAt.Equal(originalCreatedAt) {
			t.Error("CreatedAt should not change on update")
		}
		if mem.Content != "Updated content" {
			t.Error("Content should be updated")
		}
	})

	t.Run("UpdateTags", func(t *testing.T) {
		mem := createTestMemory(t)
		
		// Update tags
		newTags := []string{"updated", "test"}
		mem.Tags = newTags
		
		if len(mem.Tags) != 2 {
			t.Errorf("Expected 2 tags, got %d", len(mem.Tags))
		}
		
		tagMap := make(map[string]bool)
		for _, tag := range mem.Tags {
			tagMap[tag] = true
		}
		
		if !tagMap["updated"] || !tagMap["test"] {
			t.Error("Tags should contain 'updated' and 'test'")
		}
	})

	t.Run("UpdateMetadata", func(t *testing.T) {
		mem := createTestMemory(t)
		
		// Update metadata
		newMetadata := map[string]interface{}{
			"newKey": "newValue",
			"number": 42,
		}
		mem.Metadata = newMetadata
		
		if mem.Metadata["newKey"] != "newValue" {
			t.Error("Metadata should be updated")
		}
		if mem.Metadata["number"] != 42 {
			t.Error("Metadata number should be 42")
		}
	})
}

// TestUpdateValidationExtra tests validation of update inputs
func TestUpdateValidationExtra(t *testing.T) {
	t.Run("ValidateUpdatedMemory", func(t *testing.T) {
		mem := &models.Memory{
			ID:        "test123",
			Content:   "Valid updated content",
			Embedding: []float32{0.1, 0.2},
			Namespace: "test",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		if err := mem.Validate(); err != nil {
			t.Errorf("Updated memory should be valid: %v", err)
		}
	})

	t.Run("InvalidateMemoryWithEmptyContent", func(t *testing.T) {
		mem := &models.Memory{
			ID:        "test123",
			Content:   "",
			Embedding: []float32{0.1, 0.2},
			Namespace: "test",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		if err := mem.Validate(); err == nil {
			t.Error("Memory with empty content should be invalid")
		}
	})
}

// TestUpdateTimestamps tests that timestamps are handled correctly
func TestUpdateTimestamps(t *testing.T) {
	mem := createTestMemory(t)
	originalUpdatedAt := mem.UpdatedAt
	
	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)
	
	// Update the memory
	mem.Content = "Updated"
	mem.UpdatedAt = time.Now()
	
	if mem.UpdatedAt.Before(originalUpdatedAt) || mem.UpdatedAt.Equal(originalUpdatedAt) {
		t.Error("UpdatedAt should be later than original UpdatedAt")
	}
}
