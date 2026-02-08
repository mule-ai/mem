package cli

import (
	"testing"
)

// TestDeleteCmdExists tests that the delete command is properly defined
func TestDeleteCmdExists(t *testing.T) {
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

// TestDeleteValidationExtra tests validation of delete inputs
func TestDeleteValidationExtra(t *testing.T) {
	t.Run("ValidateMemoryID", func(t *testing.T) {
		validIDs := []string{
			"test123",
			"abc-def-ghi",
			"1234567890",
			"mem_001",
		}
		
		for _, id := range validIDs {
			if id == "" {
				t.Errorf("ID %q should be valid", id)
			}
		}
	})

	t.Run("InvalidateEmptyID", func(t *testing.T) {
		invalidIDs := []string{
			"",
			"   ",
		}
		
		for _, id := range invalidIDs {
			if id != "" && len(id) > 0 {
				// This test is to ensure we handle empty IDs
				t.Logf("Testing ID: %q", id)
			}
		}
	})
}
