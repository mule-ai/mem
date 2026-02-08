package cli

import (
	"testing"
)

// TestExportCmdExists tests that the export command is properly defined
func TestExportCmdExists(t *testing.T) {
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

// TestImportCmdExists tests that the import command is properly defined
func TestImportCmdExists(t *testing.T) {
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