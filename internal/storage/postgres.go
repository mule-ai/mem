package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jbutlerdev/mem/internal/models"
	_ "github.com/lib/pq"
)

// PostgresStorage implements the Storage interface using PostgreSQL with pgvector
type PostgresStorage struct {
	db               *sql.DB
	embeddingDim     int
	defaultNamespace string
}

// PostgresConfig holds configuration for PostgreSQL storage
type PostgresConfig struct {
	Host             string
	Port             int
	Database         string
	User             string
	Password         string
	SSLMode          string
	EmbeddingDim     int
	DefaultNamespace string
}

// NewPostgresStorage creates a new PostgreSQL storage backend
func NewPostgresStorage(config *PostgresConfig) (*PostgresStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("postgres config cannot be nil")
	}

	// Set defaults
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 5432
	}
	if config.Database == "" {
		config.Database = "mem"
	}
	if config.User == "" {
		config.User = "postgres"
	}
	if config.SSLMode == "" {
		config.SSLMode = "disable"
	}
	if config.EmbeddingDim == 0 {
		config.EmbeddingDim = 1024 // Default for text-embedding-qwen3-embedding-8b
	}
	if config.DefaultNamespace == "" {
		config.DefaultNamespace = "default"
	}

	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		config.Host,
		config.Port,
		config.Database,
		config.User,
		config.Password,
		config.SSLMode,
	)

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	storage := &PostgresStorage{
		db:               db,
		embeddingDim:     config.EmbeddingDim,
		defaultNamespace: config.DefaultNamespace,
	}

	// Initialize schema
	if err := storage.initSchema(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// initSchema creates the necessary tables and extensions
func (s *PostgresStorage) initSchema(ctx context.Context) error {
	// Enable pgvector extension
	_, err := s.db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		return fmt.Errorf("failed to create pgvector extension: %w", err)
	}

	// Create memories table
	_, err = s.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS memories (
			id TEXT PRIMARY KEY,
			content TEXT NOT NULL,
			embedding vector(1024) NOT NULL,
			namespace TEXT NOT NULL DEFAULT 'default',
			tags TEXT[],
			metadata JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create memories table: %w", err)
	}

	// Create indexes for efficient querying
	// HNSW index for vector similarity search
	_, err = s.db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS memories_embedding_idx 
		ON memories USING hnsw (embedding vector_cosine_ops)
	`)
	if err != nil {
		return fmt.Errorf("failed to create embedding index: %w", err)
	}

	// Index for namespace filtering
	_, err = s.db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS memories_namespace_idx 
		ON memories(namespace)
	`)
	if err != nil {
		return fmt.Errorf("failed to create namespace index: %w", err)
	}

	// GIN index for tag filtering
	_, err = s.db.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS memories_tags_idx 
		ON memories USING GIN (tags)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tags index: %w", err)
	}

	return nil
}

// Store saves a memory with its embedding
func (s *PostgresStorage) Store(memory *models.Memory) error {
	if err := memory.Validate(); err != nil {
		return fmt.Errorf("invalid memory: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Convert embedding to pgvector format
	embeddingStr := s.embeddingToString(memory.Embedding)

	// Convert tags to array
	tagsArray := "{" + stringSliceToCSV(memory.Tags) + "}"
	if len(memory.Tags) == 0 {
		tagsArray = "{}"
	}

	// Convert metadata to JSON
	var metadataJSON []byte
	if len(memory.Metadata) > 0 {
		var err error
		metadataJSON, err = json.Marshal(memory.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	// Set namespace
	namespace := memory.Namespace
	if namespace == "" {
		namespace = s.defaultNamespace
	}

	// Set timestamps
	now := time.Now()
	createdAt := memory.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	updatedAt := memory.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = now
	}

	// Insert memory
	query := `
		INSERT INTO memories (id, content, embedding, namespace, tags, metadata, created_at, updated_at)
		VALUES ($1, $2, $3::vector, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			embedding = EXCLUDED.embedding,
			namespace = EXCLUDED.namespace,
			tags = EXCLUDED.tags,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
	`

	_, err := s.db.ExecContext(ctx, query,
		memory.ID,
		memory.Content,
		embeddingStr,
		namespace,
		tagsArray,
		metadataJSON,
		createdAt,
		updatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to store memory: %w", err)
	}

	return nil
}

// Query performs a similarity search using the query embedding
func (s *PostgresStorage) Query(queryEmbedding []float32, opts *QueryOptions) ([]models.SearchResult, error) {
	if opts == nil {
		opts = DefaultQueryOptions()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Determine number of results to fetch
	fetchCount := opts.Limit
	if opts.TopK > 0 && opts.TopK > opts.Limit {
		fetchCount = opts.TopK
	}

	// Convert embedding to pgvector format
	embeddingStr := s.embeddingToString(queryEmbedding)

	// Build query with optional filters
	query := `
		SELECT 
			id, content, namespace, tags, metadata, created_at, updated_at,
			1 - (embedding <=> $1::vector) as similarity
		FROM memories
		WHERE 1 - (embedding <=> $1::vector) >= $2
	`

	args := []interface{}{embeddingStr, opts.Threshold}
	argIdx := 3

	// Add namespace filter
	if opts.Namespace != "" {
		query += fmt.Sprintf(" AND namespace = $%d", argIdx)
		args = append(args, opts.Namespace)
		argIdx++
	}

	// Add tag filter (any of the specified tags)
	if len(opts.Tags) > 0 {
		query += fmt.Sprintf(" AND tags && $%d", argIdx)
		args = append(args, opts.Tags)
		argIdx++
	}

	// Order by similarity and limit
	query += fmt.Sprintf(" ORDER BY similarity DESC LIMIT $%d", argIdx)
	args = append(args, fetchCount)

	// Execute query
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	// Parse results
	var results []models.SearchResult
	for rows.Next() {
		var memory models.Memory
		var similarity float32
		var tagsArray []string
		var metadataJSON []byte
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&memory.ID,
			&memory.Content,
			&memory.Namespace,
			&tagsArray,
			&metadataJSON,
			&createdAt,
			&updatedAt,
			&similarity,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		memory.Tags = tagsArray
		memory.CreatedAt = createdAt
		memory.UpdatedAt = updatedAt

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &memory.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		results = append(results, models.SearchResult{
			Memory:   memory,
			Score:    similarity,
			Reranked: false,
			Rank:     len(results) + 1,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Apply limit (in case TopK was used)
	if len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// Get retrieves a memory by its ID
func (s *PostgresStorage) Get(id string) (*models.Memory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, content, namespace, tags, metadata, created_at, updated_at
		FROM memories
		WHERE id = $1
	`

	var memory models.Memory
	var tagsArray []string
	var metadataJSON []byte
	var createdAt, updatedAt time.Time

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&memory.ID,
		&memory.Content,
		&memory.Namespace,
		&tagsArray,
		&metadataJSON,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("memory not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get memory: %w", err)
	}

	memory.Tags = tagsArray
	memory.CreatedAt = createdAt
	memory.UpdatedAt = updatedAt

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &memory.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &memory, nil
}

// GetByNamespace retrieves all memories in a namespace
func (s *PostgresStorage) GetByNamespace(namespace string) ([]*models.Memory, error) {
	return s.List(namespace, "", 10000)
}

// GetByTag retrieves all memories with a specific tag
func (s *PostgresStorage) GetByTag(tag string) ([]*models.Memory, error) {
	return s.List("", tag, 10000)
}

// Update updates an existing memory
func (s *PostgresStorage) Update(memory *models.Memory) error {
	if err := memory.Validate(); err != nil {
		return fmt.Errorf("invalid memory: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if memory exists
	existing, err := s.Get(memory.ID)
	if err != nil {
		return fmt.Errorf("memory not found: %w", err)
	}

	// Preserve created_at
	memory.CreatedAt = existing.CreatedAt
	memory.UpdatedAt = time.Now()

	// Convert embedding to pgvector format
	embeddingStr := s.embeddingToString(memory.Embedding)

	// Convert tags to array
	tagsArray := "{" + stringSliceToCSV(memory.Tags) + "}"
	if len(memory.Tags) == 0 {
		tagsArray = "{}"
	}

	// Convert metadata to JSON
	var metadataJSON []byte
	if len(memory.Metadata) > 0 {
		metadataJSON, err = json.Marshal(memory.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	// Set namespace
	namespace := memory.Namespace
	if namespace == "" {
		namespace = s.defaultNamespace
	}

	// Update memory
	query := `
		UPDATE memories
		SET content = $1, embedding = $2::vector, namespace = $3, 
		    tags = $4, metadata = $5, updated_at = $6
		WHERE id = $7
	`

	_, err = s.db.ExecContext(ctx, query,
		memory.Content,
		embeddingStr,
		namespace,
		tagsArray,
		metadataJSON,
		memory.UpdatedAt,
		memory.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update memory: %w", err)
	}

	return nil
}

// Delete removes a memory by its ID
func (s *PostgresStorage) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM memories WHERE id = $1`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("memory not found: %s", id)
	}

	return nil
}

// DeleteByNamespace removes all memories in a namespace
func (s *PostgresStorage) DeleteByNamespace(namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `DELETE FROM memories WHERE namespace = $1`

	_, err := s.db.ExecContext(ctx, query, namespace)
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	return nil
}

// List retrieves memories with optional filtering
// Namespace can be empty string for all namespaces, tag can be empty for no tag filter
// Limit of 0 means no limit
func (s *PostgresStorage) List(namespace string, tag string, limit int) ([]*models.Memory, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `SELECT id, content, namespace, tags, metadata, created_at, updated_at FROM memories WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	// Add namespace filter
	if namespace != "" {
		query += fmt.Sprintf(" AND namespace = $%d", argIdx)
		args = append(args, namespace)
		argIdx++
	}

	// Add tag filter
	if tag != "" {
		query += fmt.Sprintf(" AND $%d = ANY(tags)", argIdx)
		args = append(args, tag)
		argIdx++
	}

	// Add ordering and pagination
	if limit > 0 {
		query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", argIdx)
		args = append(args, limit)
	} else {
		query += " ORDER BY created_at DESC"
	}

	// Execute query
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}
	defer rows.Close()

	// Parse results
	var memories []*models.Memory
	for rows.Next() {
		var memory models.Memory
		var tagsArray []string
		var metadataJSON []byte
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&memory.ID,
			&memory.Content,
			&memory.Namespace,
			&tagsArray,
			&metadataJSON,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		memory.Tags = tagsArray
		memory.CreatedAt = createdAt
		memory.UpdatedAt = updatedAt

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &memory.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		memories = append(memories, &memory)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return memories, nil
}

// ListNamespaces returns all namespaces
func (s *PostgresStorage) ListNamespaces() ([]NamespaceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT namespace, MIN(created_at) as created_at
		FROM memories
		GROUP BY namespace
		ORDER BY namespace
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}
	defer rows.Close()

	var namespaces []NamespaceInfo
	for rows.Next() {
		var ns NamespaceInfo
		var createdAt time.Time

		err := rows.Scan(&ns.Name, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		ns.CreatedAt = createdAt
		namespaces = append(namespaces, ns)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return namespaces, nil
}

// Count returns the number of memories in a namespace
func (s *PostgresStorage) Count(namespace string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT COUNT(*) FROM memories WHERE namespace = $1`

	var count int
	err := s.db.QueryRowContext(ctx, query, namespace).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count memories: %w", err)
	}

	return count, nil
}

// CreateNamespace creates a new namespace
func (s *PostgresStorage) CreateNamespace(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("namespace name cannot be empty")
	}

	// In PostgreSQL, namespaces are implicit (they exist when memories are stored)
	// This is a no-op, but we validate the namespace name
	return nil
}

// DeleteNamespace deletes a namespace and all its memories
func (s *PostgresStorage) DeleteNamespace(namespace string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `DELETE FROM memories WHERE namespace = $1`

	_, err := s.db.ExecContext(ctx, query, namespace)
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

// Helper functions

// embeddingToString converts a float32 slice to pgvector string format
func (s *PostgresStorage) embeddingToString(embedding []float32) string {
	if len(embedding) == 0 {
		return "[]"
	}

	result := "["
	for i, v := range embedding {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%f", v)
	}
	result += "]"
	return result
}

// stringSliceToCSV converts a string slice to comma-separated values
func stringSliceToCSV(slice []string) string {
	if len(slice) == 0 {
		return ""
	}

	result := ""
	for i, s := range slice {
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}
