# mem Configuration Reference

Complete reference for all `mem` configuration options.

## Configuration Overview

`mem` uses a hierarchical configuration system:

1. **Built-in Defaults** - Hardcoded default values
2. **Global Config** - `~/.mem/config.yaml`
3. **Local Config** - `./.memconfig` (optional)
4. **Environment Variables** - `MEM_*` prefixed
5. **CLI Flags** - Highest priority

## Configuration File Format

Configuration files use YAML format. The top-level structure is:

```yaml
memory:          # Storage configuration
embeddings:      # Embedding API settings
reranking:       # Re-ranking settings
query:           # Query behavior defaults
cli:             # CLI output and behavior
postgres:        # PostgreSQL backend settings
```

## Complete Configuration Options

### Memory Section

Controls where and how memories are stored.

```yaml
memory:
  # Directory for chromem-go database storage
  # Type: string
  # Default: ~/.mem/data
  # Environment: MEM_PATH
  path: ~/.mem/data
  
  # Default namespace for memory operations
  # Type: string
  # Default: default
  # Environment: MEM_NAMESPACE
  default_namespace: default
  
  # Storage backend to use
  # Type: string
  # Values: chromem, postgres
  # Default: chromem
  # Environment: MEM_BACKEND
  backend: chromem
```

#### Example Configurations

**Local development:**
```yaml
memory:
  path: /tmp/mem-data
  default_namespace: dev
```

**Production with PostgreSQL:**
```yaml
memory:
  backend: postgres
  default_namespace: prod
```

**Project-specific override (`.memconfig`):**
```yaml
memory:
  default_namespace: myproject
```

### Embeddings Section

Configures the embedding API client.

```yaml
embeddings:
  # Base URL for embedding API
  # Type: string (URL)
  # Default: http://localhost:11434
  # Environment: MEM_BASE_URL
  base_url: http://10.10.199.29:8080
  
  # Model name to use for embeddings
  # Type: string
  # Default: nomic-embed-text
  # Environment: MEM_MODEL
  model: text-embedding-qwen3-embedding-8b
  
  # API key (optional for local deployments)
  # Type: string
  # Default: ""
  # Environment: MEM_API_KEY
  api_key: ""
  
  # Embedding vector dimensions
  # Type: integer
  # Default: 768 (varies by model)
  # Note: Must match your model's output dimensions
  dimensions: 1024
  
  # Number of texts to embed in each batch
  # Type: integer
  # Default: 10
  # Range: 1-100
  # Note: Higher values = faster but more memory usage
  batch_size: 10
```

#### Model-Specific Configurations

**Nomic Embed Text:**
```yaml
embeddings:
  base_url: http://localhost:11434
  model: nomic-embed-text
  dimensions: 768
  batch_size: 10
```

**Qwen3 Embedding (LM Studio):**
```yaml
embeddings:
  base_url: http://10.10.199.29:8080
  model: text-embedding-qwen3-embedding-8b
  dimensions: 1024
  batch_size: 10
```

**OpenAI API:**
```yaml
embeddings:
  base_url: https://api.openai.com/v1
  model: text-embedding-3-small
  dimensions: 1536
  api_key: sk-...
  batch_size: 100
```

### Re-ranking Section

Configures the re-ranking feature for improved search results.

```yaml
reranking:
  # Enable or disable re-ranking
  # Type: boolean
  # Default: true
  # Note: Disable if reranker unavailable
  enabled: true
  
  # Reranker model to use
  # Type: string
  # Default: qwen3-reranker-8b
  model: qwen3-reranker-8b
  
  # Number of top results to re-rank
  # Type: integer
  # Default: 10
  # Range: 1-100
  # Note: Retrieved top_k, re-ranked, then top N returned
  top_k: 10
  
  # Minimum relevance threshold for re-ranked results
  # Type: float
  # Default: 0.5
  # Range: 0.0-1.0
  # Note: Results below this score are filtered out
  threshold: 0.5
```

#### Example Configurations

**Re-ranking enabled (recommended):**
```yaml
reranking:
  enabled: true
  model: qwen3-reranker-8b
  top_k: 20
  threshold: 0.5
```

**Re-ranking disabled (faster, less accurate):**
```yaml
reranking:
  enabled: false
```

**High quality re-ranking:**
```yaml
reranking:
  enabled: true
  top_k: 50
  threshold: 0.7
```

### Query Section

Default behavior for query operations.

```yaml
query:
  # Default number of results to return
  # Type: integer
  # Default: 5
  # Range: 1-100
  # Can be overridden with: -l/--limit flag
  default_limit: 5
  
  # Minimum similarity threshold (0-1)
  # Type: float
  # Default: 0.6
  # Range: 0.0-1.0
  # Note: Results below this score are not returned
  min_similarity: 0.6
  
  # Enable re-ranking by default for queries
  # Type: boolean
  # Default: true
  # Can be overridden with: --no-rerank flag
  rerank_enabled: true
```

#### Example Configurations

**Conservative (high quality, fewer results):**
```yaml
query:
  default_limit: 3
  min_similarity: 0.8
  rerank_enabled: true
```

**Balanced (default):**
```yaml
query:
  default_limit: 5
  min_similarity: 0.6
  rerank_enabled: true
```

**Exploratory (more results, lower threshold):**
```yaml
query:
  default_limit: 10
  min_similarity: 0.5
  rerank_enabled: true
```

**Fast (no re-ranking):**
```yaml
query:
  default_limit: 5
  min_similarity: 0.7
  rerank_enabled: false
```

### CLI Section

Controls CLI output and behavior.

```yaml
cli:
  # Default output format
  # Type: string
  # Values: text, json, yaml
  # Default: text
  # Can be overridden with: -o/--output flag
  output_format: text
  
  # Enable verbose logging
  # Type: boolean
  # Default: false
  # Can be overridden with: -v/--verbose flag
  verbose: false
  
  # Color output
  # Type: string or boolean
  # Values: true, false, auto
  # Default: auto
  # Note: auto enables color for TTY only
  color: auto
```

#### Example Configurations

**Human-readable (default):**
```yaml
cli:
  output_format: text
  color: auto
```

**Script-friendly:**
```yaml
cli:
  output_format: json
  color: false
  verbose: false
```

**Debug mode:**
```yaml
cli:
  output_format: text
  verbose: true
  color: true
```

### PostgreSQL Section

Configuration for PostgreSQL backend (only used when `memory.backend: postgres`).

```yaml
memory:
  backend: postgres

postgres:
  # Database host
  # Type: string
  # Default: localhost
  host: localhost
  
  # Database port
  # Type: integer
  # Default: 5432
  port: 5432
  
  # Database name
  # Type: string
  # Default: mem
  database: mem
  
  # Database user
  # Type: string
  # Default: postgres
  user: postgres
  
  # Database password
  # Type: string
  # Default: ""
  password: ""
  
  # SSL mode
  # Type: string
  # Values: disable, require, verify-ca, verify-full
  # Default: disable
  sslmode: disable
  
  # Connection pool settings
  # Type: integer
  # Default: 10
  max_connections: 10
```

#### Example Configurations

**Local PostgreSQL:**
```yaml
memory:
  backend: postgres

postgres:
  host: localhost
  port: 5432
  database: mem
  user: postgres
  password: "your-password"
  sslmode: disable
```

**Cloud PostgreSQL (e.g., AWS RDS):**
```yaml
memory:
  backend: postgres

postgres:
  host: db.example.com
  port: 5432
  database: mem_prod
  user: mem_user
  password: "secure-password"
  sslmode: require
  max_connections: 20
```

## Environment Variables

All configuration can be overridden via environment variables:

### Memory Variables

```bash
export MEM_PATH="/path/to/data"
export MEM_NAMESPACE="default"
export MEM_BACKEND="chromem"
```

### Embedding Variables

```bash
export MEM_BASE_URL="http://localhost:11434"
export MEM_MODEL="nomic-embed-text"
export MEM_API_KEY="sk-..."
```

### Example: Development Setup

```bash
# ~/.bashrc or ~/.zshrc

# Local development with LM Studio
export MEM_BASE_URL="http://10.10.199.29:8080"
export MEM_MODEL="text-embedding-qwen3-embedding-8b"

# Production namespace
export MEM_NAMESPACE="prod"

# Enable verbose output
export MEM_VERBOSE="true"
```

## CLI Flag Reference

### Global Flags

Available on all commands:

| Flag | Short | Type | Description | Config Override |
|------|-------|------|-------------|-----------------|
| `--config` | `-c` | string | Path to config file | - |
| `--namespace` | `-n` | string | Namespace for operation | `memory.default_namespace` |
| `--backend` | - | string | Storage backend | `memory.backend` |
| `--output` | `-o` | string | Output format | `cli.output_format` |
| `--verbose` | `-v` | boolean | Enable verbose | `cli.verbose` |
| `--help` | `-h` | - | Show help | - |
| `--version` | - | - | Show version | - |

### Command-Specific Flags

#### `mem store`

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--tag` | `-t` | string[] | Add tags (repeatable) |
| `--metadata` | `-m` | string | JSON metadata |
| `--stdin` | `-` | - | Read from stdin |

#### `mem query`

| Flag | Short | Type | Description | Config Override |
|------|-------|------|-------------|-----------------|
| `--limit` | `-l` | int | Number of results | `query.default_limit` |
| `--top` | - | int | Initial retrieval count | `reranking.top_k` |
| `--threshold` | - | float | Min similarity | `query.min_similarity` |
| `--no-rerank` | - | boolean | Disable re-ranking | `query.rerank_enabled` |

#### `mem list`

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--tag` | `-t` | string | Filter by tag |
| `--limit` | `-l` | int | Max results |

#### `mem delete`

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--query` | `-q` | string | Delete by query |
| `--force` | `-f` | boolean | Skip confirmation |

#### `mem update`

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--tag` | `-t` | string[] | Replace tags |
| `--metadata` | `-m` | string | Replace metadata |

#### `mem export`

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--format` | - | string | Output format |

#### `mem import`

| Flag | Short | Type | Description |
|------|-------|------|-------------|
| `--namespace` | `-n` | string | Override namespace |

## Configuration Examples

### Minimal Configuration

```yaml
# ~/.mem/config.yaml

memory:
  path: ~/.mem/data

embeddings:
  base_url: http://localhost:11434
```

### Development Configuration

```yaml
# ~/.mem/config.yaml

memory:
  path: /tmp/mem-data
  default_namespace: dev

embeddings:
  base_url: http://10.10.199.29:8080
  model: text-embedding-qwen3-embedding-8b
  dimensions: 1024

reranking:
  enabled: true
  model: qwen3-reranker-8b

cli:
  verbose: true
  output_format: text
```

### Production Configuration

```yaml
# ~/.mem/config.yaml

memory:
  backend: postgres
  default_namespace: prod

postgres:
  host: db.production.example.com
  port: 5432
  database: mem_prod
  user: mem_user
  password: "${MEM_DB_PASSWORD}"
  sslmode: require

embeddings:
  base_url: https://api.openai.com/v1
  model: text-embedding-3-small
  dimensions: 1536
  api_key: "${OPENAI_API_KEY}"
  batch_size: 100

reranking:
  enabled: true
  top_k: 20
  threshold: 0.7

query:
  default_limit: 10
  min_similarity: 0.7

cli:
  output_format: json
  verbose: false
```

### Project-Specific Override

```yaml
# ./.memconfig (in project directory)

memory:
  default_namespace: myproject

embeddings:
  base_url: http://10.10.199.29:8080

query:
  default_limit: 10
```

## Configuration Precedence Examples

### Example 1: Namespace Resolution

Given:
- Config: `memory.default_namespace: global`
- `.memconfig`: `memory.default_namespace: project`
- CLI flag: `-n personal`

Result: Uses `personal` (CLI flag wins)

### Example 2: Output Format

Given:
- Config: `cli.output_format: text`
- CLI flag: `-o json`

Result: Uses `json` (CLI flag wins)

### Example 3: Backend Selection

Given:
- Config: `memory.backend: chromem`
- Environment: `MEM_BACKEND=postgres`

Result: Uses `postgres` (env var wins over config)

## Validating Configuration

```bash
# Show current effective configuration
mem config show

# Get specific value
mem config get memory.default_namespace
mem config get embeddings.model

# Test configuration
mem store "test configuration"
mem query "test"
```

## Common Issues

See [troubleshooting.md](troubleshooting.md) for configuration-related issues.