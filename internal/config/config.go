package config

import (
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Memory     MemoryConfig     `mapstructure:"memory"`
	Embeddings EmbeddingsConfig `mapstructure:"embeddings"`
	Reranking  RerankingConfig  `mapstructure:"reranking"`
	Query      QueryConfig      `mapstructure:"query"`
	CLI        CLIConfig        `mapstructure:"cli"`
}

// MemoryConfig contains database and storage settings
type MemoryConfig struct {
	Path             string         `mapstructure:"path"`
	DefaultNamespace string         `mapstructure:"default_namespace"`
	Backend          string         `mapstructure:"backend"`
	Postgres         PostgresConfig `mapstructure:"postgres"`
}

// PostgresConfig contains PostgreSQL connection settings
type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"sslmode"`
}

// EmbeddingsConfig contains embedding API settings
type EmbeddingsConfig struct {
	BaseURL    string `mapstructure:"base_url,BaseURL,baseurl"`
	Model      string `mapstructure:"model"`
	APIKey     string `mapstructure:"api_key,APIKey,apikey"`
	Dimensions int    `mapstructure:"dimensions,Dimensions,dimensions"`
	BatchSize  int    `mapstructure:"batch_size,BatchSize,batchsize"`
}

// RerankingConfig contains reranker settings
type RerankingConfig struct {
	Enabled   bool    `mapstructure:"enabled"`
	Model     string  `mapstructure:"model"`
	TopK      int     `mapstructure:"top_k"`
	Threshold float64 `mapstructure:"threshold"`
}

// QueryConfig contains query behavior settings
type QueryConfig struct {
	DefaultLimit  int     `mapstructure:"default_limit"`
	MinSimilarity float64 `mapstructure:"min_similarity"`
	RerankEnabled bool    `mapstructure:"rerank_enabled"`
}

// CLIConfig contains CLI behavior settings
type CLIConfig struct {
	OutputFormat string `mapstructure:"output_format"`
	Verbose      bool   `mapstructure:"verbose"`
	Color        bool   `mapstructure:"color"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Memory: MemoryConfig{
			Path:             filepath.Join(homeDir, ".mem", "data"),
			// No default namespace - empty means all namespaces
			Backend:          "chromem",
			Postgres: PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "mem",
				User:     "postgres",
				Password: "",
				SSLMode:  "disable",
			},
		},
		Embeddings: EmbeddingsConfig{
			BaseURL:    "http://10.10.199.29:8080",
			Model:      "text-embedding-qwen3-embedding-8b",
			APIKey:     "",
			Dimensions: 1024,
			BatchSize:  10,
		},
		Reranking: RerankingConfig{
			Enabled:   true,
			Model:     "qwen3-reranker-8b",
			TopK:      10,
			Threshold: 0.5,
		},
		Query: QueryConfig{
			DefaultLimit:  5,
			MinSimilarity: 0.6,
			RerankEnabled: true,
		},
		CLI: CLIConfig{
			OutputFormat: "text",
			Verbose:      false,
			Color:        true,
		},
	}
}

// GetPostgresConnectionString returns the PostgreSQL connection string
func (c *PostgresConfig) GetPostgresConnectionString() string {
	return c.getConnectionString(nil)
}

// getConnectionString builds a PostgreSQL connection string with optional overrides
func (c *PostgresConfig) getConnectionString(overrides map[string]string) string {
	host := c.Host
	port := c.Port
	database := c.Database
	user := c.User
	password := c.Password
	sslmode := c.SSLMode

	if overrides != nil {
		if v, ok := overrides["host"]; ok && v != "" {
			host = v
		}
		if v, ok := overrides["port"]; ok && v != "" {
			// Parse port from string
		}
		if v, ok := overrides["database"]; ok && v != "" {
			database = v
		}
		if v, ok := overrides["user"]; ok && v != "" {
			user = v
		}
		if v, ok := overrides["password"]; ok && v != "" {
			password = v
		}
		if v, ok := overrides["sslmode"]; ok && v != "" {
			sslmode = v
		}
	}

	connStr := ""
	if password != "" {
		connStr = "host=" + host + " port=" + string(rune(port)) + " user=" + user + " password=" + password + " dbname=" + database + " sslmode=" + sslmode
	} else {
		connStr = "host=" + host + " port=" + string(rune(port)) + " user=" + user + " dbname=" + database + " sslmode=" + sslmode
	}
	return connStr
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate backend
	if c.Memory.Backend != "chromem" && c.Memory.Backend != "postgres" {
		return &ValidationError{Field: "memory.backend", Message: "must be 'chromem' or 'postgres'"}
	}

	// Validate embeddings settings
	if c.Embeddings.BaseURL == "" {
		return &ValidationError{Field: "embeddings.base_url", Message: "cannot be empty"}
	}
	if c.Embeddings.Model == "" {
		return &ValidationError{Field: "embeddings.model", Message: "cannot be empty"}
	}
	if c.Embeddings.Dimensions <= 0 {
		return &ValidationError{Field: "embeddings.dimensions", Message: "must be positive"}
	}

	// Validate reranking settings
	if c.Reranking.Enabled {
		if c.Reranking.Model == "" {
			return &ValidationError{Field: "reranking.model", Message: "cannot be empty when reranking is enabled"}
		}
		if c.Reranking.TopK <= 0 {
			return &ValidationError{Field: "reranking.top_k", Message: "must be positive"}
		}
	}

	// Validate query settings
	if c.Query.DefaultLimit <= 0 {
		return &ValidationError{Field: "query.default_limit", Message: "must be positive"}
	}
	if c.Query.MinSimilarity < 0 || c.Query.MinSimilarity > 1 {
		return &ValidationError{Field: "query.min_similarity", Message: "must be between 0 and 1"}
	}

	return nil
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return "config validation failed: " + e.Field + " - " + e.Message
}