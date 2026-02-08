package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jbutlerdev/mem/internal/config"
	"github.com/jbutlerdev/mem/internal/embeddings"
	"github.com/jbutlerdev/mem/internal/models"
	"github.com/jbutlerdev/mem/internal/storage"
	"github.com/spf13/cobra"
)

var (
	updateNamespace string
	updateTags      []string
	updateMetadata  string
)

// UpdateCmd represents the update command
var UpdateCmd = &cobra.Command{
	Use:   "update [id] [new_content]",
	Short: "Update an existing memory",
	Long: `Update an existing memory's content, tags, or metadata.

When content is updated, a new semantic embedding is automatically generated.
Tags and metadata are replaced entirely (not merged).

USAGE:
  mem update <memory-id> [new content]
  mem update <memory-id> --tag tag1,tag2
  mem update <memory-id> --metadata '{"key": "value"}'

FLAGS:
  -n, --namespace <name>    Namespace for the memory
  -t, --tag <tags>          Replace tags (comma-separated)
  -m, --metadata <json>     Replace metadata (JSON string)

EXAMPLES:
  # Update content (regenerates embedding)
  mem update abc123 "Updated: User now prefers light mode"

  # Update tags only
  mem update abc123 -t important,archived

  # Update content and tags
  mem update abc123 -t preferences,ui "Updated content here"

  # Update metadata
  mem update abc123 -m '{"status": "resolved", "priority": "low"}'

  # Update in specific namespace
  mem update -n project-alpha abc123 "New content"`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runUpdate,
}

func init() {
	UpdateCmd.Flags().StringVarP(&updateNamespace, "namespace", "n", "", "namespace for the memory")
	UpdateCmd.Flags().StringSliceVarP(&updateTags, "tag", "t", []string{}, "update tags (replaces existing)")
	UpdateCmd.Flags().StringVarP(&updateMetadata, "metadata", "m", "", "update metadata JSON (replaces existing)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// The config and storage packages are used through getConfig() and initStorage()
	_ = (*config.Config)(nil)
	_ = (storage.Storage)(nil)

	// Parse arguments
	id := args[0]
	newContent := ""
	updateContent := false

	if len(args) == 2 {
		newContent = strings.TrimSpace(args[1])
		updateContent = true
	}

	// Determine namespace
	namespace := updateNamespace
	if namespace == "" {
		namespace = cfg.Memory.DefaultNamespace
	}
	if namespace == "" {
		namespace = "default"
	}

	// Initialize storage backend
	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Get the existing memory
	memory, err := store.Get(id)
	if err != nil {
		return fmt.Errorf("failed to retrieve memory: %w", err)
	}
	if memory == nil {
		return fmt.Errorf("memory with ID '%s' not found", id)
	}

	// Check namespace match
	if memory.Namespace != namespace {
		return fmt.Errorf("memory '%s' is in namespace '%s', not '%s'", id, memory.Namespace, namespace)
	}

	// Track if we need to update embedding
	needsEmbedding := updateContent

	// Update content if provided
	if updateContent {
		memory.Content = newContent
	}

	// Update tags if provided
	if len(updateTags) > 0 {
		memory.Tags = updateTags
	}

	// Update metadata if provided
	if updateMetadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(updateMetadata), &metadata); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
		memory.Metadata = metadata
	}

	// Validate memory
	if err := memory.Validate(); err != nil {
		return fmt.Errorf("invalid memory: %w", err)
	}

	// Generate new embedding if content changed
	if needsEmbedding {
		if cfg.CLI.Verbose {
			fmt.Fprintf(os.Stderr, "Generating new embedding for updated content...\n")
		}

		// Initialize embedding client
		embeddingConfig := embeddings.Config{
			BaseURL:    cfg.Embeddings.BaseURL,
			Model:      cfg.Embeddings.Model,
			APIKey:     cfg.Embeddings.APIKey,
			Dimensions: cfg.Embeddings.Dimensions,
			BatchSize:  cfg.Embeddings.BatchSize,
		}
		embeddingClient := embeddings.NewClient(embeddingConfig)

		embedding, err := embeddingClient.Embed(cmd.Context(), memory.Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}

		memory.Embedding = embedding

		if cfg.CLI.Verbose {
			fmt.Fprintf(os.Stderr, "Generated embedding with %d dimensions\n", len(embedding))
		}
	}

	// Update timestamp
	memory.UpdatedAt = memory.UpdatedAt // Will be set by storage layer

	// Update the memory in storage
	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Updating memory '%s'...\n", id)
	}

	if err := store.Update(memory); err != nil {
		return fmt.Errorf("failed to update memory: %w", err)
	}

	// Output success
	outputFormat := cfg.CLI.OutputFormat
	if outputFormat == "" {
		outputFormat = "text"
	}

	switch outputFormat {
	case "json":
		outputJSON(memory)
	case "yaml":
		return fmt.Errorf("yaml output not yet implemented")
	default:
		outputUpdateText(memory, updateContent, len(updateTags) > 0, updateMetadata != "")
	}

	return nil
}

func outputUpdateText(memory *models.Memory, updatedContent, updatedTags, updatedMetadata bool) {
	fmt.Printf("✓ Memory updated successfully\n")
	fmt.Printf("  ID:        %s\n", memory.ID)
	fmt.Printf("  Namespace: %s\n", memory.Namespace)

	if updatedContent {
		fmt.Printf("  Content:   %s\n", truncateString(memory.Content, 60))
	}

	if updatedTags {
		if len(memory.Tags) > 0 {
			fmt.Printf("  Tags:      %s\n", strings.Join(memory.Tags, ", "))
		} else {
			fmt.Printf("  Tags:      (cleared)\n")
		}
	}

	if updatedMetadata {
		if memory.Metadata != nil && len(memory.Metadata) > 0 {
			keys := make([]string, 0, len(memory.Metadata))
			for k := range memory.Metadata {
				keys = append(keys, k)
			}
			fmt.Printf("  Metadata:  %s\n", strings.Join(keys, ", "))
		} else {
			fmt.Printf("  Metadata:  (cleared)\n")
		}
	}
}
