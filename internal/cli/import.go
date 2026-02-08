package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jbutlerdev/mem/internal/embeddings"
	"github.com/jbutlerdev/mem/internal/models"
	"github.com/spf13/cobra"
)

var (
	importNamespace string
	importFormat    string
	importSkipEmbed bool
)

// ImportCmd represents the import command
var ImportCmd = &cobra.Command{
	Use:   "import <input_file>",
	Short: "Import memories from a backup file",
	Long: `Import memories from a file previously exported with 'mem export'.

Memories are re-indexed with new embeddings during import to ensure
compatibility with your current embedding model configuration.

USAGE:
  mem import <filename.json>
  mem import -n namespace <filename.json>

FLAGS:
  -n, --namespace <name>    Import into this namespace (overrides file)
      --format <fmt>       Input format: json or yaml (default: json)
      --skip-embeddings    Use existing embeddings (not recommended)

EXAMPLES:
  # Import from backup file
  mem import backup.json

  # Import into specific namespace
  mem import -n new-project project-backup.json

  # Import with existing embeddings (faster but may not match current model)
  mem import --skip-embeddings old-backup.json`,
	Args: cobra.ExactArgs(1),
	RunE: runImport,
}

func init() {
	ImportCmd.Flags().StringVarP(&importNamespace, "namespace", "n", "", "namespace to import into (overrides file namespace)")
	ImportCmd.Flags().StringVar(&importFormat, "format", "json", "input format: json | yaml")
	ImportCmd.Flags().BoolVar(&importSkipEmbed, "skip-embeddings", false, "skip re-generating embeddings (use existing)")
}

func runImport(cmd *cobra.Command, args []string) error {
	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Open input file
	inputFile := args[0]
	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer file.Close()

	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Reading memories from %s...\n", inputFile)
	}

	// Read and parse file based on format
	var exportData struct {
		Version   string                 `json:"version"`
		Namespace string                 `json:"namespace"`
		Count     int                    `json:"count"`
		Memories  []*models.MemoryImport `json:"memories"`
	}

	switch importFormat {
	case "json":
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&exportData); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}
	case "yaml":
		return fmt.Errorf("yaml import not yet implemented")
	default:
		return fmt.Errorf("unsupported format: %s (use json or yaml)", importFormat)
	}

	if len(exportData.Memories) == 0 {
		fmt.Println("No memories found in file")
		return nil
	}

	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d memories to import\n", len(exportData.Memories))
	}

	// Determine target namespace
	targetNamespace := importNamespace
	if targetNamespace == "" {
		targetNamespace = exportData.Namespace
	}
	if targetNamespace == "" {
		targetNamespace = cfg.Memory.DefaultNamespace
	}
	if targetNamespace == "" {
		targetNamespace = "default"
	}

	// Initialize storage backend
	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize embedding client if needed
	var embeddingClient *embeddings.Client
	if !importSkipEmbed {
		embeddingConfig := embeddings.Config{
			BaseURL:    cfg.Embeddings.BaseURL,
			Model:      cfg.Embeddings.Model,
			APIKey:     cfg.Embeddings.APIKey,
			Dimensions: cfg.Embeddings.Dimensions,
			BatchSize:  cfg.Embeddings.BatchSize,
		}
		embeddingClient = embeddings.NewClient(embeddingConfig)
	}

	// Import memories
	successCount := 0
	skippedCount := 0
	errorCount := 0

	for i, memImport := range exportData.Memories {
		if cfg.CLI.Verbose {
			fmt.Fprintf(os.Stderr, "Importing memory %d/%d...\n", i+1, len(exportData.Memories))
		}

		// Convert MemoryImport to Memory
		memory := &models.Memory{
			ID:        memImport.ID,
			Content:   memImport.Content,
			Namespace: targetNamespace,
			Tags:      memImport.Tags,
			Metadata:  memImport.Metadata,
			CreatedAt: memImport.CreatedAt,
			UpdatedAt: time.Now(), // Update timestamp on import
		}

		// Generate or reuse embedding
		if importSkipEmbed && len(memImport.Embedding) > 0 {
			memory.Embedding = memImport.Embedding
		} else {
			embedding, err := embeddingClient.Embed(cmd.Context(), memImport.Content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to generate embedding for memory %s: %v\n", memory.ID, err)
				errorCount++
				continue
			}
			memory.Embedding = embedding
		}

		// Validate memory
		if err := memory.Validate(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: invalid memory %s: %v\n", memory.ID, err)
			errorCount++
			continue
		}

		// Store memory
		if err := store.Store(memory); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to store memory %s: %v\n", memory.ID, err)
			errorCount++
			continue
		}

		successCount++
	}

	// Print summary
	fmt.Fprintf(os.Stderr, "\nImport complete:\n")
	fmt.Fprintf(os.Stderr, "  ✓ Imported: %d memories\n", successCount)
	if skippedCount > 0 {
		fmt.Fprintf(os.Stderr, "  ⊘ Skipped:  %d memories\n", skippedCount)
	}
	if errorCount > 0 {
		fmt.Fprintf(os.Stderr, "  ✗ Errors:   %d memories\n", errorCount)
	}

	return nil
}
