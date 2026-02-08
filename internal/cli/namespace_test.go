package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestNamespaceCmdExists tests that the namespace command is properly defined
func TestNamespaceCmdExists(t *testing.T) {
	t.Run("NamespaceCommandExists", func(t *testing.T) {
		if NamespaceCmd == nil {
			t.Fatal("NamespaceCmd should not be nil")
		}
		if NamespaceCmd.Use != "namespace [command]" {
			t.Errorf("Expected use 'namespace [command]', got '%s'", NamespaceCmd.Use)
		}
		if NamespaceCmd.Short != "Manage namespaces" {
			t.Errorf("Expected short description 'Manage namespaces', got '%s'", NamespaceCmd.Short)
		}
	})

	t.Run("NamespaceHasSubcommands", func(t *testing.T) {
		subcommands := NamespaceCmd.Commands()
		
		if len(subcommands) == 0 {
			t.Error("NamespaceCmd should have subcommands")
		}

		subcommandMap := make(map[string]bool)
		for _, cmd := range subcommands {
			subcommandMap[cmd.Name()] = true
		}

		expectedSubcommands := []string{"list", "create", "delete"}
		for _, expected := range expectedSubcommands {
			if !subcommandMap[expected] {
				t.Errorf("NamespaceCmd should have '%s' subcommand", expected)
			}
		}
	})
}

// TestNamespaceListCmd tests the namespace list subcommand
func TestNamespaceListCmd(t *testing.T) {
	t.Run("ListSubcommandExists", func(t *testing.T) {
		subcommands := NamespaceCmd.Commands()
		
		var listCmd *cobra.Command
		for _, cmd := range subcommands {
			if cmd.Name() == "list" {
				listCmd = cmd
				break
			}
		}
		
		if listCmd == nil {
			t.Fatal("namespace list subcommand should exist")
		}
		
		if listCmd.Use != "list" {
			t.Errorf("Expected use 'list', got '%s'", listCmd.Use)
		}
	})
}

// TestNamespaceCreateCmd tests the namespace create subcommand
func TestNamespaceCreateCmd(t *testing.T) {
	t.Run("CreateSubcommandExists", func(t *testing.T) {
		subcommands := NamespaceCmd.Commands()
		
		var createCmd *cobra.Command
		for _, cmd := range subcommands {
			if cmd.Name() == "create" {
				createCmd = cmd
				break
			}
		}
		
		if createCmd == nil {
			t.Fatal("namespace create subcommand should exist")
		}
		
		if createCmd.Use != "create <name>" {
			t.Errorf("Expected use 'create <name>', got '%s'", createCmd.Use)
		}
	})
}

// TestNamespaceDeleteCmd tests the namespace delete subcommand
func TestNamespaceDeleteCmd(t *testing.T) {
	t.Run("DeleteSubcommandExists", func(t *testing.T) {
		subcommands := NamespaceCmd.Commands()
		
		var deleteCmd *cobra.Command
		for _, cmd := range subcommands {
			if cmd.Name() == "delete" {
				deleteCmd = cmd
				break
			}
		}
		
		if deleteCmd == nil {
			t.Fatal("namespace delete subcommand should exist")
		}
		
		if deleteCmd.Use != "delete <name>" {
			t.Errorf("Expected use 'delete <name>', got '%s'", deleteCmd.Use)
		}
	})
}