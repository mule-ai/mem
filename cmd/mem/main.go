package main

import (
	"fmt"
	"os"

	"github.com/jbutlerdev/mem/internal/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Build variables (set via ldflags)
	version   = "dev"
	buildTime = "unknown"

	// Global flags
	cfgFile   string
	namespace string
	backend   string
	output    string
	verbose   bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mem",
	Short: "Semantic Memory CLI - Store and search memories using AI embeddings",
	Long: `mem is a CLI tool for storing, querying, and managing memories using vector embeddings
and semantic search. Perfect for building persistent memory for AI agents, note-taking with
semantic search, and knowledge management.

QUICK START:
  # Store your first memory
  mem store "User prefers dark mode and vim keybindings"

  # Search for memories semantically
  mem query "What are the user's preferences?"

  # List all memories
  mem list

COMMON WORKFLOWS:
  # Store with tags and namespace
  mem store -n project-alpha --tag important "Database uses PostgreSQL 14"

  # Search in specific namespace
  mem query -n project-alpha "database configuration"

  # Pipe content to store
  cat notes.txt | mem store -

  # Export memories
  mem export -n project-alpha backup.json

FEATURES:
  • Semantic search using vector embeddings
  • Re-ranking for improved relevance
  • Namespaces for organizing memories
  • Tags and metadata support
  • Multiple storage backends (ChromeM, PostgreSQL)
  • Import/export for backups

CONFIGURATION:
  Configuration is loaded from ~/.mem/config.yaml or specified with --config.
  Environment variables can also be used for API keys and settings.

EXAMPLES:
  mem store "Meeting notes: Discussed Q4 roadmap"
  mem query "what was discussed about Q4"
  mem list -n work --tag meetings
  mem delete mem-1234567890`,
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.mem/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace for the operation")
	rootCmd.PersistentFlags().StringVar(&backend, "backend", "", "storage backend to use (chromem | postgres)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "output format: text | json | yaml")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose logging")

	// Bind flags to viper
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("backend", rootCmd.PersistentFlags().Lookup("backend"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Add subcommands
	rootCmd.AddCommand(cli.StoreCmd)
	rootCmd.AddCommand(cli.QueryCmd)
	rootCmd.AddCommand(cli.ListCmd)
	rootCmd.AddCommand(cli.DeleteCmd)
	rootCmd.AddCommand(cli.UpdateCmd)
	rootCmd.AddCommand(cli.ExportCmd)
	rootCmd.AddCommand(cli.ImportCmd)
	rootCmd.AddCommand(cli.NamespaceCmd)
	rootCmd.AddCommand(cli.ConfigCmd)
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home + "/.mem")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}
}

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}