package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// NamespaceCmd represents the namespace command group
var NamespaceCmd = &cobra.Command{
	Use:   "namespace [command]",
	Short: "Manage memory namespaces",
	Long: `Manage memory namespaces for organizing memories into separate contexts.

Namespaces allow you to group memories by project, client, or any other
category. They're automatically created when you store memories with a
namespace flag.

USAGE:
  mem namespace list
  mem namespace create <name>
  mem namespace delete <name>

EXAMPLES:
  # List all namespaces
  mem namespace list

  # Create a namespace (optional - namespaces auto-create)
  mem namespace create work

  # Delete a namespace and all its memories
  mem namespace delete old-project`,
}

var (
	namespaceListVerbose bool
)

// NamespaceListCmd represents the namespace list command
var NamespaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all namespaces",
	Long: `List all namespaces that contain memories, along with memory counts.

USAGE:
  mem namespace list

FLAGS:
  -v, --verbose    Show detailed information including creation date

EXAMPLES:
  mem namespace list
  mem namespace list --verbose`,
	RunE: runNamespaceList,
}

func init() {
	NamespaceListCmd.Flags().BoolVarP(&namespaceListVerbose, "verbose", "v", false, "show detailed information")
}

func runNamespaceList(cmd *cobra.Command, args []string) error {
	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize storage backend
	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Get list of namespaces
	namespaces, err := store.ListNamespaces()
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %w", err)
	}

	if len(namespaces) == 0 {
		fmt.Println("No namespaces found")
		return nil
	}

	// Display namespaces
	outputFormat := cfg.CLI.OutputFormat
	if outputFormat == "" {
		outputFormat = "text"
	}

	switch outputFormat {
	case "text":
		fmt.Printf("Namespaces (%d):\n\n", len(namespaces))
		for _, ns := range namespaces {
			count, err := store.Count(ns.Name)
			if err != nil {
				fmt.Printf("  %s (error: %v)\n", ns.Name, err)
				continue
			}
			fmt.Printf("  • %s (%d memories)\n", ns.Name, count)
			if namespaceListVerbose {
				if ct, ok := ns.CreatedAt.(time.Time); ok {
					fmt.Printf("      Created: %s\n", ct.Format("2006-01-02 15:04:05"))
				}
			}
		}
	default:
		return fmt.Errorf("output format '%s' not yet implemented for namespace list", outputFormat)
	}

	return nil
}

// NamespaceCreateCmd represents the namespace create command
var NamespaceCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new namespace",
	Long: `Create a new namespace for organizing memories.

Note: Namespaces are automatically created when storing memories with
the -n flag, so this command is optional and mainly for explicit
initialization.

USAGE:
  mem namespace create <name>

EXAMPLES:
  mem namespace create work
  mem namespace create personal`,
	Args: cobra.ExactArgs(1),
	RunE: runNamespaceCreate,
}

func runNamespaceCreate(cmd *cobra.Command, args []string) error {
	namespaceName := args[0]

	// Validate namespace name
	if namespaceName == "" {
		return fmt.Errorf("namespace name cannot be empty")
	}

	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize storage backend
	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Check if namespace already exists
	namespaces, err := store.ListNamespaces()
	if err != nil {
		return fmt.Errorf("failed to check existing namespaces: %w", err)
	}

	for _, ns := range namespaces {
		if ns.Name == namespaceName {
			fmt.Printf("Namespace '%s' already exists\n", namespaceName)
			return nil
		}
	}

	// Create namespace (by storing a marker memory or using backend-specific method)
	if err := store.CreateNamespace(namespaceName); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	fmt.Printf("✓ Namespace '%s' created successfully\n", namespaceName)
	return nil
}

var (
	namespaceDeleteForce bool
	namespaceDeletePrompt bool
)

// NamespaceDeleteCmd represents the namespace delete command
var NamespaceDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a namespace and all its memories",
	Long: `Delete a namespace and all memories contained within it.

WARNING: This operation is irreversible and permanently deletes all
memories in the namespace. You will be prompted for confirmation unless
--force is used.

USAGE:
  mem namespace delete <name>

FLAGS:
  -f, --force    Skip confirmation prompt

EXAMPLES:
  mem namespace delete old-project
  mem namespace delete --force temp-ns`,
	Args: cobra.ExactArgs(1),
	RunE: runNamespaceDelete,
}

func init() {
	NamespaceDeleteCmd.Flags().BoolVarP(&namespaceDeleteForce, "force", "f", false, "skip confirmation prompt")
}

func runNamespaceDelete(cmd *cobra.Command, args []string) error {
	namespaceName := args[0]

	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize storage backend
	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Check if namespace exists
	namespaces, err := store.ListNamespaces()
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaceExists := false
	for _, ns := range namespaces {
		if ns.Name == namespaceName {
			namespaceExists = true
			break
		}
	}

	if !namespaceExists {
		return fmt.Errorf("namespace '%s' does not exist", namespaceName)
	}

	// Get memory count
	count, err := store.Count(namespaceName)
	if err != nil {
		return fmt.Errorf("failed to get memory count: %w", err)
	}

	// Confirm deletion
	if !namespaceDeleteForce {
		fmt.Printf("Are you sure you want to delete namespace '%s' and all %d memories in it? [y/N]: ", namespaceName, count)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	// Delete namespace
	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Deleting namespace '%s'...\n", namespaceName)
	}

	if err := store.DeleteNamespace(namespaceName); err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	fmt.Printf("✓ Namespace '%s' deleted successfully\n", namespaceName)
	return nil
}

func init() {
	NamespaceCmd.AddCommand(NamespaceListCmd)
	NamespaceCmd.AddCommand(NamespaceCreateCmd)
	NamespaceCmd.AddCommand(NamespaceDeleteCmd)
}
