package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/jbutlerdev/mem/internal/models"
	"github.com/spf13/cobra"
)

var (
	exportNamespace string
	exportFormat    string
	exportFile      string
)

// ExportCmd represents the export command
var ExportCmd = &cobra.Command{
	Use:   "export [output_file]",
	Short: "Export memories to a file for backup",
	Long: `Export memories to a file for backup or migration.

Exported files can be imported later using 'mem import'. Useful for backups,
migration between storage backends, or sharing memory collections.

USAGE:
  mem export [filename.json]
  mem export -n namespace [filename.json]

FLAGS:
  -n, --namespace <name>    Export only this namespace
      --format <fmt>       Output format: json or yaml (default: json)

EXAMPLES:
  # Export all memories to file
  mem export backup.json

  # Export specific namespace
  mem export -n project-alpha project-backup.json

  # Export to stdout
  mem export | jq '.memories[] | .content'

  # Export as YAML
  mem export --format yaml backup.yaml`,
	Args: cobra.MaximumNArgs(1),
	RunE: runExport,
}

func init() {
	ExportCmd.Flags().StringVarP(&exportNamespace, "namespace", "n", "", "namespace to export")
	ExportCmd.Flags().StringVar(&exportFormat, "format", "json", "output format: json | yaml")
}

func runExport(cmd *cobra.Command, args []string) error {
	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine output file
	var outputFile *os.File
	if len(args) == 1 {
		exportFile = args[0]
		file, err := os.Create(exportFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		outputFile = file
	} else {
		outputFile = os.Stdout
	}

	// Initialize storage backend
	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Determine namespace
	namespace := exportNamespace
	if namespace == "" {
		namespace = cfg.Memory.DefaultNamespace
	}

	// Retrieve memories
	var memories []*models.Memory
	if namespace != "" {
		// Export specific namespace
		if cfg.CLI.Verbose {
			fmt.Fprintf(os.Stderr, "Exporting memories from namespace '%s'...\n", namespace)
		}
		memories, err = store.List(namespace, "", 0) // No tag filter, no limit
		if err != nil {
			return fmt.Errorf("failed to retrieve memories: %w", err)
		}
	} else {
		// Export all memories from all namespaces
		if cfg.CLI.Verbose {
			fmt.Fprintf(os.Stderr, "Exporting all memories...\n")
		}
		// For now, we'll export from default namespace if none specified
		// TODO: Implement a way to list all namespaces and export from each
		namespace = cfg.Memory.DefaultNamespace
		if namespace == "" {
			namespace = "default"
		}
		memories, err = store.List(namespace, "", 0)
		if err != nil {
			return fmt.Errorf("failed to retrieve memories: %w", err)
		}
	}

	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d memories to export\n", len(memories))
	}

	// Prepare export data
	exportData := map[string]interface{}{
		"version":  "1.0",
		"exported_at": cmd.Context().Value("timestamp"),
		"namespace": namespace,
		"count":     len(memories),
		"memories":  memories,
	}

	// Write output based on format
	switch exportFormat {
	case "json":
		encoder := json.NewEncoder(outputFile)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(exportData); err != nil {
			return fmt.Errorf("failed to write JSON: %w", err)
		}
	case "yaml":
		return fmt.Errorf("yaml export not yet implemented")
	default:
		return fmt.Errorf("unsupported format: %s (use json or yaml)", exportFormat)
	}

	// Print success message if writing to file
	if len(args) == 1 {
		fmt.Fprintf(os.Stderr, "✓ Exported %d memories to %s\n", len(memories), exportFile)
	}

	return nil
}
