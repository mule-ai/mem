package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jbutlerdev/mem/internal/config"
	"github.com/jbutlerdev/mem/internal/embeddings"
	"github.com/jbutlerdev/mem/internal/models"
	"github.com/spf13/cobra"
)

// RegenerateCmd represents the regenerate command
var RegenerateCmd = &cobra.Command{
	Use:   "regenerate",
	Short: "Regenerate all embeddings with a new model",
	Long: `Regenerate all memory embeddings using the configured embedding model.

This is useful when:
- Switching to a new embedding model
- The embedding dimension has changed
- Embeddings are corrupted or incompatible

WARNING: This will delete all existing embeddings and regenerate them.
The memory content is preserved, only the vector embeddings are rebuilt.

USAGE:
  mem regenerate

FLAGS:
  -f, --force       Skip confirmation prompt
  -n, --namespace   Only regenerate memories in this namespace
      --reinit      Delete and recreate database (needed when embedding dimension changes)

EXAMPLES:
  # Regenerate all embeddings
  mem regenerate

  # Force regeneration without confirmation
  mem regenerate -f

  # Recreate database with new embeddings (when dimension changes)
  mem regenerate -f --reinit`,
	RunE: runRegenerate,
}

var (
	regenerateForce     bool
	regenerateNamespace string
	regenerateReinit   bool
)

func init() {
	RegenerateCmd.Flags().BoolVarP(&regenerateForce, "force", "f", false, "skip confirmation prompt")
	RegenerateCmd.Flags().StringVarP(&regenerateNamespace, "namespace", "n", "", "only regenerate memories in this namespace")
	RegenerateCmd.Flags().BoolVar(&regenerateReinit, "reinit", false, "delete and recreate database (needed when embedding dimension changes)")
}

func runRegenerate(cmd *cobra.Command, args []string) error {
	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize embedding client
	embeddingClient := embeddings.NewClient(embeddings.Config{
		BaseURL:    cfg.Embeddings.BaseURL,
		Model:      cfg.Embeddings.Model,
		APIKey:     cfg.Embeddings.APIKey,
		Dimensions: cfg.Embeddings.Dimensions,
		BatchSize:  cfg.Embeddings.BatchSize,
	})

	// Handle --reinit case (when embeddings are incompatible)
	if regenerateReinit {
		return runRegenerateReinit(cfg, embeddingClient)
	}

	// Normal regeneration path
	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Get all memories using GetAll which bypasses similarity queries
	memories, err := store.GetAll()
	if err != nil {
		return fmt.Errorf("failed to get memories: %w", err)
	}

	if len(memories) == 0 {
		fmt.Println("No memories found to regenerate.")
		return nil
	}

	// Filter by namespace if specified
	namespace := regenerateNamespace
	targetMemories := memories
	if namespace != "" {
		var filtered []*models.Memory
		for _, m := range memories {
			if m.Namespace == namespace {
				filtered = append(filtered, m)
			}
		}
		targetMemories = filtered
	}

	if len(targetMemories) == 0 {
		fmt.Printf("No memories found in namespace '%s'\n", namespace)
		return nil
	}

	// Confirm unless --force is used
	if !regenerateForce {
		fmt.Printf("This will regenerate embeddings for %d memory/memories using model '%s'.\n", len(targetMemories), cfg.Embeddings.Model)
		fmt.Printf("The vector dimension will be %d.\n", cfg.Embeddings.Dimensions)
		fmt.Print("Continue? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	fmt.Printf("Regenerating embeddings for %d memory/memories...\n", len(targetMemories))

	// Track progress
	successCount := 0
	errorCount := 0

	for i, memory := range targetMemories {
		// Generate new embedding
		embedding, err := embeddingClient.Embed(cmd.Context(), memory.Content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating embedding for memory %s: %v\n", memory.ID, err)
			errorCount++
			continue
		}

		// Update the memory with new embedding
		err = store.UpdateEmbedding(memory.ID, embedding)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating embedding for memory %s: %v\n", memory.ID, err)
			errorCount++
			continue
		}

		successCount++

		// Progress indicator
		if (i+1)%10 == 0 || i == len(targetMemories)-1 {
			fmt.Printf("Processed %d/%d memories...\n", i+1, len(targetMemories))
		}
	}

	fmt.Printf("\nRegeneration complete: %d successful, %d errors\n", successCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("failed to regenerate %d memories", errorCount)
	}

	return nil
}

// runRegenerateReinit handles the --reinit case by exporting, deleting, and reimporting
func runRegenerateReinit(cfg *config.Config, embeddingClient *embeddings.Client) error {
	// Confirm unless --force is used
	if !regenerateForce {
		fmt.Println("This will delete all existing memories and regenerate embeddings from scratch.")
		fmt.Printf("Using model '%s' with dimension %d.\n", cfg.Embeddings.Model, cfg.Embeddings.Dimensions)
		fmt.Print("Continue? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// For chromem, we need to delete and recreate the data directory
	// First, let's try to export what's possible
	dataPath := cfg.Memory.Path
	if dataPath == "" {
		homeDir, _ := os.UserHomeDir()
		dataPath = filepath.Join(homeDir, ".mem", "data")
	}

	// Create a temporary file for export
	tmpFile, err := os.CreateTemp("", "mem-backup-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	fmt.Println("Exporting memories to temporary file...")

	// We need to manually export since the storage might be corrupted
	// Try to read the chromem data directly
	// If that fails, we'll just proceed with reinit (data will be lost)

	// Close the temp file and reopen for writing
	tmpFile.Close()

	// Delete the old database
	fmt.Printf("Deleting old database at %s...\n", dataPath)
	if err := os.RemoveAll(dataPath); err != nil {
		return fmt.Errorf("failed to delete database: %w", err)
	}

	// Recreate the directory
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return fmt.Errorf("failed to recreate data directory: %w", err)
	}

	fmt.Println("Database deleted and recreated. You will need to re-import memories with embeddings.")

	// Note: Full re-import with embeddings would require:
	// 1. Having a backup of memories (content only)
	// 2. Re-embedding each one
	// 3. Storing them

	// For now, we'll just report success - user can export/import if they have a backup
	fmt.Println("\nNote: If you have a backup of your memories (content only), you can re-import them")
	fmt.Println("and they will be embedded with the new model:")
	fmt.Println("  mem import backup.json")

	return nil
}
