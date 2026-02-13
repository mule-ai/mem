package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jbutlerdev/mem/internal/models"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
)

var (
	listNamespace string
	listTag       string
	listLimit     int
	listOutput    string
)

// ListCmd represents the list command
var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all stored memories",
	Long: `List stored memories with optional filtering by namespace and tags.

This shows all memories without semantic ranking - useful for browsing
and managing your stored memories.

USAGE:
  mem list

FLAGS:
  -n, --namespace <name>    List only memories in this namespace
  -t, --tag <tag>           Filter by tag
  -l, --limit <num>         Maximum number to show (0 = unlimited)
  -f, --format <format>     Output format: text, json, yaml (default: text)

EXAMPLES:
  # List all memories
  mem list

  # List in specific namespace
  mem list -n project-alpha

  # Filter by tag
  mem list --tag important

  # JSON output for processing
  mem list -f json | jq '.[] | select(.content | contains("deadline"))'

  # Limit results
  mem list --limit 20`,
	Args: cobra.NoArgs,
	RunE: runList,
}

func init() {
	ListCmd.Flags().StringVarP(&listNamespace, "namespace", "n", "", "namespace to list")
	ListCmd.Flags().StringVarP(&listTag, "tag", "t", "", "filter by tag")
	ListCmd.Flags().IntVarP(&listLimit, "limit", "l", 0, "maximum number to show")
	ListCmd.Flags().StringVarP(&listOutput, "format", "f", "", "output format: text | json | yaml")
}

func runList(cmd *cobra.Command, args []string) error {
	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine namespace
	namespace := listNamespace
	if namespace == "" {
		namespace = cfg.Memory.DefaultNamespace
	}
	// If namespace is still empty, don't default to "default"
	// Leave it empty to list all namespaces

	// Determine output format
	outputFormat := listOutput
	if outputFormat == "" {
		outputFormat = cfg.CLI.OutputFormat
	}
	if outputFormat == "" {
		outputFormat = "text"
	}

	// Initialize storage backend
	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Apply limit
	limit := listLimit
	if limit == 0 {
		// Default to showing all if no limit specified
		limit = 0 // 0 means no limit with new signature
	}

	if cfg.CLI.Verbose {
		if namespace != "" {
			fmt.Fprintf(os.Stderr, "Listing memories in namespace '%s'", namespace)
			if listTag != "" {
				fmt.Fprintf(os.Stderr, " with tag '%s'", listTag)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Listing memories in all namespaces")
			if listTag != "" {
				fmt.Fprintf(os.Stderr, " with tag '%s'", listTag)
			}
		}
		fmt.Fprintf(os.Stderr, "...\n")
	}

	// Execute list using new signature
	memories, err := store.List(namespace, listTag, limit)
	if err != nil {
		return fmt.Errorf("failed to list memories: %w", err)
	}

	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d memory(s)\n", len(memories))
	}

	// Output results
	switch outputFormat {
	case "json":
		outputListResultsJSON(memories)
	case "yaml":
		outputListResultsYAML(memories)
	default:
		outputListResultsText(memories)
	}

	return nil
}

func outputListResultsText(memories []*models.Memory) {
	if len(memories) == 0 {
		fmt.Println("No memories found.")
		return
	}

	fmt.Printf("Found %d memory(s):\n\n", len(memories))
	for i, memory := range memories {
		fmt.Printf("%d. [%s] %s\n", i+1, memory.Namespace, truncateString(memory.Content, 70))
		fmt.Printf("   ID:      %s\n", memory.ID)
		fmt.Printf("   Created: %s\n", memory.CreatedAt.Format("2006-01-02 15:04:05"))
		if len(memory.Tags) > 0 {
			fmt.Printf("   Tags:    %s\n", strings.Join(memory.Tags, ", "))
		}
		if memory.Metadata != nil && len(memory.Metadata) > 0 {
			keys := make([]string, 0, len(memory.Metadata))
			for k := range memory.Metadata {
				keys = append(keys, k)
			}
			fmt.Printf("   Metadata: %s\n", strings.Join(keys, ", "))
		}
		fmt.Println()
	}
}

func outputListResultsJSON(memories []*models.Memory) {
	data, err := json.MarshalIndent(memories, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func outputListResultsYAML(memories []*models.Memory) {
	data, err := yaml.Marshal(memories)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling YAML: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
