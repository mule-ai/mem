# mem Troubleshooting Guide

Common issues and their solutions when using `mem`.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Configuration Issues](#configuration-issues)
- [Storage Issues](#storage-issues)
- [Embedding Issues](#embedding-issues)
- [Query Issues](#query-issues)
- [Performance Issues](#performance-issues)
- [Error Messages](#error-messages)

## Installation Issues

### `go: module github.com/jbutlerdev/mem: not found`

**Problem:** The module is not published yet.

**Solution:** Install from source:

```bash
git clone https://github.com/jbutlerdev/mem.git
cd mem
go build -o mem ./cmd/mem
sudo mv mem /usr/local/bin/
```

### `command not found: mem`

**Problem:** The binary is not in your PATH.

**Solution:** Add to PATH or move to a standard location:

```bash
# Option 1: Add to PATH
echo 'export PATH=$PATH:$(pwd)' >> ~/.bashrc
source ~/.bashrc

# Option 2: Move to standard location
sudo mv mem /usr/local/bin/

# Verify
which mem
mem --version
```

### Build errors: `cannot find package`

**Problem:** Missing Go dependencies.

**Solution:** Download dependencies:

```bash
cd /path/to/mem
go mod download
go mod tidy
go build -o mem ./cmd/mem
```

## Configuration Issues

### `config file not found: ~/.mem/config.yaml`

**Problem:** No configuration file exists.

**Solution:** Create a default configuration:

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
  enabled: false

query:
  default_limit: 5
  min_similarity: 0.6
EOF
```

### `invalid configuration: cannot unmarshal yaml`

**Problem:** YAML syntax error in config file.

**Solution:** Validate YAML syntax:

```bash
# Install yamllint if needed
pip install yamllint

# Check syntax
yamllint ~/.mem/config.yaml

# Common issues:
# - Use spaces, not tabs (2 spaces recommended)
# - Ensure colons have space after them
# - Quote strings with special characters
```

Example of correct syntax:

```yaml
# ✓ Correct
memory:
  path: ~/.mem/data
  default_namespace: work

# ✗ Wrong (tabs, no space after colon)
memory:
	path: ~/.mem/data
```

### Configuration not being applied

**Problem:** Settings in config file are being ignored.

**Solution:** Check configuration precedence:

1. CLI flags override everything
2. `.memconfig` overrides `~/.mem/config.yaml`
3. Environment variables override config files

```bash
# Check what config is being used
mem config show

# Check for local config
ls -la .memconfig

# Check environment variables
env | grep MEM_
```

## Storage Issues

### `error opening database: directory not found`

**Problem:** Storage directory doesn't exist.

**Solution:** Create the directory:

```bash
# Check configured path
mem config get memory.path

# Create directory
mkdir -p ~/.mem/data
chmod 755 ~/.mem/data
```

### `permission denied: /path/to/mem-data`

**Problem:** Insufficient permissions on storage directory.

**Solution:** Fix permissions:

```bash
# Own the directory
sudo chown -R $USER:$USER ~/.mem/data

# Or allow read/write
chmod 755 ~/.mem/data
```

### `corrupted database: invalid collection`

**Problem:** chromem-go database is corrupted.

**Solution:** Reset the database:

```bash
# Export existing data first (if possible)
mem export backup.json

# Remove and recreate
rm -rf ~/.mem/data
mkdir -p ~/.mem/data

# Import data back
mem import backup.json
```

### PostgreSQL connection refused

**Problem:** Cannot connect to PostgreSQL.

**Solution:** Verify PostgreSQL is running and accessible:

```bash
# Check PostgreSQL is running
sudo systemctl status postgresql

# Test connection
psql -h localhost -U postgres -d mem

# Check config
mem config get postgres.host
mem config get postgres.port

# Common fixes:
sudo systemctl start postgresql
# or
docker start postgres
```

### `pgvector extension not found`

**Problem:** PostgreSQL doesn't have pgvector installed.

**Solution:** Install pgvector:

```bash
# For PostgreSQL 14+
git clone --branch v0.5.0 https://github.com/pgvector/pgvector.git
cd pgvector
make
sudo make install
```

Then enable in database:

```sql
CREATE EXTENSION vector;
```

## Embedding Issues

### `connection refused: localhost:11434`

**Problem:** Embedding API is not running.

**Solution:** Start your embedding service:

**For Ollama:**
```bash
ollama serve
# Verify: curl http://localhost:11434/api/tags
```

**For LM Studio:**
1. Open LM Studio
2. Start the server
3. Check the port in settings

**Check configured URL:**
```bash
mem config get embeddings.base_url

# Test connection
curl http://localhost:11434/api/tags
curl http://10.10.199.29:8080/v1/models
```

### `model not found: text-embedding-qwen3-embedding-8b`

**Problem:** The configured model is not available.

**Solution:** Check available models and update config:

```bash
# List available models (Ollama)
curl http://localhost:11434/api/tags

# List available models (LM Studio/OpenAI)
curl http://10.10.199.29:8080/v1/models

# Update config with available model
mem config set embeddings.model nomic-embed-text
```

Alternative: Edit config directly:

```yaml
embeddings:
  model: nomic-embed-text  # or other available model
```

### `embedding dimension mismatch`

**Problem:** Configured dimensions don't match model output.

**Solution:** Update dimensions to match your model:

```bash
# Check model documentation for dimensions
# Common values:
# - nomic-embed-text: 768
# - text-embedding-qwen3-embedding-8b: 1024
# - text-embedding-3-small: 1536
# - all-MiniLM-L6-v2: 384

mem config set embeddings.dimensions 768
```

### `timeout waiting for embedding response`

**Problem:** Embedding API is slow or overloaded.

**Solution:** Adjust batch size or timeout:

```yaml
embeddings:
  batch_size: 5  # Reduce batch size
```

Or check API performance:

```bash
# Test API speed
time curl -X POST http://localhost:11434/api/embeddings \
  -H "Content-Type: application/json" \
  -d '{"model": "nomic-embed-text", "input": "test"}'
```

### `401 unauthorized` or `403 forbidden`

**Problem:** API key is missing or invalid.

**Solution:** Set correct API key:

```bash
# Set environment variable
export MEM_API_KEY="sk-..."

# Or set in config
mem config set embeddings.api_key "sk-..."
```

For local deployments (Ollama, LM Studio), ensure `api_key` is empty:

```yaml
embeddings:
  api_key: ""  # Empty for local APIs
```

## Query Issues

### `no results found`

**Problem:** Query returns no memories.

**Possible causes:**

1. **Similarity threshold too high**
   ```bash
   # Lower threshold
   mem query --threshold 0.4 "your query"
   
   # Or set in config
   mem config set query.min_similarity 0.4
   ```

2. **Wrong namespace**
   ```bash
   # Check available namespaces
   mem namespace list
   
   # Try without namespace filter
   mem query "your query"
   ```

3. **No memories stored**
   ```bash
   # List all memories
   mem list
   
   # Store some if needed
   mem store "test memory"
   ```

4. **Query not specific enough**
   ```bash
   # Try different phrasing
   mem query "configuration settings"
   mem query "setup and config"
   ```

### Results not relevant

**Problem:** Query returns results, but they're not what you expected.

**Solutions:**

1. **Enable re-ranking** (if disabled):
   ```yaml
   reranking:
     enabled: true
   ```

2. **Increase top_k for re-ranking:**
   ```bash
   mem query --top 30 "your query"
   ```

3. **Try different query phrasing:**
   ```bash
   # Be more specific
   mem query "database connection timeout settings"
   
   # Use natural language
   mem query "how do I configure the database connection timeout"
   ```

4. **Check your memory content:**
   ```bash
   # List and review stored memories
   mem list -l 100
   ```

### Re-ranking not working

**Problem:** Results seem the same with or without re-ranking.

**Possible causes:**

1. **Reranker unavailable**
   ```bash
   # Check reranker endpoint
   curl http://localhost:11434/api/generate
   ```

2. **Reranker disabled**
   ```bash
   # Enable in config
   mem config set reranking.enabled true
   
   # Check it's enabled for queries
   mem config get query.rerank_enabled
   ```

3. **Invalid reranker model**
   ```yaml
   reranking:
     model: qwen3-reranker-8b  # Verify this model is available
   ```

## Performance Issues

### Slow query performance

**Problem:** Queries take more than 1-2 seconds.

**Solutions:**

1. **Disable re-ranking** (fastest):
   ```bash
   mem query --no-rerank "your query"
   ```

2. **Reduce top_k:**
   ```bash
   mem query --top 5 --limit 3 "your query"
   ```

3. **Increase similarity threshold:**
   ```bash
   mem query --threshold 0.8 "your query"
   ```

4. **Reduce batch size** (for stores):
   ```yaml
   embeddings:
     batch_size: 1
   ```

### Slow store performance

**Problem:** Storing memories takes more than 500ms.

**Solutions:**

1. **Check embedding API performance:**
   ```bash
   time curl -X POST http://localhost:11434/api/embeddings \
     -H "Content-Type: application/json" \
     -d '{"model": "nomic-embed-text", "input": "test"}'
   ```

2. **Reduce embedding dimensions** (if model supports):
   ```yaml
   embeddings:
     dimensions: 512  # If supported
   ```

3. **Batch multiple stores** (if storing many):
   ```bash
   # Store in parallel
   cat memories.txt | xargs -P 4 -I {} mem store "{}"
   ```

### High memory usage

**Problem:** `mem` process uses too much memory.

**Solutions:**

1. **Reduce batch size:**
   ```yaml
   embeddings:
     batch_size: 5
   ```

2. **Limit query results:**
   ```yaml
   query:
     default_limit: 5
   ```

3. **Use chromem-go backend** (more memory-efficient than PostgreSQL for small datasets)

## Error Messages

### `failed to generate embedding: context deadline exceeded`

**Timeout waiting for embedding API.**

Solution: Check API is running and responsive:

```bash
# Test connection
curl http://localhost:11434/api/tags

# Increase timeout by using smaller batches
mem config set embeddings.batch_size 1
```

### `failed to store memory: embedding dimension mismatch`

**Embedding vector size doesn't match configuration.**

Solution: Update dimensions to match model:

```bash
# Check what dimensions model outputs
# Update config
mem config set embeddings.dimensions 768
```

### `namespace not found: xyz`

**Trying to access non-existent namespace.**

Solution: Create namespace or use existing:

```bash
# List namespaces
mem namespace list

# Create namespace
mem namespace create xyz

# Or use default namespace
mem query "your query"
```

### `invalid memory ID format`

**Memory ID is not a valid UUID.**

Solution: Use correct memory ID from `mem list` or `mem query`:

```bash
# Get correct ID
mem list -o json | jq '.[0].memory.id'

# Use that ID
mem delete abc123-def456-...
```

### `yaml: unmarshal errors`

**Invalid YAML in metadata or configuration.**

Solution: Validate JSON/YAML:

```bash
# For metadata, ensure valid JSON
mem store -m '{"key": "value"}' "content"

# Validate your JSON
echo '{"key": "value"}' | jq .
```

## Debug Mode

Enable verbose output to diagnose issues:

```bash
# Enable verbose output
mem --verbose query "test"

# Or set in config
mem config set cli.verbose true
```

## Getting Help

If issues persist:

1. **Check the logs:**
   ```bash
   mem --verbose store "test" 2>&1 | tee mem-debug.log
   ```

2. **Verify configuration:**
   ```bash
   mem config show
   ```

3. **Test components individually:**
   ```bash
   # Test storage
   mem store "test"
   mem list
   
   # Test embeddings
   curl -X POST $MEM_BASE_URL/v1/embeddings \
     -H "Content-Type: application/json" \
     -d '{"model": "'$MEM_MODEL'", "input": "test"}'
   
   # Test query
   mem query "test"
   ```

4. **Report issues:**
   - Include: OS, Go version, `mem` version
   - Include: Error messages
   - Include: Configuration (with sensitive data removed)
   - Include: Steps to reproduce

## Useful Commands

```bash
# Health check
mem config show
mem store "health check"
mem query "health check"
mem delete --force --query "health check"

# Reset everything (careful!)
rm -rf ~/.mem/data
mem export backup.json  # Backup first!

# Start fresh
mkdir -p ~/.mem/data
mem store "first memory in new database"
```