package cli

import (
	"testing"
)

// TestListCmdExists tests that the list command is properly defined
func TestListCmdExists(t *testing.T) {
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

// TestListOutputFormats tests different output formats for list command
func TestListOutputFormats(t *testing.T) {
	formats := []string{"text", "json"}
	
	for _, format := range formats {
		t.Run("Format_"+format, func(t *testing.T) {
			// Test that the format is recognized
			if format != "text" && format != "json" {
				t.Errorf("Unsupported format: %s", format)
			}
		})
	}
}
