package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jbutlerdev/mem/internal/config"
	"github.com/jbutlerdev/mem/internal/embeddings"
	"github.com/jbutlerdev/mem/internal/storage"
	"github.com/spf13/cobra"
)

var (
	deleteNamespace string
	deleteQuery     string
	deleteForce     bool
)

// DeleteCmd represents the delete command
var DeleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "Delete a stored memory",
	Long: `Delete a stored memory by ID or by semantic query.

By default, deletion requires confirmation. Use --force to skip confirmation.

USAGE:
  mem delete <memory-id>
  mem delete --query "search query"

FLAGS:
  -n, --namespace <name>    Namespace for the memory
  -q, --query <query>       Delete by semantic search (interactive)
  -f, --force              Skip confirmation prompt

EXAMPLES:
  # Delete by ID (find ID with 'mem list')
  mem delete abc123def456

  # Delete in specific namespace
  mem delete -n project-alpha abc123def456

  # Delete by semantic search (interactive)
  mem delete --query "old temporary notes"

  # Force delete without confirmation
  mem delete --query "test data" --force`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDelete,
}

func init() {
	DeleteCmd.Flags().StringVarP(&deleteNamespace, "namespace", "n", "", "namespace for the memory")
	DeleteCmd.Flags().StringVarP(&deleteQuery, "query", "q", "", "delete by query match (interactive confirmation)")
	DeleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "skip confirmation prompt")
}

func runDelete(cmd *cobra.Command, args []string) error {
	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if we're deleting by ID or by query
	if deleteQuery != "" {
		// Delete by query
		if len(args) > 0 {
			return fmt.Errorf("cannot specify both an ID and --query flag")
		}
		return deleteByQuery(cmd, cfg)
	}

	// Delete by ID
	if len(args) == 0 {
		return fmt.Errorf("an ID is required when not using --query")
	}
	return deleteByID(cmd, cfg, args[0])
}

func deleteByID(cmd *cobra.Command, cfg *config.Config, id string) error {
	// Determine namespace
	namespace := deleteNamespace
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

	// Get the memory to show what will be deleted
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

	// Show what will be deleted
	fmt.Printf("Memory to be deleted:\n")
	fmt.Printf("  ID:        %s\n", memory.ID)
	fmt.Printf("  Namespace: %s\n", memory.Namespace)
	fmt.Printf("  Content:   %s\n", truncateString(memory.Content, 60))
	if len(memory.Tags) > 0 {
		fmt.Printf("  Tags:      %s\n", strings.Join(memory.Tags, ", "))
	}
	fmt.Println()

	// Confirm deletion
	if !deleteForce {
		if !confirmAction("Delete this memory?") {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	// Delete the memory
	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Deleting memory '%s'...\n", id)
	}

	if err := store.Delete(id); err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	fmt.Println("✓ Memory deleted successfully.")
	return nil
}

func deleteByQuery(cmd *cobra.Command, cfg *config.Config) error {
	// Determine namespace
	namespace := deleteNamespace
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

	// Initialize embedding client
	embeddingConfig := embeddings.Config{
		BaseURL:    cfg.Embeddings.BaseURL,
		Model:      cfg.Embeddings.Model,
		APIKey:     cfg.Embeddings.APIKey,
		Dimensions: cfg.Embeddings.Dimensions,
		BatchSize:  cfg.Embeddings.BatchSize,
	}
	embeddingClient := embeddings.NewClient(embeddingConfig)

	// Generate embedding for query
	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Generating embedding for query...\n")
	}

	queryEmbedding, err := embeddingClient.Embed(cmd.Context(), deleteQuery)
	if err != nil {
		return fmt.Errorf("failed to generate embedding for query: %w", err)
	}

	// Search for matching memories
	queryOpts := &storage.QueryOptions{
		Namespace: namespace,
		Limit:     20,  // Get more results for review
		Threshold: 0.5, // Lower threshold to find potential matches
	}

	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Searching for memories matching query...\n")
	}

	results, err := store.Query(queryEmbedding, queryOpts)
	if err != nil {
		return fmt.Errorf("failed to search for memories: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("No memories found matching the query.")
		return nil
	}

	// Show results and ask which to delete
	fmt.Printf("Found %d memory(s) matching query '%s':\n\n", len(results), deleteQuery)

	for i, result := range results {
		fmt.Printf("%d. [Score: %.2f] %s\n", i+1, result.Score, truncateString(result.Memory.Content, 70))
		fmt.Printf("   ID:      %s\n", result.Memory.ID)
		fmt.Printf("   Created: %s\n", result.Memory.CreatedAt.Format("2006-01-02 15:04:05"))
		if len(result.Memory.Tags) > 0 {
			fmt.Printf("   Tags:    %s\n", strings.Join(result.Memory.Tags, ", "))
		}
		fmt.Println()
	}

	// Ask which memories to delete
	reader := bufio.NewReader(os.Stdin)

	var idsToDelete []string
	if !deleteForce {
		fmt.Print("Enter numbers to delete (comma-separated, or 'all' for all, 'none' to cancel): ")
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		input = strings.TrimSpace(input)

		if strings.ToLower(input) == "none" || input == "" {
			fmt.Println("Deletion cancelled.")
			return nil
		}

		if strings.ToLower(input) == "all" {
			for _, result := range results {
				idsToDelete = append(idsToDelete, result.Memory.ID)
			}
		} else {
			// Parse comma-separated numbers
			parts := strings.Split(input, ",")
			for _, part := range parts {
				var idx int
				_, err := fmt.Sscanf(strings.TrimSpace(part), "%d", &idx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: invalid number '%s', skipping\n", part)
					continue
				}
				if idx < 1 || idx > len(results) {
					fmt.Fprintf(os.Stderr, "Warning: index %d out of range, skipping\n", idx)
					continue
				}
				idsToDelete = append(idsToDelete, results[idx-1].Memory.ID)
			}
		}
	} else {
		// Force mode: delete all results
		for _, result := range results {
			idsToDelete = append(idsToDelete, result.Memory.ID)
		}
	}

	if len(idsToDelete) == 0 {
		fmt.Println("No memories selected for deletion.")
		return nil
	}

	// Confirm deletion
	if !deleteForce {
		fmt.Printf("\nAbout to delete %d memory(s). ", len(idsToDelete))
		if !confirmAction("Continue?") {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	// Delete the memories
	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Deleting %d memory(s)...\n", len(idsToDelete))
	}

	successCount := 0
	for _, id := range idsToDelete {
		if err := store.Delete(id); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to delete memory '%s': %v\n", id, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("✓ Successfully deleted %d memory(s).\n", successCount)
	if successCount < len(idsToDelete) {
		fmt.Fprintf(os.Stderr, "Warning: %d memory(s) failed to delete\n", len(idsToDelete)-successCount)
	}

	return nil
}

// confirmAction prompts the user for confirmation
func confirmAction(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/N]: ", prompt)
		input, err := reader.ReadString('\n')
		if err != nil {
			return false
		}
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "y" || input == "yes" {
			return true
		}
		if input == "n" || input == "no" || input == "" {
			return false
		}
		fmt.Println("Please enter 'y' or 'n'.")
	}
}
