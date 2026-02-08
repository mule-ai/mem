# Memory CLI Tool - Product Specification

**Version:** 1.0  
**Date:** 2026-02-07  
**Status:** Draft

## Executive Summary

A command-line interface (CLI) tool for persistent, semantic memory storage and retrieval using vector embeddings. The tool enables users to store, query, and manage memories with advanced features including namespacing, re-ranking for improved accuracy, and flexible backend storage options.

## Core Technologies

- **Language:** Go (Golang)
- **Vector Database:** chromem-go (primary)
- **Alternative Backend:** PostgreSQL + pgvector (provisional)
- **Embeddings Provider:** OpenAI-compatible API
- **Development Endpoint:** http://10.10.199.29:8080 (LM Studio)
- **Embedding Model:** `text-embedding-qwen3-embedding-8b`
- **Reranker Model:** `qwen3-reranker-8b`

## Architecture Overview

```
� CLI Tool
├── Configuration Layer
│   ├── Global: ~/.mem/config.yaml
│   └── Local: ./.memconfig (overrides)
├── Storage Layer
│   ├── chromem-go (default)
│   └── PostgreSQL + pgvector (optional)
├── Embedding Layer
│   └── OpenAI-compatible API client
└── Reranking Layer
    └── Query result re-ranking
```

## Configuration

### Global Configuration: `~/.mem/config.yaml`

```yaml
# Default configuration location
memory:
  # Database storage location (for chromem-go)
  path: ~/.mem/data
  
  # Default namespace for memories
  default_namespace: default
  
  # Storage backend: chromem | postgres
  backend: chromem
  
  # PostgreSQL configuration (when backend=postgres)
  postgres:
    host: localhost
    port: 5432
    database: mem
    user: postgres
    password: ""
    sslmode: disable

# Embedding provider configuration
embeddings:
  # API endpoint
  base_url: http://10.10.199.29:8080
  
  # Model to use for embeddings
  model: text-embedding-qwen3-embedding-8b
  
  # API key (optional for local deployments)
  api_key: ""
  
  # Embedding dimensions
  dimensions: 1024
  
  # Batch size for embedding requests
  batch_size: 10

# Re-ranking configuration
reranking:
  enabled: true
  model: qwen3-reranker-8b
  top_k: 10  # Re-rank top N results
  threshold: 0.5  # Minimum relevance score

# Query behavior
query:
  # Default number of results to return
  default_limit: 5
  
  # Minimum similarity threshold (0-1)
  min_similarity: 0.6
  
  # Enable re-ranking by default
  rerank_enabled: true

# CLI behavior
cli:
  # Output format: text | json | yaml
  output_format: text
  
  # Enable verbose logging
  verbose: false
  
  # Color output
  color: true
```

### Local Configuration: `./.memconfig`

Overrides global configuration in the current working directory. This is useful for project-specific memory contexts.

```yaml
memory:
  default_namespace: project-name

embeddings:
  base_url: http://10.10.199.29:8080
  model: text-embedding-qwen3-embedding-8b
```

### Configuration Precedence

1. Command-line flags (highest priority)
2. `.memconfig` in current working directory
3. `~/.mem/config.yaml` (global config)
4. Built-in defaults (lowest priority)

## CLI Interface

### Command Structure

```bash
mem [global_flags] <command> [command_flags] [arguments]
```

### Global Flags

```
--config, -c     Path to config file (default: ~/.mem/config.yaml)
--namespace, -n  Namespace for the operation
--backend        Storage backend to use (chromem | postgres)
--output, -o     Output format: text | json | yaml
--verbose, -v    Enable verbose logging
--help, -h       Show help
--version        Show version information
```

### Commands

#### 1. Store Memory

```bash
mem store [flags] <content>

# Examples:
mem store "User prefers dark mode and vim keybindings"
mem store -n project-alpha "Database uses PostgreSQL 14"
mem store --tag preferences "Important: Always use TypeScript strict mode"
echo "Multi-line memory" | mem store -
cat file.txt | mem store -
```

**Flags:**
```
--namespace, -n   Namespace for the memory (default: from config)
--tag, -t         Tags to attach (can be used multiple times)
--metadata, -m    JSON metadata string
--stdin, -        Read content from stdin
```

#### 2. Query Memory

```bash
mem query [flags] <query>

# Examples:
mem query "What are the user's preferences?"
mem query -n project-alpha "database version"
mem query --top 20 "typescript configuration"
mem query --no-rerank "quick search"
```

**Flags:**
```
--namespace, -n    Namespace to search (default: from config)
--limit, -l        Number of results (default: 5)
--top              Initial retrieval count before re-ranking (default: 10)
--threshold        Minimum similarity score (default: 0.6)
--no-rerank        Disable re-ranking
--output, -o       Output format: text | json | yaml
```

#### 3. List Memories

```bash
mem list [flags]

# Examples:
mem list
mem list -n project-alpha
mem list --tag preferences
mem list --format json
```

**Flags:**
```
--namespace, -n    Namespace to list
--tag, -t          Filter by tag
--limit, -l        Maximum number to show
--format, -o       Output format: text | json | yaml
```

#### 4. Delete Memory

```bash
mem delete [flags] <id>

# Examples:
mem delete abc123
mem delete -n project-alpha abc123
mem delete --query "old preferences"  # Delete by query match
```

**Flags:**
```
--namespace, -n    Namespace for the memory
--query, -q        Delete by query match (interactive confirmation)
--force, -f        Skip confirmation
```

#### 5. Update Memory

```bash
mem update [flags] <id> <new_content>

# Examples:
mem update abc123 "Updated: User now prefers light mode"
```

**Flags:**
```
--namespace, -n   Namespace for the memory
--tag, -t         Update tags (replaces existing)
--metadata, -m    Update metadata (replaces existing)
```

#### 6. Import/Export

```bash
# Export memories
mem export [flags] [output_file]

# Import memories
mem import [flags] <input_file>

# Examples:
mem export backup.json
mem import backup.json
mem export -n project-alpha project-backup.json
```

**Flags:**
```
--namespace, -n    Namespace to export/import
--format           Format: json | yaml (default: json)
```

#### 7. Namespace Management

```bash
# List namespaces
mem namespace list

# Create namespace
mem namespace create <name>

# Delete namespace (and all memories)
mem namespace delete <name>
```

#### 8. Configuration Management

```bash
# Show current configuration
mem config show

# Set configuration value
mem config set <key> <value>

# Get configuration value
mem config get <key>

# Edit configuration
mem config edit
```

## Data Model

### Memory Structure

```go
type Memory struct {
    ID          string                 `json:"id"`
    Content     string                 `json:"content"`
    Embedding   []float32              `json:"-"`  // Not stored in metadata
    Namespace   string                 `json:"namespace"`
    Tags        []string               `json:"tags,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

type SearchResult struct {
    Memory     Memory  `json:"memory"`
    Score      float32 `json:"score"`      // Similarity score
    Reranked   bool    `json:"reranked"`   // Whether result was re-ranked
    Rank       int     `json:"rank"`       // Final rank after re-ranking
}
```

## Features

### 1. Semantic Search with Re-ranking

1. **Initial Retrieval:** Fetch top-K results (default: 20) via vector similarity
2. **Re-ranking:** Use `qwen3-reranker-8b` to re-rank results based on query relevance
3. **Final Output:** Return top-N results (default: 5) after re-ranking

### 2. Namespacing

- Memories are organized into namespaces
- Default namespace can be set in config
- Namespace isolation prevents cross-contamination
- Useful for multi-project or multi-context scenarios

### 3. Tagging

- Attach arbitrary tags to memories
- Filter memories by tags during queries
- Support for multiple tags per memory

### 4. Metadata Support

- Store arbitrary JSON metadata with memories
- Useful for storing context, sources, or additional info

### 5. Flexible Output Formats

- **Text:** Human-readable, formatted output
- **JSON:** Machine-readable, for scripting
- **YAML:** Human-readable structured format

### 6. Batch Operations

- Store multiple memories efficiently
- Batch embedding requests for performance

## Technical Implementation Details

### Storage Backends

#### chromem-go (Primary)

- Embedded vector database
- No external dependencies
- Fast for local development
- Stores data in local filesystem

#### PostgreSQL + pgvector (Provisional)

- Scalable for large datasets
- Supports concurrent access
- Requires external database
- Useful for production deployments

### Embedding Pipeline

1. **Input Processing:** Clean and prepare text for embedding
2. **Batching:** Group texts into batches for efficiency
3. **API Request:** Call embedding API with batch
4. **Response Parsing:** Extract embedding vectors
5. **Storage:** Store vectors in chosen backend

### Re-ranking Pipeline

1. **Vector Search:** Retrieve top-K candidates via similarity
2. **Rerank API:** Call reranker with query and candidates
3. **Score Adjustment:** Update scores based on reranker output
4. **Final Ranking:** Sort by reranked scores
5. **Output:** Return top-N results

### Error Handling

- Graceful degradation when reranker unavailable
- Retry logic for API failures
- Clear error messages for configuration issues
- Validation of user input

## Development Notes

### Local Development Setup

- **LM Studio Endpoint:** http://10.10.199.29:8080
- **Embedding Model:** text-embedding-qwen3-embedding-8b
- **Reranker Model:** qwen3-reranker-8b

### Environment Variables

For development/testing:

```bash
export MEM_BASE_URL=http://10.10.199.29:8080
export MEM_EMBEDDING_MODEL=text-embedding-qwen3-embedding-8b
export MEM_RERANKER_MODEL=qwen3-reranker-8b
```

### Project Structure

```
mem/
├── cmd/
│   └── mem/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── loader.go
│   ├── storage/
│   │   ├── storage.go
│   │   ├── chromem.go
│   │   └── postgres.go
│   ├── embeddings/
│   │   ├── client.go
│   │   └── openai.go
│   ├── reranker/
│   │   └── reranker.go
│   ├── cli/
│   │   ├── root.go
│   │   ├── store.go
│   │   ├── query.go
│   │   └── delete.go
│   └── models/
│       └── memory.go
├── pkg/
│   └── api/
│       └── client.go
├── docs/
│   └── product-spec.md
├── go.mod
├── go.sum
└── README.md
```

## Future Enhancements

### Phase 2 Features

- [ ] Memory expiration/TTL
- [ ] Memory deduplication
- [ ] Fuzzy matching alongside semantic search
- [ ] Memory relationships/graph
- [ ] CLI autocomplete
- [ ] Interactive mode (REPL)
- [ ] Memory statistics and analytics
- [ ] Backup and restore utilities
- [ ] Encryption at rest
- [ ] Access control and permissions

### Phase 3 Features

- [ ] Multi-modal support (images, audio)
- [ ] Memory summarization
- [ ] Automatic memory organization
- [ ] Web UI for memory management
- [ ] API server mode
- [ ] Plugins/extensions system
- [ ] Integration with other tools (Obsidian, Notion, etc.)

## Testing Strategy

### Unit Tests

- Configuration loading and merging
- Embedding API client
- Reranker client
- Storage backend operations

### Integration Tests

- End-to-end memory operations
- Configuration overrides
- API interactions with LM Studio

### Performance Tests

- Large dataset handling
- Concurrent operations
- Embedding batch optimization

## Documentation Requirements

- User guide with examples
- Configuration reference
- API documentation (if server mode added)
- Troubleshooting guide
- Contributing guidelines

## Security Considerations

- API key storage in config (optional for local LM Studio)
- File permissions on memory database
- Input sanitization
- SQL injection prevention (PostgreSQL backend)
- Sensitive data handling in memories

## Performance Targets

- **Store Latency:** < 500ms per memory (including embedding)
- **Query Latency:** < 1s for typical queries
- **Reranking:** < 2s for top-20 results
- **Storage:** Support for 100k+ memories
- **Concurrent Users:** 10+ (with PostgreSQL backend)

## Success Criteria

- [ ] Functional CLI with all core commands
- [ ] Reliable embedding and storage with chromem-go
- [ ] Working re-ranking with qwen3-reranker-8b
- [ ] Flexible configuration system
- [ ] Clear error messages and help text
- [ ] Test coverage > 70%
- [ ] Performance targets met
- [ ] User documentation complete

## Open Questions

1. Should memories support versioning?
2. What's the max memory size we should support?
3. Should we support memory compression for storage efficiency?
4. Do we need memory pinning (prevent deletion)?
5. Should we implement a memory merge feature for duplicates?