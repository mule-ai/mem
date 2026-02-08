package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jbutlerdev/mem/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// ConfigCmd represents the config command group
var ConfigCmd = &cobra.Command{
	Use:   "config [command]",
	Short: "Manage mem configuration",
	Long: `Manage mem configuration for embeddings, storage, and behavior.

The configuration file is located at ~/.mem/config.yaml by default.
Local overrides can be placed in ./.memconfig in the current directory.

USAGE:
  mem config show
  mem config get <key>
  mem config set <key> <value>
  mem config edit

CONFIGURATION PRECEDENCE:
  1. CLI flags (highest priority)
  2. ./.memconfig (local override)
  3. ~/.mem/config.yaml (global config)
  4. Built-in defaults (lowest priority)

EXAMPLES:
  # View current configuration
  mem config show

  # Get a specific value
  mem config get embeddings.model

  # Set a value
  mem config set query.default_limit 10

  # Open config in editor
  mem config edit`,
}

// ConfigShowCmd represents the config show command
var ConfigShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current effective configuration",
	Long: `Show the current effective configuration after applying all precedence rules.

Displays the merged configuration from defaults, global config, local config,
and CLI flags.

EXAMPLES:
  mem config show`,
	RunE: runConfigShow,
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get config file path
	cfgFile := viper.GetString("config")
	if cfgFile == "" {
		homeDir, _ := os.UserHomeDir()
		cfgFile = filepath.Join(homeDir, ".mem", "config.yaml")
	}

	// Check if local config exists
	localCfgExists := false
	if localCfg, err := os.Stat("./.memconfig"); err == nil {
		if !localCfg.IsDir() {
			localCfgExists = true
		}
	}

	// Display configuration info
	fmt.Printf("Configuration File: %s\n", cfgFile)
	if localCfgExists {
		fmt.Printf("Local Override:     ./.memconfig (active)\n")
	}
	fmt.Println()

	// Output config as YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

// ConfigGetCmd represents the config get command
var ConfigGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Get a specific configuration value by key.

Use dot notation for nested keys.

USAGE:
  mem config get <key>

EXAMPLES:
  mem config get memory.backend
  mem config get embeddings.model
  mem config get query.default_limit`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	// Get configuration
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get value by key using a helper function
	value, err := getConfigValue(cfg, key)
	if err != nil {
		return err
	}

	// Display value
	fmt.Printf("%s: %v\n", key, value)
	return nil
}

// ConfigSetCmd represents the config set command
var ConfigSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value in the global config file.

Use dot notation for nested keys. The value is written to ~/.mem/config.yaml.

USAGE:
  mem config set <key> <value>

EXAMPLES:
  mem config set memory.backend chromem
  mem config set embeddings.model text-embedding-qwen3-embedding-8b
  mem config set query.default_limit 10
  mem config set reranking.enabled true`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	// Get config file path
	cfgFile := viper.GetString("config")
	if cfgFile == "" {
		homeDir, _ := os.UserHomeDir()
		cfgFile = filepath.Join(homeDir, ".mem", "config.yaml")
	}

	// Load current config
	cfg := config.DefaultConfig()

	// Check if config file exists
	if _, err := os.Stat(cfgFile); err == nil {
		// File exists, load it
		data, err := os.ReadFile(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Set the new value
	if err := setConfigValue(cfg, key, value); err != nil {
		return err
	}

	// Validate the updated config
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Ensure directory exists
	cfgDir := filepath.Dir(cfgFile)
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(cfgFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✓ Set %s = %v\n", key, value)
	fmt.Printf("  Config file: %s\n", cfgFile)
	return nil
}

// ConfigEditCmd represents the config edit command
var ConfigEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open configuration in editor",
	Long: `Open the configuration file in your default editor.

Uses the EDITOR environment variable, or defaults to 'vi'.
After editing, the configuration will be validated for errors.

EXAMPLES:
  EDITOR=nano mem config edit
  mem config edit`,
	RunE: runConfigEdit,
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	// Get config file path
	cfgFile := viper.GetString("config")
	if cfgFile == "" {
		homeDir, _ := os.UserHomeDir()
		cfgFile = filepath.Join(homeDir, ".mem", "config.yaml")
	}

	// Ensure config file exists
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		// Create default config
		cfg := config.DefaultConfig()

		// Ensure directory exists
		cfgDir := filepath.Dir(cfgFile)
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Marshal and write
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		if err := os.WriteFile(cfgFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		fmt.Printf("Created new config file: %s\n", cfgFile)
	}

	// Get editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	// Open editor
	cmdArgs := []string{cfgFile}
	editCmd := exec.Command(editor, cmdArgs...)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	fmt.Printf("Opening %s with %s...\n", cfgFile, editor)

	if err := editCmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	// Validate the edited config
	data, err := os.ReadFile(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := config.DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	fmt.Println("✓ Configuration is valid")
	return nil
}

func init() {
	ConfigCmd.AddCommand(ConfigShowCmd)
	ConfigCmd.AddCommand(ConfigGetCmd)
	ConfigCmd.AddCommand(ConfigSetCmd)
	ConfigCmd.AddCommand(ConfigEditCmd)
}

// Helper functions for config access

func getConfigValue(cfg *config.Config, key string) (interface{}, error) {
	switch key {
	case "memory.path":
		return cfg.Memory.Path, nil
	case "memory.default_namespace":
		return cfg.Memory.DefaultNamespace, nil
	case "memory.backend":
		return cfg.Memory.Backend, nil
	case "embeddings.base_url":
		return cfg.Embeddings.BaseURL, nil
	case "embeddings.model":
		return cfg.Embeddings.Model, nil
	case "embeddings.dimensions":
		return cfg.Embeddings.Dimensions, nil
	case "embeddings.batch_size":
		return cfg.Embeddings.BatchSize, nil
	case "reranking.enabled":
		return cfg.Reranking.Enabled, nil
	case "reranking.model":
		return cfg.Reranking.Model, nil
	case "reranking.top_k":
		return cfg.Reranking.TopK, nil
	case "reranking.threshold":
		return cfg.Reranking.Threshold, nil
	case "query.default_limit":
		return cfg.Query.DefaultLimit, nil
	case "query.min_similarity":
		return cfg.Query.MinSimilarity, nil
	case "query.rerank_enabled":
		return cfg.Query.RerankEnabled, nil
	case "cli.output_format":
		return cfg.CLI.OutputFormat, nil
	case "cli.verbose":
		return cfg.CLI.Verbose, nil
	case "cli.color":
		return cfg.CLI.Color, nil
	default:
		return nil, fmt.Errorf("unknown configuration key: %s", key)
	}
}

func setConfigValue(cfg *config.Config, key, value string) error {
	switch key {
	case "memory.path":
		cfg.Memory.Path = value
	case "memory.default_namespace":
		cfg.Memory.DefaultNamespace = value
	case "memory.backend":
		cfg.Memory.Backend = value
	case "embeddings.base_url":
		cfg.Embeddings.BaseURL = value
	case "embeddings.model":
		cfg.Embeddings.Model = value
	case "embeddings.api_key":
		cfg.Embeddings.APIKey = value
	case "embeddings.dimensions":
		var dim int
		if _, err := fmt.Sscanf(value, "%d", &dim); err != nil {
			return fmt.Errorf("invalid dimensions value: %s", value)
		}
		cfg.Embeddings.Dimensions = dim
	case "embeddings.batch_size":
		var size int
		if _, err := fmt.Sscanf(value, "%d", &size); err != nil {
			return fmt.Errorf("invalid batch_size value: %s", value)
		}
		cfg.Embeddings.BatchSize = size
	case "reranking.enabled":
		var enabled bool
		if value == "true" {
			enabled = true
		} else if value == "false" {
			enabled = false
		} else {
			return fmt.Errorf("invalid enabled value: %s (use true or false)", value)
		}
		cfg.Reranking.Enabled = enabled
	case "reranking.model":
		cfg.Reranking.Model = value
	case "reranking.top_k":
		var topK int
		if _, err := fmt.Sscanf(value, "%d", &topK); err != nil {
			return fmt.Errorf("invalid top_k value: %s", value)
		}
		cfg.Reranking.TopK = topK
	case "reranking.threshold":
		var threshold float64
		if _, err := fmt.Sscanf(value, "%f", &threshold); err != nil {
			return fmt.Errorf("invalid threshold value: %s", value)
		}
		cfg.Reranking.Threshold = threshold
	case "query.default_limit":
		var limit int
		if _, err := fmt.Sscanf(value, "%d", &limit); err != nil {
			return fmt.Errorf("invalid default_limit value: %s", value)
		}
		cfg.Query.DefaultLimit = limit
	case "query.min_similarity":
		var similarity float64
		if _, err := fmt.Sscanf(value, "%f", &similarity); err != nil {
			return fmt.Errorf("invalid min_similarity value: %s", value)
		}
		cfg.Query.MinSimilarity = similarity
	case "query.rerank_enabled":
		var enabled bool
		if value == "true" {
			enabled = true
		} else if value == "false" {
			enabled = false
		} else {
			return fmt.Errorf("invalid rerank_enabled value: %s (use true or false)", value)
		}
		cfg.Query.RerankEnabled = enabled
	case "cli.output_format":
		cfg.CLI.OutputFormat = value
	case "cli.verbose":
		var verbose bool
		if value == "true" {
			verbose = true
		} else if value == "false" {
			verbose = false
		} else {
			return fmt.Errorf("invalid verbose value: %s (use true or false)", value)
		}
		cfg.CLI.Verbose = verbose
	case "cli.color":
		var color bool
		if value == "true" {
			color = true
		} else if value == "false" {
			color = false
		} else {
			return fmt.Errorf("invalid color value: %s (use true or false)", value)
		}
		cfg.CLI.Color = color
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	return nil
}
