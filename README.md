# mem - Semantic Memory CLI Tool

A command-line interface for persistent, semantic memory storage and retrieval using vector embeddings.

## Overview

`mem` enables you to store, query, and manage memories with advanced semantic search capabilities. Built in Go with support for chromem-go vector database (with provisional PostgreSQL/pgvector support), OpenAI-compatible embeddings, and re-ranking for improved accuracy.

## Features

- **Semantic Search**: Find memories by meaning, not just keywords
- **Re-ranking**: Enhanced accuracy using `qwen3-reranker-8b`
- **Namespacing**: Organize memories into separate contexts
- **Tagging**: Add custom tags for filtering and organization
- **Flexible Storage**: chromem-go (default) or PostgreSQL + pgvector
- **Multiple Output Formats**: text, JSON, or YAML
- **Configurable**: Global and project-specific configuration

## Quick Start

### Installation

```bash
# From source
git clone https://github.com/jbutlerdev/mem.git
cd mem
go build -o mem ./cmd/mem
sudo mv mem /usr/local/bin/  # Optional: add to PATH

# Or using go install
go install github.com/jbutlerdev/mem@latest
```

### Configuration

Create `~/.mem/config.yaml`:

```yaml
memory:
  path: ~/.mem/data
  default_namespace: default
  backend: chromem

embeddings:
  base_url: http://localhost:11434
  model: nomic-embed-text
  dimensions: 768

reranking:
  enabled: false  # Enable if you have a reranker

query:
  default_limit: 5
  min_similarity: 0.6
```

For LM Studio development setup:

```yaml
embeddings:
  base_url: http://10.10.199.29:8080
  model: text-embedding-qwen3-embedding-8b
  dimensions: 1024

reranking:
  enabled: true
  model: qwen3-reranker-8b
```

### Basic Usage

```bash
# Store a memory
mem store "User prefers dark mode and vim keybindings"

# Query memories
mem query "What are the user's preferences?"

# List all memories
mem list

# Use a specific namespace
mem store -n project-alpha "Database uses PostgreSQL 14"
mem query -n project-alpha "database version"

# Delete a memory
mem delete <memory-id>
```

## Usage Examples

### Personal Knowledge Management

```bash
# Store things you learn
mem store -t learning "Go's context package is used for cancellation and deadlines"
mem store -t tip "In Vim, use ciw to change inner word"

# Query your knowledge later
mem query "context cancellation"
mem list -t tip
```

### Development Workflow

```bash
# In your project directory, create .memconfig
echo 'memory: {default_namespace: myproject}' > .memconfig

# Store project-specific memories
mem store "Database: PostgreSQL 14 with pgvector"
mem store -t bug "Race condition in payment processing"

# Query project context
mem query "database configuration"
mem list -t bug
```

### Meeting Notes

```bash
# Store meeting notes with metadata
mem store -n meetings -t standup -m '{"date": "2026-02-07"}' \
  "Discussed API migration. Target date: March 15th."

# Query meetings
mem query -n meetings "api migration deadline"
mem list -n meetings -t standup
```

### Command Library

```bash
# Store useful commands
mem store -n commands -t docker "Remove stopped containers: docker container prune"
mem store -n commands -t git "Undo last commit: git reset --soft HEAD~1"

# Find commands later
mem query -n commands "docker container remove"
```

### Research & Notes

```bash
# Store research findings
mem store -n research -t ml "Transformers use self-attention mechanisms for sequence processing"
mem store -n research -t database "Vector databases enable semantic search via embeddings"

# Cross-reference research topics
mem query -n research "attention mechanisms"
mem query "machine learning and databases"
```

### Quick Reference

```bash
# Store documentation snippets
mem store -n docs -t api "POST /api/v1/memories - Create new memory with embedding"
mem store -n docs -t config "Set base_url in config.yaml to change embedding endpoint"

# Find documentation quickly
mem query -n docs -t api "create memory"
mem query "configuration settings"
```

### Workflow Automation

```bash
# Store decision logs
mem store -n decisions -m '{"date": "2026-02-07", "team": "backend"}' \
  "Chose chromem-go over PostgreSQL for initial MVP due to simpler deployment"

# Track technical debt
mem store -t tech-debt "Refactor: Embedding client needs retry logic for production"

# Query decisions and debt
mem query -n decisions "deployment strategy"
mem list -t tech-debt
```

## Documentation

- **[User Guide](docs/user-guide.md)** - Comprehensive usage documentation
- **[Configuration Reference](docs/configuration-reference.md)** - All configuration options
- **[Example Workflows](docs/examples.md)** - Practical examples and workflows
- **[Troubleshooting Guide](docs/troubleshooting.md)** - Common issues and solutions
- **[Product Specification](spec.md)** - Complete technical specification

## Commands

| Command | Description |
|---------|-------------|
| `mem store` | Store a new memory |
| `mem query` | Search memories semantically |
| `mem list` | List all memories with filtering |
| `mem delete` | Delete memories by ID or query |
| `mem update` | Update existing memory content |
| `mem export` | Export memories to file |
| `mem import` | Import memories from file |
| `mem namespace` | Manage namespaces |
| `mem config` | Manage configuration |

### Global Flags

- `-n, --namespace` - Namespace for the operation
- `-o, --output` - Output format: text, json, yaml
- `-v, --verbose` - Enable verbose logging
- `-c, --config` - Path to config file

## Project-Specific Configuration

Create `.memconfig` in your project directory:

```yaml
memory:
  default_namespace: myproject

embeddings:
  base_url: http://10.10.199.29:8080
  model: text-embedding-qwen3-embedding-8b
```

This overrides global config for the current directory.

## Output Formats

```bash
# Text (default, human-readable)
mem query "database"

# JSON (for scripting)
mem query -o json "database" | jq '.[].memory.content'

# YAML (structured)
mem query -o yaml "database"
```

## Advanced Usage

### Import/Export Workflows

```bash
# Export all memories to JSON
mem export backup.json

# Export specific namespace
mem export -n work work-memories.json

# Import memories from backup
mem import backup.json

# Export and process with jq
mem export -o json | jq '.[] | select(.tags[]? == "important")' > important.json
```

### Batch Operations

```bash
# Store multiple memories from a file
while IFS= read -r line; do
  mem store "$line"
done < notes.txt

# Store with heredoc
mem store << 'EOF'
Multi-line memory content.
Can include multiple paragraphs.
Useful for detailed notes.
EOF

# Pipe content to store
echo "Quick memory" | mem store -
cat document.txt | mem store -
```

### Query Strategies

```bash
# Narrow search with higher threshold
mem query --threshold 0.8 "exact match needed"

# Broad search with more results
mem query --limit 20 --threshold 0.4 "broad search"

# Compare with and without reranking
mem query "search query"
mem query --no-rerank "search query"

# Search in specific namespace only
mem query -n project "architecture decisions"
```

### Memory Management

```bash
# Update memory content
MEM_ID=$(mem query -o json "old content" | jq -r '.[0].memory.id')
mem update $MEM_ID "updated content with new information"

# Delete by query (with confirmation)
mem delete --query "outdated information"
# Or force delete without confirmation
mem delete --query "temp notes" --force

# List memories with filters
mem list -n project -t bug --limit 20
mem list --tag important --format json
```

### Namespace Management

```bash
# List all namespaces
mem namespace list

# Create new namespace (automatically created when storing)
mem store -n new-namespace "First memory in new namespace"

# Delete namespace and all memories
mem namespace delete old-project
```

### Configuration Management

```bash
# Show current configuration
mem config show

# Get specific value
mem config get memory.default_namespace

# Set configuration value
mem config set memory.default_namespace work

# Edit configuration in editor
mem config edit
```

### Scripting Integration

```bash
# Use mem in scripts with JSON output
#!/bin/bash
QUERY="database configuration"
RESULTS=$(mem query -o json "$QUERY")
COUNT=$(echo "$RESULTS" | jq 'length')
echo "Found $COUNT memories for: $QUERY"

# Loop through results
mem query -o json "api endpoints" | jq -c '.[]' | while read -r result; do
  CONTENT=$(echo "$result" | jq -r '.memory.content')
  echo "- $CONTENT"
done
```

## Development

### Local Development Setup

The project can use LM Studio for embeddings and re-ranking during development:

**LM Studio Configuration:**
- Endpoint: `http://10.10.199.29:8080`
- Embedding Model: `text-embedding-qwen3-embedding-8b`
- Reranker Model: `qwen3-reranker-8b`

### Building

```bash
# Clone the repository
git clone https://github.com/jbutlerdev/mem.git
cd mem

# Build
go build -o mem ./cmd/mem

# Run
./mem --help
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./internal/storage/

# Run benchmarks
go test -bench=. ./internal/embeddings/
```

### Performance Benchmarks

Detailed performance benchmarks are available in [docs/benchmarks/performance-report.md](docs/benchmarks/performance-report.md).

**Summary:**
- **Store:** ~0.08ms per memory (chromem-go storage only)
- **Query:** ~0.67-0.90ms per query
- **Get by ID:** ~0.002ms
- **Re-ranking:** ~0.35ms API overhead + model inference time

All operations are well under performance targets. Actual end-to-end latency includes embedding generation time from the external API.

### Project Structure

```
mem/
├── cmd/mem/           # CLI application
├── internal/
│   ├── config/        # Configuration management
│   ├── storage/       # Storage backends (chromem, postgres)
│   ├── embeddings/    # Embedding API client
│   ├── reranker/      # Re-ranking client
│   ├── cli/           # CLI commands
│   └── models/        # Data models
├── pkg/api/           # Shared API client
├── docs/              # Documentation
├── examples/          # Configuration examples
└── tests/             # Integration tests
```

## Contributing

Contributions are welcome! Please see:

1. [Product Specification](spec.md) for roadmap and architecture
2. [Implementation Plan](plan.md) for current progress
3. [Issue Tracker](https://github.com/jbutlerdev/mem/issues) for open tasks

## License

MIT License - see LICENSE for details.

## Roadmap

### Current (v1.0)
- ✅ Core CLI commands
- ✅ chromem-go storage backend
- ✅ OpenAI-compatible embeddings
- ✅ Semantic search with re-ranking
- ✅ Namespacing and tagging

### Future Enhancements
- [ ] Memory expiration/TTL
- [ ] Memory deduplication
- [ ] CLI autocomplete
- [ ] Interactive mode (REPL)
- [ ] Memory statistics and analytics
- [ ] Encryption at rest
- [ ] Web UI for memory management
- [ ] Multi-modal support (images, audio)
