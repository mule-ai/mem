package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// No default namespace - empty means all namespaces
	if cfg.Memory.DefaultNamespace != "" {
		t.Errorf("Expected default namespace '', got '%s'", cfg.Memory.DefaultNamespace)
	}

	if cfg.Memory.Backend != "chromem" {
		t.Errorf("Expected backend 'chromem', got '%s'", cfg.Memory.Backend)
	}

	if cfg.Embeddings.BaseURL != "http://10.10.199.29:8080" {
		t.Errorf("Expected base URL 'http://10.10.199.29:8080', got '%s'", cfg.Embeddings.BaseURL)
	}

	if cfg.Embeddings.Model != "text-embedding-qwen3-embedding-8b" {
		t.Errorf("Expected model 'text-embedding-qwen3-embedding-8b', got '%s'", cfg.Embeddings.Model)
	}

	if cfg.Reranking.Model != "qwen3-reranker-8b" {
		t.Errorf("Expected reranker model 'qwen3-reranker-8b', got '%s'", cfg.Reranking.Model)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid backend",
			cfg: &Config{
				Memory: MemoryConfig{
					Backend: "invalid",
				},
				Embeddings: EmbeddingsConfig{
					BaseURL: "http://example.com",
					Model:   "test-model",
					Dimensions: 1024,
				},
			},
			wantErr: true,
		},
		{
			name: "empty base URL",
			cfg: &Config{
				Memory: MemoryConfig{
					Backend: "chromem",
				},
				Embeddings: EmbeddingsConfig{
					BaseURL: "",
					Model:   "test-model",
					Dimensions: 1024,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfigWithoutFiles(t *testing.T) {
	// Create a temp directory to ensure no config files exist
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)

	os.Chdir(tmpDir)

	// Load config without any files - should use defaults
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify defaults are applied
	if cfg.Memory.DefaultNamespace != "default" {
		t.Errorf("Expected default namespace 'default', got '%s'", cfg.Memory.DefaultNamespace)
	}
}

func TestInitConfig(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)

	// Set HOME to temp directory
	os.Setenv("HOME", tmpDir)

	err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() failed: %v", err)
	}

	// Verify config file was created
	configPath := filepath.Join(tmpDir, ".mem", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created at %s", configPath)
	}

	// Verify data directory was created
	dataPath := filepath.Join(tmpDir, ".mem", "data")
	if _, err := os.Stat(dataPath); os.IsNotExist(err) {
		t.Errorf("Data directory was not created at %s", dataPath)
	}

	// Load and verify the created config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// DefaultConfig creates config with empty default namespace (all namespaces)
	if cfg.Memory.DefaultNamespace != "" {
		t.Errorf("Expected default namespace '', got '%s'", cfg.Memory.DefaultNamespace)
	}
}

func TestGetDataPath(t *testing.T) {
	cfg := &Config{
		Memory: MemoryConfig{
			Path: "/custom/path",
		},
	}

	path := GetDataPath(cfg)
	if path != "/custom/path" {
		t.Errorf("Expected '/custom/path', got '%s'", path)
	}

	// Test with nil config
	path = GetDataPath(nil)
	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".mem", "data")
	if path != expectedPath {
		t.Errorf("Expected '%s', got '%s'", expectedPath, path)
	}
}

func TestConfigWithLocalOverride(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	originalDir, _ := os.Getwd()
	defer os.Setenv("HOME", originalHome)
	defer os.Chdir(originalDir)

	os.Setenv("HOME", tmpDir)
	os.Chdir(tmpDir)

	// Create global config
	os.MkdirAll(filepath.Join(tmpDir, ".mem"), 0755)
	globalConfig := filepath.Join(tmpDir, ".mem", "config.yaml")
	cfg := DefaultConfig()
	cfg.Memory.DefaultNamespace = "global"
	if err := WriteConfig(globalConfig, cfg); err != nil {
		t.Fatalf("Failed to write global config: %v", err)
	}

	// Create local override - write as YAML manually since WriteConfig needs an extension
	localConfig := filepath.Join(tmpDir, ".memconfig")
	localYAML := "memory:\n  default_namespace: local\n"
	if err := os.WriteFile(localConfig, []byte(localYAML), 0644); err != nil {
		t.Fatalf("Failed to write local config: %v", err)
	}

	// Load and verify local override takes precedence
	loaded, err := Load("")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if loaded.Memory.DefaultNamespace != "local" {
		t.Errorf("Expected namespace 'local' (from override), got '%s'", loaded.Memory.DefaultNamespace)
	}
}
