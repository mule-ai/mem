package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestConfigCmdExists tests that the config command is properly defined
func TestConfigCmdExists(t *testing.T) {
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

	t.Run("ConfigHasSubcommands", func(t *testing.T) {
		subcommands := ConfigCmd.Commands()
		
		if len(subcommands) == 0 {
			t.Error("ConfigCmd should have subcommands")
		}

		subcommandMap := make(map[string]bool)
		for _, cmd := range subcommands {
			subcommandMap[cmd.Name()] = true
		}

		expectedSubcommands := []string{"show", "get", "set", "edit"}
		for _, expected := range expectedSubcommands {
			if !subcommandMap[expected] {
				t.Errorf("ConfigCmd should have '%s' subcommand", expected)
			}
		}
	})
}

// TestConfigShowCmd tests the config show subcommand
func TestConfigShowCmd(t *testing.T) {
	t.Run("ShowSubcommandExists", func(t *testing.T) {
		subcommands := ConfigCmd.Commands()
		
		var showCmd *cobra.Command
		for _, cmd := range subcommands {
			if cmd.Name() == "show" {
				showCmd = cmd
				break
			}
		}
		
		if showCmd == nil {
			t.Fatal("config show subcommand should exist")
		}
		
		if showCmd.Use != "show" {
			t.Errorf("Expected use 'show', got '%s'", showCmd.Use)
		}
		if showCmd.Short != "Show current effective configuration" {
			t.Errorf("Expected short 'Show current effective configuration', got '%s'", showCmd.Short)
		}
	})
}

// TestConfigGetCmd tests the config get subcommand
func TestConfigGetCmd(t *testing.T) {
	t.Run("GetSubcommandExists", func(t *testing.T) {
		subcommands := ConfigCmd.Commands()
		
		var getCmd *cobra.Command
		for _, cmd := range subcommands {
			if cmd.Name() == "get" {
				getCmd = cmd
				break
			}
		}
		
		if getCmd == nil {
			t.Fatal("config get subcommand should exist")
		}
		
		if getCmd.Use != "get <key>" {
			t.Errorf("Expected use 'get <key>', got '%s'", getCmd.Use)
		}
	})
}

// TestConfigSetCmd tests the config set subcommand
func TestConfigSetCmd(t *testing.T) {
	t.Run("SetSubcommandExists", func(t *testing.T) {
		subcommands := ConfigCmd.Commands()
		
		var setCmd *cobra.Command
		for _, cmd := range subcommands {
			if cmd.Name() == "set" {
				setCmd = cmd
				break
			}
		}
		
		if setCmd == nil {
			t.Fatal("config set subcommand should exist")
		}
		
		if setCmd.Use != "set <key> <value>" {
			t.Errorf("Expected use 'set <key> <value>', got '%s'", setCmd.Use)
		}
	})
}

// TestConfigEditCmd tests the config edit subcommand
func TestConfigEditCmd(t *testing.T) {
	t.Run("EditSubcommandExists", func(t *testing.T) {
		subcommands := ConfigCmd.Commands()
		
		var editCmd *cobra.Command
		for _, cmd := range subcommands {
			if cmd.Name() == "edit" {
				editCmd = cmd
				break
			}
		}
		
		if editCmd == nil {
			t.Fatal("config edit subcommand should exist")
		}
		
		if editCmd.Use != "edit" {
			t.Errorf("Expected use 'edit', got '%s'", editCmd.Use)
		}
	})
}