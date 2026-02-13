package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/jbutlerdev/mem/internal/config"
	"github.com/jbutlerdev/mem/internal/embeddings"
	"github.com/jbutlerdev/mem/internal/models"
	"github.com/jbutlerdev/mem/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	storeTags      []string
	storeMetadata  string
	storeStdin     bool
	storeNamespace string
)

// StoreCmd represents the store command
var StoreCmd = &cobra.Command{
	Use:   "store [content]",
	Short: "Store a new memory in semantic memory",
	Long: `Store a new memory with semantic embedding for later retrieval.

The content is converted to a vector embedding, enabling semantic search.
Content can be provided as an argument, piped via stdin, or from a file.

USAGE:
  mem store "content to remember"
  echo "content" | mem store
  cat file.txt | mem store

FLAGS:
  -n, --namespace <name>    Namespace for organizing memories (default: default)
  -t, --tag <tag>           Add tags (can be used multiple times)
  -m, --metadata <json>     Attach JSON metadata
      --stdin              Read from stdin explicitly

EXAMPLES:
  # Basic storage
  mem store "User prefers dark mode and vim keybindings"

  # With namespace and tags
  mem store -n work --tag important "Deadline is Friday at 5pm"

  # Pipe from echo
  echo "Multi-line note about project" | mem store

  # Read from file
  cat meeting_notes.txt | mem store -n meetings

  # With metadata
  mem store -m '{"source":"email","priority":"high"}' "Follow up with client"`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStore,
}

func init() {
	StoreCmd.Flags().StringVarP(&storeNamespace, "namespace", "n", "", "namespace for the memory")
	StoreCmd.Flags().StringSliceVarP(&storeTags, "tag", "t", []string{}, "tags to attach (can be used multiple times)")
	StoreCmd.Flags().StringVarP(&storeMetadata, "metadata", "m", "", "JSON metadata string")
	StoreCmd.Flags().BoolVar(&storeStdin, "stdin", false, "read content from stdin")
}

func runStore(cmd *cobra.Command, args []string) error {
	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine content source
	var content string
	if storeStdin || (len(args) == 1 && args[0] == "-") {
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		content = strings.TrimSpace(string(data))
	} else if len(args) == 1 {
		// Read from argument
		content = strings.TrimSpace(args[0])
	} else {
		// No content provided, try to read from stdin anyway (for piping)
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			// Stdin is available (not a terminal)
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			content = strings.TrimSpace(string(data))
		} else {
			return fmt.Errorf("no content provided. Use 'mem store <content>' or pipe content via stdin")
		}
	}

	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	// Determine namespace
	namespace := storeNamespace
	if namespace == "" {
		namespace = cfg.Memory.DefaultNamespace
	}
	if namespace == "" {
		namespace = "default"
	}

	// Parse metadata
	var metadata map[string]interface{}
	if storeMetadata != "" {
		if err := json.Unmarshal([]byte(storeMetadata), &metadata); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
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

	// Generate embedding
	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Generating embedding for content...\n")
	}

	embedding, err := embeddingClient.Embed(cmd.Context(), content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Generated embedding with %d dimensions\n", len(embedding))
	}

	// Initialize storage backend
	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Generate memory ID
	id := generateID()

	// Create memory object
	now := time.Now()
	memory := &models.Memory{
		ID:        id,
		Content:   content,
		Embedding: embedding,
		Namespace: namespace,
		Tags:      storeTags,
		Metadata:  metadata,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Validate memory
	if err := memory.Validate(); err != nil {
		return fmt.Errorf("invalid memory: %w", err)
	}

	// Store memory
	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Storing memory in namespace '%s'...\n", namespace)
	}

	if err := store.Store(memory); err != nil {
		return fmt.Errorf("failed to store memory: %w", err)
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
		outputText(memory)
	}

	return nil
}

func outputText(memory *models.Memory) {
	fmt.Printf("✓ Memory stored successfully\n")
	fmt.Printf("  ID:        %s\n", memory.ID)
	fmt.Printf("  Namespace: %s\n", memory.Namespace)
	if len(memory.Tags) > 0 {
		fmt.Printf("  Tags:      %s\n", strings.Join(memory.Tags, ", "))
	}
	fmt.Printf("  Content:   %s\n", truncateString(memory.Content, 60))
}

func outputJSON(memory *models.Memory) {
	data, err := json.MarshalIndent(memory, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func getConfig() (*config.Config, error) {
	// Use the centralized config loader which handles all defaults and aliases properly
	cfg, err := config.Load(viper.GetString("config"))
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Apply CLI flag overrides
	if viper.GetString("namespace") != "" {
		storeNamespace = viper.GetString("namespace")
	}
	if viper.GetString("backend") != "" {
		cfg.Memory.Backend = viper.GetString("backend")
	}
	if viper.GetString("output") != "" {
		cfg.CLI.OutputFormat = viper.GetString("output")
	}
	if viper.GetBool("verbose") {
		cfg.CLI.Verbose = true
	}

	return cfg, nil
}

func initStorage(cfg *config.Config) (storage.Storage, error) {
	switch cfg.Memory.Backend {
	case "chromem":
		return storage.NewChromeMStorageWithDim(cfg.Memory.Path, cfg.Embeddings.Dimensions)
	case "postgres":
		postgresConfig := &storage.PostgresConfig{
			Host:     cfg.Memory.Postgres.Host,
			Port:     cfg.Memory.Postgres.Port,
			Database: cfg.Memory.Postgres.Database,
			User:     cfg.Memory.Postgres.User,
			Password: cfg.Memory.Postgres.Password,
			SSLMode:  cfg.Memory.Postgres.SSLMode,
		}
		return storage.NewPostgresStorage(postgresConfig)
	default:
		return nil, fmt.Errorf("unsupported backend: %s", cfg.Memory.Backend)
	}
}

// generateID generates a unique ID for a memory
func generateID() string {
	// Simple ID generation based on timestamp
	// In production, you might want to use UUIDs or a more sophisticated method
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
