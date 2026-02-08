package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jbutlerdev/mem/internal/models"
)

// Storage defines the interface for memory storage backends
type Storage interface {
	// Store saves a memory with its embedding
	Store(memory *models.Memory) error

	// Query performs a similarity search using the query embedding
	// Returns results sorted by similarity score (descending)
	Query(queryEmbedding []float32, opts *QueryOptions) ([]models.SearchResult, error)

	// Get retrieves a memory by its ID
	Get(id string) (*models.Memory, error)

	// GetByNamespace retrieves all memories in a namespace
	GetByNamespace(namespace string) ([]*models.Memory, error)

	// GetByTag retrieves all memories with a specific tag
	GetByTag(tag string) ([]*models.Memory, error)

	// Update updates an existing memory
	// If content changes, a new embedding should be provided
	Update(memory *models.Memory) error

	// Delete removes a memory by its ID
	Delete(id string) error

	// DeleteByNamespace removes all memories in a namespace
	DeleteByNamespace(namespace string) error

	// List retrieves memories with optional filtering
	// Namespace can be empty string for all namespaces, tag can be empty for no tag filter
	List(namespace string, tag string, limit int) ([]*models.Memory, error)

	// ListNamespaces returns all namespaces
	ListNamespaces() ([]NamespaceInfo, error)

	// Count returns the number of memories in a namespace
	Count(namespace string) (int, error)

	// CreateNamespace creates a new namespace
	CreateNamespace(namespace string) error

	// DeleteNamespace deletes a namespace and all its memories
	DeleteNamespace(namespace string) error

	// Close closes any open connections or resources
	Close() error
}

// NamespaceInfo contains information about a namespace
type NamespaceInfo struct {
	Name      string
	CreatedAt interface{} // Can be time.Time or string depending on backend
}

// QueryOptions defines options for querying memories
type QueryOptions struct {
	// Namespace restricts search to a specific namespace
	Namespace string

	// Tags restricts search to memories with at least one of these tags
	Tags []string

	// Limit is the maximum number of results to return
	Limit int

	// Threshold is the minimum similarity score (0-1)
	Threshold float32

	// TopK is the number of initial results to fetch before re-ranking
	// If re-ranking is disabled, this is effectively the same as Limit
	TopK int
}

// DefaultQueryOptions returns default query options
func DefaultQueryOptions() *QueryOptions {
	return &QueryOptions{
		Namespace: "",
		Tags:      nil,
		Limit:     5,
		Threshold: 0.6,
		TopK:      10,
	}
}

// NewChromeMStorage creates a new chromem-go storage backend with default settings
func NewChromeMStorage(dataPath string) (Storage, error) {
	return NewChromeMStorageWithDim(dataPath, 1024)
}

// NewChromeMStorageWithDim creates a new chromem-go storage backend with a specific embedding dimension
func NewChromeMStorageWithDim(dataPath string, embeddingDim int) (Storage, error) {
	if dataPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dataPath = filepath.Join(homeDir, ".mem", "data")
	}

	if embeddingDim <= 0 {
		embeddingDim = 1024 // default
	}

	config := &ChromemConfig{
		DataPath:         dataPath,
		CollectionName:   "memories",
		DefaultNamespace: "default",
		EmbeddingDim:     embeddingDim,
		EmbeddingFunc:    nil, // We provide embeddings manually
	}

	return NewChromemStorage(config)
}
