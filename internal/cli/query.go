package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/jbutlerdev/mem/internal/embeddings"
	"github.com/jbutlerdev/mem/internal/models"
	"github.com/jbutlerdev/mem/internal/reranker"
	"github.com/jbutlerdev/mem/internal/storage"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v3"
)

var (
	queryNamespace string
	queryLimit     int
	queryTop       int
	queryThreshold float32
	queryNoRerank  bool
	queryOutput    string
)

// QueryCmd represents the query command
var QueryCmd = &cobra.Command{
	Use:   "query [query]",
	Short: "Search memories using semantic search",
	Long: `Query memories using semantic search with optional re-ranking.

Your query is converted to a vector embedding and matched against stored memories
using cosine similarity. Results can be re-ranked for improved relevance.

USAGE:
  mem query "search query"

FLAGS:
  -n, --namespace <name>    Search only in this namespace
  -l, --limit <num>         Number of results to return (default: 5)
      --top <num>           Initial retrieval count before re-ranking
      --threshold <0-1>     Minimum similarity score (default: 0.6)
      --no-rerank           Disable re-ranking for faster results
  -o, --output <format>     Output format: text, json, yaml (default: text)

EXAMPLES:
  # Basic semantic search
  mem query "What are the user's preferences?"

  # Search in specific namespace
  mem query -n project-alpha "database version"

  # Get more results
  mem query --limit 10 "meeting notes about roadmap"

  # Disable reranking for faster queries
  mem query --no-rerank "quick search"

  # JSON output for scripting
  mem query -o json "typescript config" | jq '.[].content'`,
	Args: cobra.ExactArgs(1),
	RunE: runQuery,
}

func init() {
	QueryCmd.Flags().StringVarP(&queryNamespace, "namespace", "n", "", "namespace to search")
	QueryCmd.Flags().IntVarP(&queryLimit, "limit", "l", 5, "number of results to return")
	QueryCmd.Flags().IntVar(&queryTop, "top", 10, "initial retrieval count before re-ranking")
	QueryCmd.Flags().Float32Var(&queryThreshold, "threshold", 0.6, "minimum similarity score (0-1)")
	QueryCmd.Flags().BoolVar(&queryNoRerank, "no-rerank", false, "disable re-ranking")
	QueryCmd.Flags().StringVarP(&queryOutput, "output", "o", "", "output format: text | json | yaml")
}

func runQuery(cmd *cobra.Command, args []string) error {
	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	query := strings.TrimSpace(args[0])
	if query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	// Determine namespace
	namespace := queryNamespace
	if namespace == "" {
		namespace = cfg.Memory.DefaultNamespace
	}
	if namespace == "" {
		namespace = "default"
	}

	// Apply config defaults
	limit := queryLimit
	if limit == 5 && cfg.Query.DefaultLimit != 0 {
		limit = cfg.Query.DefaultLimit
	}

	top := queryTop
	if top == 10 && cfg.Query.DefaultLimit != 0 {
		// Default top to 2x limit if not specified
		top = limit * 2
	}

	threshold := queryThreshold
	if threshold == 0.6 && cfg.Query.MinSimilarity != 0 {
		threshold = float32(cfg.Query.MinSimilarity)
	}

	// Determine output format
	outputFormat := queryOutput
	if outputFormat == "" {
		outputFormat = cfg.CLI.OutputFormat
	}
	if outputFormat == "" {
		outputFormat = "text"
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

	// Generate query embedding
	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Generating embedding for query...\n")
	}

	queryEmbedding, err := embeddingClient.Embed(cmd.Context(), query)
	if err != nil {
		return fmt.Errorf("failed to generate query embedding: %w", err)
	}

	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Generated query embedding with %d dimensions\n", len(queryEmbedding))
	}

	// Initialize storage backend
	store, err := initStorage(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer store.Close()

	// Determine if reranking should be used
	rerankEnabled := !queryNoRerank && cfg.Query.RerankEnabled

	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Reranking: %s\n", map[bool]string{true: "enabled", false: "disabled"}[rerankEnabled])
	}

	// Set initial retrieval count based on reranking
	initialTop := top
	if rerankEnabled {
		// Fetch more results for reranking
		if cfg.Reranking.TopK > 0 {
			initialTop = cfg.Reranking.TopK
		}
		if queryTop > initialTop {
			initialTop = queryTop
		}
	}

	// Execute vector search
	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Searching for memories (top %d)...\n", initialTop)
	}

	queryOpts := &storage.QueryOptions{
		Namespace: namespace,
		Limit:     initialTop,
		Threshold: threshold,
	}

	results, err := store.Query(queryEmbedding, queryOpts)
	if err != nil {
		return fmt.Errorf("failed to query memories: %w", err)
	}

	if cfg.CLI.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d initial results\n", len(results))
	}

	// Apply re-ranking if enabled
	var finalResults []models.SearchResult
	if rerankEnabled && len(results) > 0 {
		if cfg.CLI.Verbose {
			fmt.Fprintf(os.Stderr, "Re-ranking top %d results...\n", len(results))
		}

		// Initialize reranker client
		rerankerConfig := reranker.Config{
			BaseURL:   cfg.Embeddings.BaseURL,
			Model:     cfg.Reranking.Model,
			APIKey:    cfg.Embeddings.APIKey,
			Enabled:   true,
			TopK:      cfg.Reranking.TopK,
			Threshold: float64(cfg.Reranking.Threshold),
		}
		rerankerClient := reranker.NewClient(rerankerConfig)

		// Prepare documents for reranking
		docsToRerank := make([]reranker.DocumentToRerank, len(results))
		for i, result := range results {
			docsToRerank[i] = reranker.DocumentToRerank{
				ID:      result.Memory.ID,
				Content: result.Memory.Content,
				Score:   result.Score,
			}
		}

		// Perform reranking
		rerankedDocs, err := rerankerClient.Rerank(cmd.Context(), query, docsToRerank)
		if err != nil {
			if cfg.CLI.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: reranking failed, using original results: %v\n", err)
			}
			// Fall back to original results
			finalResults = results
		} else {
			// Convert reranked docs back to search results
			resultMap := make(map[string]*models.SearchResult)
			for i := range results {
				resultMap[results[i].Memory.ID] = &results[i]
			}

			finalResults = make([]models.SearchResult, len(rerankedDocs))
			for i, rerankedDoc := range rerankedDocs {
				if originalResult, exists := resultMap[rerankedDoc.ID]; exists {
					finalResults[i] = models.SearchResult{
						Memory:   originalResult.Memory,
						Score:    rerankedDoc.FinalScore,
						Reranked: true,
						Rank:     rerankedDoc.Rank,
					}
				}
			}
		}

		if cfg.CLI.Verbose {
			fmt.Fprintf(os.Stderr, "Re-ranking complete\n")
		}
	} else {
		finalResults = results
	}

	// Apply limit
	if len(finalResults) > limit {
		finalResults = finalResults[:limit]
	}

	// Output results
	switch outputFormat {
	case "json":
		outputQueryResultsJSON(finalResults)
	case "yaml":
		outputQueryResultsYAML(finalResults)
	default:
		outputQueryResultsText(finalResults)
	}

	return nil
}

func outputQueryResultsText(results []models.SearchResult) {
	if len(results) == 0 {
		fmt.Println("No results found.")
		return
	}

	fmt.Printf("Found %d result(s):\n\n", len(results))
	for i, result := range results {
		fmt.Printf("%d. [%s] %s\n", i+1, result.Memory.Namespace, truncateString(result.Memory.Content, 70))
		fmt.Printf("   ID:     %s\n", result.Memory.ID)
		fmt.Printf("   Score:  %.4f", result.Score)
		if result.Reranked {
			fmt.Printf(" (reranked)")
		}
		fmt.Println()
		if len(result.Memory.Tags) > 0 {
			fmt.Printf("   Tags:   %s\n", strings.Join(result.Memory.Tags, ", "))
		}
		fmt.Println()
	}
}

func outputQueryResultsJSON(results []models.SearchResult) {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func outputQueryResultsYAML(results []models.SearchResult) {
	data, err := yaml.Marshal(results)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling YAML: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
