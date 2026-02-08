package storage

import (
	"fmt"
	"testing"
)

// TestPostgresStorage_NewWithoutDB tests that NewPostgresStorage fails gracefully without a database
func TestPostgresStorage_NewWithoutDB(t *testing.T) {
	config := &PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "nonexistent_test_db",
		User:     "postgres",
		Password: "",
		SSLMode:  "disable",
	}

	storage, err := NewPostgresStorage(config)

	// Should fail because database doesn't exist
	if err == nil {
		storage.Close()
		t.Error("Expected error when connecting to non-existent database, got nil")
	}
}

// TestPostgresStorage_NilConfig tests that nil config is handled
func TestPostgresStorage_NilConfig(t *testing.T) {
	storage, err := NewPostgresStorage(nil)

	if err == nil {
		storage.Close()
		t.Error("Expected error for nil config, got nil")
	}

	if storage != nil {
		t.Error("Expected nil storage for nil config")
	}
}

// TestPostgresStorage_EmbeddingToString tests the embedding string conversion
func TestPostgresStorage_EmbeddingToString(t *testing.T) {
	// We can't test without a real database, but we can test the helper
	storage := &PostgresStorage{
		embeddingDim: 3,
	}

	tests := []struct {
		name      string
		embedding []float32
		expected  string
	}{
		{
			name:      "empty embedding",
			embedding: []float32{},
			expected:  "[]",
		},
		{
			name:      "single value",
			embedding: []float32{1.5},
			expected:  "[1.500000]",
		},
		{
			name:      "multiple values",
			embedding: []float32{1.0, 2.5, -3.7},
			expected:  "[1.000000,2.500000,-3.700000]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage.embeddingToString(tt.embedding)
			if result != tt.expected {
				t.Errorf("embeddingToString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestPostgresStorage_DefaultConfig tests default value assignment
func TestPostgresStorage_DefaultConfig(t *testing.T) {
	config := &PostgresConfig{
		User:     "testuser",
		Database: "testdb",
	}

	if config.Host == "" {
		config.Host = "localhost" // Should be set by NewPostgresStorage
	}
	if config.Port == 0 {
		config.Port = 5432
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	if config.EmbeddingDim == 0 {
		config.EmbeddingDim = 1024
	}
	if config.DefaultNamespace == "" {
		config.DefaultNamespace = "default"
	}

	// Verify defaults were set
	if config.Host != "localhost" {
		t.Errorf("Expected host to be localhost, got %s", config.Host)
	}
	if config.Port != 5432 {
		t.Errorf("Expected port to be 5432, got %d", config.Port)
	}
	if config.SSLMode != "disable" {
		t.Errorf("Expected sslmode to be disable, got %s", config.SSLMode)
	}
	if config.EmbeddingDim != 1024 {
		t.Errorf("Expected embedding dim to be 1024, got %d", config.EmbeddingDim)
	}
	if config.DefaultNamespace != "default" {
		t.Errorf("Expected default namespace to be default, got %s", config.DefaultNamespace)
	}
}

// TestPostgresStorage_ConnectionString tests connection string building
func TestPostgresStorage_ConnectionString(t *testing.T) {
	config := &PostgresConfig{
		Host:     "testhost",
		Port:     5433,
		Database: "testdb",
		User:     "testuser",
		Password: "testpass",
		SSLMode:  "require",
	}

	expected := "host=testhost port=5433 dbname=testdb user=testuser password=testpass sslmode=require"
	actual := formatConnStr(config)

	if actual != expected {
		t.Errorf("Connection string = %v, want %v", actual, expected)
	}
}

// Helper function to test connection string formatting
func formatConnStr(config *PostgresConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		config.Host,
		config.Port,
		config.Database,
		config.User,
		config.Password,
		config.SSLMode,
	)
}

// Note: Full integration tests require a running PostgreSQL instance with pgvector
// These would typically be run in a CI environment with docker-compose

/*
Integration test example (requires PostgreSQL with pgvector):

func TestPostgresStorage_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := &PostgresConfig{
		Host:         "localhost",
		Port:         5432,
		Database:     "mem_test",
		User:         "postgres",
		Password:     "postgres",
		SSLMode:      "disable",
		EmbeddingDim: 1024,
	}

	storage, err := NewPostgresStorage(config)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Test Store
	memory := &models.Memory{
		ID:        "test-1",
		Content:   "Test memory content",
		Embedding: make([]float32, 1024),
		Namespace: "test",
		Tags:      []string{"test", "unit"},
		Metadata:  map[string]interface{}{"key": "value"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = storage.Store(memory)
	if err != nil {
		t.Fatalf("Failed to store memory: %v", err)
	}

	// Test Get
	retrieved, err := storage.Get("test-1")
	if err != nil {
		t.Fatalf("Failed to get memory: %v", err)
	}

	if retrieved.Content != memory.Content {
		t.Errorf("Expected content %s, got %s", memory.Content, retrieved.Content)
	}

	// Test Delete
	err = storage.Delete("test-1")
	if err != nil {
		t.Fatalf("Failed to delete memory: %v", err)
	}

	// Verify deletion
	_, err = storage.Get("test-1")
	if err == nil {
		t.Error("Expected error when getting deleted memory")
	}
}
*/
