package cli

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jbutlerdev/mem/internal/config"
	"github.com/jbutlerdev/mem/internal/models"
	"github.com/spf13/viper"
)

// setupTestConfig creates a test configuration in a temporary directory
func setupTestConfig(t *testing.T) (string, *config.Config) {
	t.Helper()
	
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create a test config
	testConfig := config.DefaultConfig()
	testConfig.Memory.Path = tmpDir
	testConfig.Memory.Backend = "chromem"
	testConfig.Embeddings.BaseURL = "http://localhost:8080"
	testConfig.Embeddings.Model = "test-model"
	testConfig.Embeddings.Dimensions = 768
	testConfig.CLI.Verbose = false
	testConfig.CLI.OutputFormat = "text"

	// Write config file using viper
	v := viper.New()
	v.Set("memory.path", tmpDir)
	v.Set("memory.backend", "chromem")
	v.Set("memory.default_namespace", "default")
	v.Set("embeddings.base_url", "http://localhost:8080")
	v.Set("embeddings.model", "test-model")
	v.Set("embeddings.dimensions", 768)
	v.Set("cli.verbose", false)
	v.Set("cli.output_format", "text")

	return configPath, testConfig
}

// createTestMemory creates a test memory with default values
func createTestMemory(t *testing.T) *models.Memory {
	t.Helper()
	
	now := time.Now()
	return &models.Memory{
		ID:        "test-" + now.Format("20060102150405"),
		Content:   "Test memory content",
		Embedding: make([]float32, 768),
		Namespace: "test",
		Tags:      []string{"test", "sample"},
		Metadata:  map[string]interface{}{"key": "value"},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// createTestMemoryWithNamespace creates a test memory with a specific namespace
func createTestMemoryWithNamespace(t *testing.T, namespace, content string) *models.Memory {
	t.Helper()
	
	now := time.Now()
	mem := &models.Memory{
		ID:        "test-" + namespace + "-" + now.Format("20060102150405"),
		Content:   content,
		Embedding: make([]float32, 768),
		Namespace: namespace,
		Tags:      []string{"test"},
		CreatedAt: now,
		UpdatedAt: now,
	}
	
	// Set a non-zero embedding value for testing
	for i := range mem.Embedding {
		mem.Embedding[i] = 0.1
	}
	
	return mem
}

// captureOutput captures stdout during a test
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	
	fn()
	
	w.Close()
	os.Stdout = old
	
	var buf []byte
	buf, _ = io.ReadAll(r)
	return string(buf)
}

// captureStderr captures stderr during a test
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	
	fn()
	
	w.Close()
	os.Stderr = old
	
	var buf []byte
	buf, _ = io.ReadAll(r)
	return string(buf)
}