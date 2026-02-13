package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/philippgille/chromem-go"
	"github.com/jbutlerdev/mem/internal/models"
)

// ChromemStorage implements the Storage interface using chromem-go
type ChromemStorage struct {
	mu         sync.RWMutex
	db         *chromem.DB
	collection *chromem.Collection
	dataPath   string
	namespace  string // Default namespace
	embeddingDim int   // Embedding dimension
}

// ChromemConfig holds configuration for chromem-go storage
type ChromemConfig struct {
	// DataPath is the directory where chromem-go stores its data
	DataPath string

	// CollectionName is the name of the collection to use
	CollectionName string

	// DefaultNamespace is the default namespace for memories
	DefaultNamespace string

	// EmbeddingDim is the dimension of embeddings
	EmbeddingDim int

	// EmbeddingFunc is the function to generate embeddings
	// If nil, embeddings must be provided when storing documents
	EmbeddingFunc chromem.EmbeddingFunc
}

// NewChromemStorage creates a new chromem-go storage backend
func NewChromemStorage(config *ChromemConfig) (*ChromemStorage, error) {
	if config == nil {
		config = &ChromemConfig{}
	}

	if config.DataPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		config.DataPath = filepath.Join(homeDir, ".mem", "data")
	}

	if config.CollectionName == "" {
		config.CollectionName = "memories"
	}

	if config.DefaultNamespace == "" {
		config.DefaultNamespace = "default"
	}

	if config.EmbeddingDim == 0 {
		config.EmbeddingDim = 1024 // Default for text-embedding-qwen3-embedding-8b
	}

	// Ensure data directory exists
	if err := os.MkdirAll(config.DataPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create persistent database
	persistDir := filepath.Join(config.DataPath, "chromem")
	db, err := chromem.NewPersistentDB(persistDir, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create chromem database: %w", err)
	}

	// Get or create collection
	// If EmbeddingFunc is provided, chromem-go will generate embeddings automatically
	// Otherwise, we must provide embeddings when adding documents
	// Note: chromem-go doesn't require embedding dimension in GetOrCreateCollection
	collection, err := db.GetOrCreateCollection(config.CollectionName, nil, config.EmbeddingFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create collection: %w", err)
	}

	return &ChromemStorage{
		db:         db,
		collection: collection,
		dataPath:   config.DataPath,
		namespace:  config.DefaultNamespace,
		embeddingDim: config.EmbeddingDim,
	}, nil
}

// Store saves a memory with its embedding
func (s *ChromemStorage) Store(memory *models.Memory) error {
	if err := memory.Validate(); err != nil {
		return fmt.Errorf("invalid memory: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Prepare metadata for chromem-go
	metadata := s.buildMetadata(memory)

	// Create document
	doc := chromem.Document{
		ID:        memory.ID,
		Content:   memory.Content,
		Embedding: memory.Embedding,
		Metadata:  metadata,
	}

	// Store in chromem-go
	err := s.collection.AddDocument(context.Background(), doc)
	if err != nil {
		return fmt.Errorf("failed to add document to collection: %w", err)
	}

	return nil
}

// Query performs a similarity search using the query embedding
func (s *ChromemStorage) Query(queryEmbedding []float32, opts *QueryOptions) ([]models.SearchResult, error) {
	if opts == nil {
		opts = DefaultQueryOptions()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Determine number of results to fetch
	// If TopK is set, fetch that many (for re-ranking), otherwise fetch Limit
	fetchCount := opts.Limit
	if opts.TopK > 0 && opts.TopK > opts.Limit {
		fetchCount = opts.TopK
	}

	// Check if collection is empty or has fewer documents than requested
	count := s.collection.Count()
	if count == 0 {
		// Empty collection - return empty results immediately
		return []models.SearchResult{}, nil
	}
	if fetchCount > count {
		fetchCount = count
	}

	// Build filter for namespace
	where := make(map[string]string)
	if opts.Namespace != "" {
		where["namespace"] = opts.Namespace
	}

	// Query chromem-go with embedding
	results, err := s.collection.QueryEmbedding(context.Background(), queryEmbedding, fetchCount, where, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query collection: %w", err)
	}

	// Convert chromem-go results to our SearchResult format
	searchResults := make([]models.SearchResult, 0, len(results))
	for i, result := range results {
		// Apply threshold filter
		if float32(result.Similarity) < opts.Threshold {
			continue
		}

		// Apply tag filter (post-processing)
		if len(opts.Tags) > 0 {
			if !s.hasAnyTag(result.Metadata, opts.Tags) {
				continue
			}
		}

		// Reconstruct memory from result
		memory, err := s.resultToMemory(result)
		if err != nil {
			// Skip this result if we can't reconstruct it
			continue
		}

		searchResults = append(searchResults, models.SearchResult{
			Memory:   *memory,
			Score:    float32(result.Similarity),
			Reranked: false, // Will be set to true by reranker if used
			Rank:     i + 1,
		})
	}

	// Apply limit (in case TopK was used and we got more after filtering)
	if len(searchResults) > opts.Limit {
		searchResults = searchResults[:opts.Limit]
	}

	return searchResults, nil
}

// Get retrieves a memory by its ID
func (s *ChromemStorage) Get(id string) (*models.Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Use GetByID to retrieve the document
	doc, err := s.collection.GetByID(context.Background(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	memory, err := s.docToMemory(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to convert document to memory: %w", err)
	}

	return memory, nil
}

// GetByNamespace retrieves all memories in a namespace
func (s *ChromemStorage) GetByNamespace(namespace string) ([]*models.Memory, error) {
	return s.List(namespace, "", 10000) // High limit to get all
}

// GetByTag retrieves all memories with a specific tag
func (s *ChromemStorage) GetByTag(tag string) ([]*models.Memory, error) {
	return s.List("", tag, 10000) // High limit to get all
}

// Update updates an existing memory
func (s *ChromemStorage) Update(memory *models.Memory) error {
	if err := memory.Validate(); err != nil {
		return fmt.Errorf("invalid memory: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if memory exists
	existing, err := s.collection.GetByID(context.Background(), memory.ID)
	if err != nil {
		return fmt.Errorf("memory not found: %w", err)
	}

	// Update timestamps
	memory.CreatedAt, _ = time.Parse(time.RFC3339, existing.Metadata["created_at"])
	memory.UpdatedAt = time.Now()

	// Create updated document
	metadata := s.buildMetadata(memory)
	doc := chromem.Document{
		ID:        memory.ID,
		Content:   memory.Content,
		Embedding: memory.Embedding,
		Metadata:  metadata,
	}

	// chromem-go doesn't have a direct Update method
	// We need to delete and re-add
	// Delete old version
	err = s.collection.Delete(context.Background(), nil, nil, memory.ID)
	if err != nil {
		return fmt.Errorf("failed to delete old document: %w", err)
	}

	// Add new version
	err = s.collection.AddDocument(context.Background(), doc)
	if err != nil {
		return fmt.Errorf("failed to add updated document: %w", err)
	}

	return nil
}

// UpdateEmbedding updates only the embedding for an existing memory
// This is useful when regenerating embeddings with a new model
func (s *ChromemStorage) UpdateEmbedding(id string, embedding []float32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if memory exists
	existing, err := s.collection.GetByID(context.Background(), id)
	if err != nil {
		return fmt.Errorf("memory not found: %w", err)
	}

	// Parse existing metadata to reconstruct memory
	// Parse tags from JSON
	var tags []string
	if tagsJSON, ok := existing.Metadata["tags"]; ok && tagsJSON != "" {
		json.Unmarshal([]byte(tagsJSON), &tags)
	}

	memory := &models.Memory{
		ID:        id,
		Content:   existing.Content,
		Embedding: embedding,
		Namespace: existing.Metadata["namespace"],
		Tags:      tags,
	}

	if createdAt, ok := existing.Metadata["created_at"]; ok {
		memory.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	}
	memory.UpdatedAt = time.Now()

	// Create updated document
	metadata := s.buildMetadata(memory)
	doc := chromem.Document{
		ID:        id,
		Content:   existing.Content,
		Embedding: embedding,
		Metadata:  metadata,
	}

	// Delete old version
	err = s.collection.Delete(context.Background(), nil, nil, id)
	if err != nil {
		return fmt.Errorf("failed to delete old document: %w", err)
	}

	// Add new version with new embedding
	err = s.collection.AddDocument(context.Background(), doc)
	if err != nil {
		return fmt.Errorf("failed to add updated document: %w", err)
	}

	return nil
}

// Delete removes a memory by its ID
func (s *ChromemStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.collection.Delete(context.Background(), nil, nil, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// DeleteByNamespace removes all memories in a namespace
func (s *ChromemStorage) DeleteByNamespace(namespace string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Delete all documents with this namespace
	err := s.collection.Delete(context.Background(), map[string]string{"namespace": namespace}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	return nil
}

// List retrieves memories with optional filtering
// Namespace can be empty string for all namespaces, tag can be empty for no tag filter
// Limit of 0 means no limit
func (s *ChromemStorage) List(namespace string, tag string, limit int) ([]*models.Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// chromem-go doesn't have a direct List/GetAll method
	// We'll use a workaround: query with a dummy embedding to get all documents
	// and filter by metadata. This is a limitation of chromem-go.

	// Create a minimal embedding (all zeros) for the query
	// This won't give meaningful similarity scores, but will return matching documents
	embeddingDim := s.embeddingDim
	if embeddingDim == 0 {
		// Default to 1024 if we can't get the dimension
		embeddingDim = 1024
	}
	dummyEmbedding := make([]float32, embeddingDim)

	// Build where clause for namespace filter
	where := make(map[string]string)
	if namespace != "" {
		where["namespace"] = namespace
	}

	// Fetch all documents (use the count or a high limit)
	count := s.collection.Count()
	fetchLimit := count
	if fetchLimit == 0 {
		// Empty collection - return empty results immediately
		return []*models.Memory{}, nil
	}
	results, err := s.collection.QueryEmbedding(context.Background(), dummyEmbedding, fetchLimit, where, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query collection: %w", err)
	}

	// Convert results to memories
	memories := make([]*models.Memory, 0, len(results))
	for _, result := range results {
		memory, err := s.resultToMemory(result)
		if err != nil {
			// Skip this result if we can't reconstruct it
			continue
		}

		// Apply tag filter
		if tag != "" && !memory.HasTag(tag) {
			continue
		}

		memories = append(memories, memory)
	}

	// Apply limit
	if limit > 0 && limit < len(memories) {
		memories = memories[:limit]
	}

	return memories, nil
}

// GetAll returns all memories regardless of embedding dimension
// This is useful for regeneration when embeddings may be corrupted
// It uses text-based query to avoid embedding dimension issues
func (s *ChromemStorage) GetAll() ([]*models.Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get the count of documents
	count := s.collection.Count()
	if count == 0 {
		return []*models.Memory{}, nil
	}

	// Use text query with empty string to get all documents
	// This bypasses embedding-based similarity
	results, err := s.collection.Query(context.Background(), "", count, nil, nil)
	if err != nil {
		// If text query fails, try embedding-based query with current dimension
		dummyEmbedding := make([]float32, s.embeddingDim)
		results, err = s.collection.QueryEmbedding(context.Background(), dummyEmbedding, count, nil, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to get all memories: %w", err)
		}
	}

	// Convert results to memories
	memories := make([]*models.Memory, 0, len(results))
	for _, result := range results {
		memory, err := s.resultToMemory(result)
		if err != nil {
			continue
		}
		memories = append(memories, memory)
	}

	return memories, nil
}

// ListNamespaces returns all namespaces
func (s *ChromemStorage) ListNamespaces() ([]NamespaceInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get all memories to extract unique namespaces
	memories, err := s.List("", "", 0)
	if err != nil {
		return nil, err
	}

	// Extract unique namespaces
	namespaceMap := make(map[string]time.Time)
	for _, mem := range memories {
		if _, exists := namespaceMap[mem.Namespace]; !exists {
			namespaceMap[mem.Namespace] = mem.CreatedAt
		}
	}

	// Convert to NamespaceInfo slice
	namespaces := make([]NamespaceInfo, 0, len(namespaceMap))
	for name, createdAt := range namespaceMap {
		namespaces = append(namespaces, NamespaceInfo{
			Name:      name,
			CreatedAt: createdAt,
		})
	}

	return namespaces, nil
}

// Count returns the number of memories in a namespace
func (s *ChromemStorage) Count(namespace string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Get memories in namespace
	memories, err := s.List(namespace, "", 0)
	if err != nil {
		return 0, err
	}

	return len(memories), nil
}

// CreateNamespace creates a new namespace
func (s *ChromemStorage) CreateNamespace(namespace string) error {
	// In chromem-go, namespaces are implicit (they exist when memories are stored)
	// This is a no-op, but we validate the namespace name
	if namespace == "" {
		return fmt.Errorf("namespace name cannot be empty")
	}

	// Check if namespace already exists
	namespaces, err := s.ListNamespaces()
	if err != nil {
		return err
	}

	for _, ns := range namespaces {
		if ns.Name == namespace {
			return nil // Already exists, that's fine
		}
	}

	// Namespace doesn't exist yet - it will be created when first memory is stored
	// We could optionally store a marker memory, but for now this is a no-op
	return nil
}

// DeleteNamespace deletes a namespace and all its memories
func (s *ChromemStorage) DeleteNamespace(namespace string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if namespace == "" {
		return fmt.Errorf("namespace name cannot be empty")
	}

	// Delete all documents with this namespace
	err := s.collection.Delete(context.Background(), map[string]string{"namespace": namespace}, nil)
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	return nil
}

// Close closes any open connections or resources
func (s *ChromemStorage) Close() error {
	// chromem-go handles persistence automatically
	// No explicit close needed
	return nil
}

// Helper methods

// buildMetadata creates metadata map for chromem-go from a memory
func (s *ChromemStorage) buildMetadata(memory *models.Memory) map[string]string {
	metadata := make(map[string]string)

	// Store core fields
	metadata["namespace"] = memory.Namespace
	metadata["created_at"] = memory.CreatedAt.Format(time.RFC3339)
	metadata["updated_at"] = memory.UpdatedAt.Format(time.RFC3339)

	// Store tags as JSON string in metadata
	if len(memory.Tags) > 0 {
		tagsJSON, _ := json.Marshal(memory.Tags)
		metadata["tags"] = string(tagsJSON)
	}

	// Store custom metadata as JSON string
	if len(memory.Metadata) > 0 {
		customMetaJSON, _ := json.Marshal(memory.Metadata)
		metadata["metadata"] = string(customMetaJSON)
	}

	return metadata
}

// hasAnyTag checks if the metadata contains any of the specified tags
func (s *ChromemStorage) hasAnyTag(metadata map[string]string, tags []string) bool {
	tagsJSON, ok := metadata["tags"]
	if !ok || tagsJSON == "" {
		return false
	}

	var docTags []string
	if err := json.Unmarshal([]byte(tagsJSON), &docTags); err != nil {
		return false
	}

	for _, tag := range tags {
		for _, docTag := range docTags {
			if docTag == tag {
				return true
			}
		}
	}

	return false
}

// resultToMemory converts a chromem-go query result to a Memory
func (s *ChromemStorage) resultToMemory(result chromem.Result) (*models.Memory, error) {
	return s.docToMemory(chromem.Document{
		ID:        result.ID,
		Content:   result.Content,
		Metadata:  result.Metadata,
		Embedding: nil, // Not included in query results
	})
}

// docToMemory converts a chromem-go Document to a Memory
func (s *ChromemStorage) docToMemory(doc chromem.Document) (*models.Memory, error) {
	// Parse metadata
	namespace := doc.Metadata["namespace"]
	if namespace == "" {
		namespace = s.namespace
	}

	// Parse timestamps
	createdAt, _ := time.Parse(time.RFC3339, doc.Metadata["created_at"])
	updatedAt, _ := time.Parse(time.RFC3339, doc.Metadata["updated_at"])

	// Parse tags
	var tags []string
	if tagsJSON, ok := doc.Metadata["tags"]; ok && tagsJSON != "" {
		json.Unmarshal([]byte(tagsJSON), &tags)
	}

	// Parse custom metadata
	var customMetadata map[string]interface{}
	if metadataJSON, ok := doc.Metadata["metadata"]; ok && metadataJSON != "" {
		json.Unmarshal([]byte(metadataJSON), &customMetadata)
	}

	// Create memory
	memory := &models.Memory{
		ID:        doc.ID,
		Content:   doc.Content,
		Embedding: doc.Embedding,
		Namespace: namespace,
		Tags:      tags,
		Metadata:  customMetadata,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return memory, nil
}