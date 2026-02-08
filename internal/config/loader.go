package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Load loads the configuration from multiple sources with proper precedence
// Precedence order (highest to lowest):
// 1. CLI flags (handled by cobra, not here)
// 2. Environment variables
// 3. .memconfig in current working directory
// 4. ~/.mem/config.yaml (global config)
// 5. Built-in defaults
func Load(configFile string) (*Config, error) {
	// Start with default configuration
	v := viper.New()
	setDefaults(v)

	// Try to load global config from ~/.mem/config.yaml
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	globalConfigPath := filepath.Join(homeDir, ".mem", "config.yaml")

	// Only load global config if it exists
	if _, err := os.Stat(globalConfigPath); err == nil {
		v.SetConfigFile(globalConfigPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read global config from %s: %w", globalConfigPath, err)
		}
	}

	// Try to load local config override from ./.memconfig
	localConfigPath := filepath.Join(".", ".memconfig")
	if _, err := os.Stat(localConfigPath); err == nil {
		// Create a new viper instance for local config to merge
		localV := viper.New()
		localV.SetConfigFile(localConfigPath)
		// Set config type explicitly since .memconfig has no standard extension
		localV.SetConfigType("yaml")
		if err := localV.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read local config from %s: %w", localConfigPath, err)
		}
		// Merge local config on top of global config
		if err := v.MergeConfigMap(localV.AllSettings()); err != nil {
			return nil, fmt.Errorf("failed to merge local config: %w", err)
		}
	}

	// If a specific config file was provided via flag, load it (highest precedence before env vars)
	if configFile != "" {
		if _, err := os.Stat(configFile); err != nil {
			return nil, fmt.Errorf("config file not found: %s", configFile)
		}
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read specified config from %s: %w", configFile, err)
		}
	}

	// Set up environment variable support
	// Environment variables should be prefixed with MEM_
	// For example: MEM_EMBEDDINGS_BASE_URL, MEMORY_BACKEND
	v.SetEnvPrefix("MEM")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Unmarshal into config struct
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// setDefaults sets the default configuration values
func setDefaults(v *viper.Viper) {
	homeDir, _ := os.UserHomeDir()

	// Memory defaults
	v.SetDefault("memory.path", filepath.Join(homeDir, ".mem", "data"))
	v.SetDefault("memory.default_namespace", "default")
	v.SetDefault("memory.backend", "chromem")
	v.SetDefault("memory.postgres.host", "localhost")
	v.SetDefault("memory.postgres.port", 5432)
	v.SetDefault("memory.postgres.database", "mem")
	v.SetDefault("memory.postgres.user", "postgres")
	v.SetDefault("memory.postgres.password", "")
	v.SetDefault("memory.postgres.sslmode", "disable")

	// Embeddings defaults
	v.SetDefault("embeddings.base_url", "http://10.10.199.29:8080")
	v.SetDefault("embeddings.model", "text-embedding-qwen3-embedding-8b")
	v.SetDefault("embeddings.api_key", "")
	v.SetDefault("embeddings.dimensions", 1024)
	v.SetDefault("embeddings.batch_size", 10)

	// Reranking defaults
	v.SetDefault("reranking.enabled", true)
	v.SetDefault("reranking.model", "qwen3-reranker-8b")
	v.SetDefault("reranking.top_k", 10)
	v.SetDefault("reranking.threshold", 0.5)

	// Query defaults
	v.SetDefault("query.default_limit", 5)
	v.SetDefault("query.min_similarity", 0.6)
	v.SetDefault("query.rerank_enabled", true)

	// CLI defaults
	v.SetDefault("cli.output_format", "text")
	v.SetDefault("cli.verbose", false)
	v.SetDefault("cli.color", true)
}

// InitConfig initializes the configuration directory and files
func InitConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	memDir := filepath.Join(homeDir, ".mem")
	dataDir := filepath.Join(memDir, "data")

	// Create .mem directory and data subdirectory
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create .mem directory: %w", err)
	}

	globalConfigPath := filepath.Join(memDir, "config.yaml")

	// Only create default config if it doesn't exist
	if _, err := os.Stat(globalConfigPath); os.IsNotExist(err) {
		defaultConfig := DefaultConfig()
		if err := WriteConfig(globalConfigPath, defaultConfig); err != nil {
			return fmt.Errorf("failed to write default config: %w", err)
		}
	}

	return nil
}

// WriteConfig writes the configuration to a file in YAML format
func WriteConfig(path string, cfg *Config) error {
	// Use viper to write config
	v := viper.New()
	v.Set("memory", cfg.Memory)
	v.Set("embeddings", cfg.Embeddings)
	v.Set("reranking", cfg.Reranking)
	v.Set("query", cfg.Query)
	v.Set("cli", cfg.CLI)

	// Set config type based on file extension
	ext := filepath.Ext(path)
	if ext == "" {
		ext = ".yaml"
	}
	v.SetConfigType(strings.TrimPrefix(ext, "."))

	if err := v.WriteConfigAs(path); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the path to the global config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".mem", "config.yaml"), nil
}

// GetDataPath returns the path to the data directory
func GetDataPath(cfg *Config) string {
	if cfg != nil && cfg.Memory.Path != "" {
		return cfg.Memory.Path
	}
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".mem", "data")
}
