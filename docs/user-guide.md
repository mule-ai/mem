# mem User Guide

A comprehensive guide to using the `mem` semantic memory CLI tool.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Core Commands](#core-commands)
- [Advanced Features](#advanced-features)
- [Workflows](#workflows)
- [Tips & Best Practices](#tips--best-practices)

## Installation

### From Source

```bash
git clone https://github.com/jbutlerdev/mem.git
cd mem
go build -o mem ./cmd/mem
sudo mv mem /usr/local/bin/  # Optional: add to PATH
```

### Using Go Install

```bash
go install github.com/jbutlerdev/mem@latest
```

### Verify Installation

```bash
mem --version
mem --help
```

## Quick Start

### 1. Initial Setup

Create your configuration file:

```bash
mkdir -p ~/.mem
cat > ~/.mem/config.yaml << 'EOF'
memory:
  path: ~/.mem/data
  default_namespace: default
  backend: chromem

embeddings:
  base_url: http://localhost:11434
  model: nomic-embed-text
  dimensions: 768

reranking:
  enabled: false  # Disable if no reranker available

query:
  default_limit: 5
  min_similarity: 0.6
EOF
```

### 2. Store Your First Memory

```bash
mem store "I prefer working in dark mode and use vim keybindings"
```

Output:
```
✓ Memory stored successfully
  ID: 550e8400-e29b-41d4-a716-446655440000
  Namespace: default
  Tags: (none)
```

### 3. Query Your Memories

```bash
mem query "editor preferences"
```

Output:
```
Top 2 results for "editor preferences":

1. [0.89] I prefer working in dark mode and use vim keybindings
   ID: 550e8400-e29b-41d4-a716-446655440000
   Tags: (none)
   Created: 2026-02-07 14:30:00
```

### 4. List All Memories

```bash
mem list
```

## Configuration

### Configuration File Locations

`mem` uses a cascading configuration system:

1. **Global Config**: `~/.mem/config.yaml` (always loaded)
2. **Local Config**: `./.memconfig` (optional, overrides global)
3. **CLI Flags**: (highest priority, override everything)

### Basic Configuration

```yaml
# ~/.mem/config.yaml

memory:
  # Where to store the vector database
  path: ~/.mem/data
  
  # Default namespace for operations
  default_namespace: default
  
  # Storage backend: chromem or postgres
  backend: chromem

embeddings:
  # API endpoint for embeddings
  base_url: http://localhost:11434
  
  # Model to use
  model: nomic-embed-text
  
  # Optional API key
  api_key: ""
  
  # Embedding dimensions (model-specific)
  dimensions: 768
  
  # Batch size for efficiency
  batch_size: 10

reranking:
  # Enable re-ranking for better results
  enabled: true
  
  # Reranker model
  model: qwen3-reranker-8b
  
  # Top K results to re-rank
  top_k: 10
  
  # Minimum relevance threshold
  threshold: 0.5

query:
  # Default number of results
  default_limit: 5
  
  # Minimum similarity (0-1)
  min_similarity: 0.6
  
  # Enable re-ranking by default
  rerank_enabled: true

cli:
  # Output format: text, json, yaml
  output_format: text
  
  # Verbose logging
  verbose: false
  
  # Color output: true, false, auto
  color: auto
```

### Project-Specific Configuration

Create `.memconfig` in your project directory:

```yaml
# ./.memconfig (project-specific)

memory:
  default_namespace: myproject

embeddings:
  base_url: http://10.10.199.29:8080
  model: text-embedding-qwen3-embedding-8b
```

### Environment Variables

```bash
export MEM_BASE_URL="http://localhost:11434"
export MEM_API_KEY="sk-..."
export MEM_NAMESPACE="default"
export MEM_BACKEND="chromem"
```

## Core Commands

### `mem store` - Store a Memory

Store text content as a semantic memory.

```bash
# Basic usage
mem store "Your memory content here"

# With namespace
mem store -n work "Meeting notes: Q1 planning complete"

# With tags
mem store -t important -t review "Deploy to production on Friday"

# With metadata (JSON)
mem store -m '{"source": "slack", "author": "john"}' "Important announcement"

# From stdin
echo "Multi-line memory content" | mem store -
cat file.txt | mem store -

# With everything
mem store -n project -t frontend -m '{"priority": "high"}' "Fix navigation bug"
```

**Flags:**
- `-n, --namespace` - Namespace for the memory
- `-t, --tag` - Add tags (can be used multiple times)
- `-m, --metadata` - JSON metadata string
- `-` - Read content from stdin

### `mem query` - Search Memories

Search memories using semantic similarity.

```bash
# Basic query
mem query "database configuration"

# With namespace
mem query -n work "meeting notes"

# Limit results
mem query -l 10 "project updates"

# Higher initial retrieval for re-ranking
mem query --top 30 "best practices"

# Without re-ranking (faster)
mem query --no-rerank "quick search"

# With similarity threshold
mem query --threshold 0.8 "specific topic"

# Different output formats
mem query -o json "api endpoints" | jq .
mem query -o yaml "configuration"
```

**Flags:**
- `-n, --namespace` - Namespace to search
- `-l, --limit` - Number of results (default: 5)
- `--top` - Initial retrieval count (default: 10)
- `--threshold` - Minimum similarity score (0-1)
- `--no-rerank` - Disable re-ranking
- `-o, --output` - Output format: text, json, yaml

### `mem list` - List Memories

List all memories with optional filtering.

```bash
# List all memories
mem list

# List in specific namespace
mem list -n work

# Filter by tag
mem list -t important

# Limit results
mem list -l 20

# Different output formats
mem list -o json | jq '.[].memory.content'
mem list -o yaml
```

**Flags:**
- `-n, --namespace` - Namespace to list
- `-t, --tag` - Filter by tag
- `-l, --limit` - Maximum number to show
- `-o, --output` - Output format: text, json, yaml

### `mem delete` - Delete Memories

Delete memories by ID or query.

```bash
# Delete by ID
mem delete abc123-def456-ghi789

# Delete with confirmation
mem delete --query "old meeting notes"

# Force delete without confirmation
mem delete abc123-def456 --force

# Delete in namespace
mem delete -n work abc123-def456
```

**Flags:**
- `-n, --namespace` - Namespace for the memory
- `-q, --query` - Delete by query match (interactive)
- `-f, --force` - Skip confirmation

### `mem update` - Update Memories

Update existing memory content, tags, or metadata.

```bash
# Update content
mem update abc123 "Updated content here"

# Update tags (replaces all tags)
mem update abc123 -t newtag -t another

# Update metadata
mem update abc123 -m '{"status": "resolved"}'

# Update everything
mem update -n work abc123 -t fixed -m '{"verified": true}' "Fixed the bug"
```

**Flags:**
- `-n, --namespace` - Namespace for the memory
- `-t, --tag` - Update tags (replaces existing)
- `-m, --metadata` - Update metadata (replaces existing)

## Advanced Features

### Namespaces

Namespaces allow you to organize memories into separate contexts:

```bash
# Store in different namespaces
mem store -n work "Project deadline is March 15th"
mem store -n personal "Buy groceries on Saturday"
mem store -n ideas "Build a CLI tool for semantic memory"

# Query specific namespace
mem query -n work "deadlines"

# List specific namespace
mem list -n ideas
```

**Setting Default Namespace:**

```yaml
# ~/.mem/config.yaml
memory:
  default_namespace: work
```

```yaml
# ./.memconfig (project-specific)
memory:
  default_namespace: myproject
```

### Tagging

Tags provide flexible organization:

```bash
# Store with tags
mem store -t important -t review "Deploy on Friday"
mem store -t bug "Login page has alignment issue"
mem store -t feature -t frontend "Add dark mode toggle"

# Query by tag (note: use list for tag filtering)
mem list -t important
mem list -t bug

# Multiple tags (AND logic - shows memories with ALL tags)
mem list -t feature -t frontend
```

### Metadata

Attach arbitrary JSON metadata to memories:

```bash
# Store with metadata
mem store -m '{"priority": "high", "assignee": "john"}' "Fix authentication bug"
mem store -m '{"source": "slack", "channel": "#general"}' "Team announcement"
mem store -m '{"deadline": "2026-03-15", "estimated_hours": 8}' "Q1 report"
```

### Import/Export

Backup and restore your memories:

```bash
# Export all memories
mem export backup.json

# Export specific namespace
mem export -n work work-backup.json

# Export in different formats
mem export --format yaml backup.yaml

# Import memories
mem import backup.json

# Import with namespace override
mem import -n archive old-memories.json
```

### Namespace Management

```bash
# List all namespaces
mem namespace list

# Create a new namespace
mem namespace create project-alpha

# Delete a namespace and all its memories
mem namespace delete project-alpha

# Delete with confirmation
mem namespace delete --force old-project
```

### Configuration Management

```bash
# Show current configuration
mem config show

# Get a specific value
mem config get memory.default_namespace
mem config get embeddings.model

# Set a configuration value
mem config set memory.default_namespace work
mem config set cli.output_format json

# Edit configuration in your default editor
mem config edit
```

## Workflows

### Workflow 1: Meeting Notes

```bash
# Store meeting notes with context
mem store -n meetings -t q1 -m '{"date": "2026-02-07", "attendees": ["alice", "bob"]' } \
  "Q1 Planning: Focus on MVP launch, target end of March. Key features: auth, dashboard, reports."

# Query meeting decisions
mem query -n meetings "what was decided about launch timeline"

# List all Q1 meetings
mem list -n meetings -t q1
```

### Workflow 2: Development Notes

```bash
# Store code decisions
mem store -n dev -t architecture -m '{"component": "api", "ticket": "DEV-123"}' \
  "API Gateway uses rate limiting of 1000 req/min per user"

# Store bug notes
mem store -n dev -t bug -m '{"status": "open", "severity": "high"}' \
  "Memory leak in WebSocket handler under high concurrency"

# Query architectural decisions
mem query -n dev "rate limiting"

# Find open bugs
mem list -n dev -t bug
```

### Workflow 3: Personal Knowledge Base

```bash
# Store articles/notes
mem store -t article "JWT tokens should have short expiration (15min) and use refresh tokens"

# Store commands
mem store -t command "ffmpeg -i input.mp4 -vf scale=1280:-1 output.mp4"

# Store recipes
mem store -n recipes -t dessert "Chocolate cake: 2 cups flour, 2 cups sugar, 3/4 cup cocoa, 2 eggs, 1 cup milk, 1/2 cup butter. Bake at 350°F for 30-35 min."

# Query later
mem query "how to resize video"
mem query -n recipes "chocolate"
mem list -t command
```

### Workflow 4: Project Context Management

```bash
# In project directory, create .memconfig
cat > .memconfig << EOF
memory:
  default_namespace: myproject
EOF

# Store project-specific memories
mem store "Database: PostgreSQL 14 with pgvector for embeddings"
mem store "API: RESTful, uses JWT auth, rate limited to 1000/min"
mem store -t config "LM Studio runs on http://10.10.199.29:8080"

# Team members can query project context
mem query "database version"
mem list
```

### Workflow 5: Backup & Sync

```bash
# Regular backups
mem export backup-$(date +%Y%m%d).json

# Sync between machines
scp ~/.mem/data remote:~/.mem/data/
mem import backup-20260207.json

# Namespace-specific backup
mem export -n work work-$(date +%Y%m%d).json
```

## Tips & Best Practices

### 1. Use Namespaces Effectively

- Separate work/personal/projects into different namespaces
- Set project-specific namespace with `.memconfig`
- Use descriptive namespace names: `work`, `personal`, `project-alpha`

### 2. Tag Strategy

- Use consistent tag names: `important`, `todo`, `done`, `bug`, `feature`
- Tag by type: `command`, `note`, `decision`, `config`
- Tag by status: `open`, `in-progress`, `resolved`

### 3. Memory Content Quality

- Be specific and detailed for better semantic matching
- Include context: "Database: PostgreSQL 14" vs "PostgreSQL 14"
- Use complete sentences for better embedding quality

### 4. Query Techniques

- Use natural language queries: "how do I configure the database"
- Try different phrasings if results aren't relevant
- Use `--no-rerank` for faster searches when appropriate
- Increase `--top` when looking for obscure information

### 5. Performance Tuning

```yaml
# For faster stores (lower quality)
embeddings:
  batch_size: 20  # Process more at once

# For faster queries (less accurate)
query:
  default_limit: 3
reranking:
  enabled: false

# For better quality (slower)
embeddings:
  batch_size: 1
query:
  default_limit: 10
reranking:
  top_k: 30
```

### 6. Regular Maintenance

```bash
# Export regular backups
mem export "backup-$(date +%Y%m%d).json"

# Clean up old memories
mem delete --query "outdated meeting notes from 2025"

# Review and organize
mem list -l 100 | less
```

### 7. Working with JSON Output

```bash
# Extract specific fields
mem query -o json "database" | jq -r '.[].memory.content'

# Count memories by namespace
mem list -o json | jq 'group_by(.memory.namespace) | map({namespace: .[0].memory.namespace, count: length})'

# Find memories with specific metadata
mem list -o json | jq '.[] | select(.memory.metadata.priority == "high")'
```

### 8. Integration with Other Tools

```bash
# Pipe from other commands
grep -r "TODO" src/ | mem store -t todo -

# Store git commit messages
git log --oneline -10 | mem store -t git -

# Store command outputs
df -h | mem store -t sysadmin "disk usage on $(hostname)"
```

## Troubleshooting

See [troubleshooting.md](troubleshooting.md) for common issues and solutions.

## Further Reading

- [Configuration Reference](configuration-reference.md)
- [Product Specification](../spec.md)
- [Examples](../examples/)