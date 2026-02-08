# mem Example Workflows

Practical examples and workflows for common use cases.

## Table of Contents

- [Personal Knowledge Management](#personal-knowledge-management)
- [Development Workflow](#development-workflow)
- [Meeting Notes](#meeting-notes)
- [Command Library](#command-library)
- [Recipe Collection](#recipe-collection)
- [Research Notes](#research-notes)
- [Project Documentation](#project-documentation)
- [Backup & Migration](#backup--migration)

## Personal Knowledge Management

### Setting Up a Personal Knowledge Base

```bash
# Create configuration
cat > ~/.mem/config.yaml << 'EOF'
memory:
  path: ~/.mem/data
  default_namespace: personal

embeddings:
  base_url: http://localhost:11434
  model: nomic-embed-text
  dimensions: 768

cli:
  output_format: text
EOF
```

### Storing Daily Notes

```bash
# Store things learned
mem store -t learning "Go's context package is used for cancellation and deadlines across API boundaries"

mem store -t tip "In Vim, use ciw to change inner word, diw to delete inner word"

# Store ideas
mem store -t idea "Build a CLI tool that combines git notes with semantic search"
```

### Querying Your Knowledge

```bash
# Find what you learned about Go
mem query "context cancellation go"

# Find all tips
mem list -t tip

# Find ideas
mem query "cli tools semantic search"
mem list -t idea
```

## Development Workflow

### Project-Specific Memory

```bash
# In your project directory
cat > .memconfig << 'EOF'
memory:
  default_namespace: myproject

embeddings:
  base_url: http://10.10.199.29:8080
  model: text-embedding-qwen3-embedding-8b
  dimensions: 1024
EOF
```

### Storing Development Decisions

```bash
# Architecture decisions
mem store -t architecture -t api "API Gateway uses JWT authentication with 15-minute token expiration"

mem store -t architecture -t database "User data stored in PostgreSQL, cached in Redis for hot data"

# Bug tracking
mem store -t bug -m '{"status": "open", "priority": "high"}' "Race condition in payment processing under high load"

mem store -t bug -m '{"status": "fixed", "ticket": "DEV-123"}' "Fixed memory leak in WebSocket handler by closing connections properly"

# Configuration notes
mem store -t config "Development database: postgresql://localhost:5432/devdb"

mem store -t config "LM Studio for embeddings: http://10.10.199.29:8080"
```

### Querying Development Context

```bash
# Find architecture info
mem query "database user storage"

# Find open bugs
mem list -t bug -o json | jq '.[] | select(.memory.metadata.status == "open")'

# Find config details
mem query "database connection development"

# Search by ticket number
mem query "DEV-123"
```

### Integration with Git Workflow

```bash
# Store commit message context
git log --oneline -1 | mem store -t git -m '{"branch": "'$(git branch --show-current)'"}' -

# Store before major refactor
mem store -t milestone -m '{"date": "'$(date -I)'"}' "Started payment system refactor - moved from monolith to microservices"

# Query context before working on feature
mem query "payment architecture"

# Store resolved issues
git log --grep="fix" --oneline -10 | while read commit; do
  echo "$commit" | mem store -t fixed -
done
```

## Meeting Notes

### Setup for Meeting Notes

```bash
# Create meetings namespace
mem namespace create meetings
```

### Storing Meeting Notes

```bash
# Daily standup
mem store -n meetings -t standup -m '{"date": "2026-02-07", "attendees": ["alice", "bob", "carol"]}' \
  "Standup: Alice working on auth API, Bob debugging payment issue, Carol deploying to staging. Blocker: Bob waiting on database credentials."

# Sprint planning
mem store -n meetings -t planning -m '{"date": "2026-02-07", "sprint": "Sprint 14"}' \
  "Sprint 14 Planning: Focus on MVP launch. Key features: user auth, dashboard, reports. Target date: March 15th."

# Retrospective
mem store -n meetings -t retro -m '{"sprint": "Sprint 13", "date": "2026-02-07"}' \
  "Sprint 13 Retro: Good progress on frontend. Need better test coverage. Action: Add integration tests next sprint."

# Decision records
mem store -n meetings -t decision -m '{"decision": "001", "date": "2026-02-07"}' \
  "Decision 001: Use PostgreSQL as primary database with Redis caching. Rejected MongoDB due to lack of transaction support needed for payments."
```

### Querying Meeting Notes

```bash
# Find all standups
mem list -n meetings -t standup

# Find decisions about database
mem query -n meetings "database decision"

# Find what was decided in sprint 13
mem query -n meetings "sprint 13"

# Search retrospective action items
mem query -n meetings "action item test coverage"

# Find all meetings from specific date
mem list -n meetings -o json | jq '.[] | select(.memory.metadata.date == "2026-02-07")'
```

## Command Library

### Building a Command Reference

```bash
# Create commands namespace
mem namespace create commands
```

### Storing Commands

```bash
# Docker commands
mem store -n commands -t docker "Remove all stopped containers: docker container prune"

mem store -n commands -t docker "View container logs: docker logs -f <container-name>"

mem store -n commands -t docker "Execute command in container: docker exec -it <container> bash"

# Git commands
mem store -n commands -t git "Undo last commit but keep changes: git reset --soft HEAD~1"

mem store -n commands -t git "Show file history: git log --follow --patch -- <file>"

mem store -n commands -t git "Clean untracked files: git clean -fd"

# System administration
mem store -n commands -t sysadmin "Find large files: find . -type f -size +100M -exec ls -lh {} \;"

mem store -n commands -t sysadmin "Monitor disk usage: watch -n 1 df -h"

# Video processing
mem store -n commands -t video "Resize video: ffmpeg -i input.mp4 -vf scale=1280:-1 output.mp4"

mem store -n commands -t video "Extract audio: ffmpeg -i input.mp4 -vn -acodec copy output.aac"
```

### Finding Commands

```bash
# Find Docker container commands
mem query -n commands "docker container"

# Find video resize command
mem query -n commands "resize video ffmpeg"

# List all git commands
mem list -n commands -t git

# Find system monitoring commands
mem query -n commands "monitor disk usage"
```

### Quick Command Retrieval

```bash
# Create an alias for quick command lookup
alias cmd='mem query -n commands -l 1 -o json | jq -r ".[0].memory.content"'

# Usage
cmd docker remove containers
cmd git undo commit
```

## Recipe Collection

### Organizing Recipes

```bash
# Create recipes namespace
mem namespace create recipes
```

### Storing Recipes

```bash
# Desserts
mem store -n recipes -t dessert -t chocolate "Chocolate Cake: 2 cups flour, 2 cups sugar, 3/4 cup cocoa, 2 eggs, 1 cup milk, 1/2 cup butter, 1 cup boiling water. Mix dry, add wet, pour into 9x13 pan. Bake at 350°F for 30-35 minutes."

mem store -n recipes -t dessert "Tiramisu: Ladyfingers, espresso, mascarpone cheese, eggs, sugar, cocoa powder. Dip ladyfingers in coffee, layer with cheese mixture, dust with cocoa. Chill 4+ hours."

# Main dishes
mem store -n recipes -t dinner -t pasta "Spaghetti Carbonara: 400g spaghetti, 200g pancetta, 4 egg yolks, 100g pecorino, black pepper. Cook pasta. Fry pancetta. Mix yolks with cheese. Combine all with pasta water."

mem store -n recipes -t dinner -t chicken "Lemon Herb Chicken: 4 chicken breasts, lemon juice, olive oil, garlic, rosemary, thyme. Marinate 2 hours. Bake at 400°F for 25-30 minutes."

# Quick meals
mem store -n recipes -t quick "Avocado Toast: Bread, ripe avocado, salt, red pepper flakes, lemon juice. Toast bread, mash avocado with seasonings, spread on toast."
```

### Finding Recipes

```bash
# Find chocolate desserts
mem query -n recipes "chocolate dessert"

# Find quick meals
mem list -n recipes -t quick

# Find chicken recipes
mem query -n recipes "chicken dinner"

# Find pasta dishes
mem query -n recipes "pasta italian"

# Browse all desserts
mem list -n recipes -t dessert
```

## Research Notes

### Academic Research

```bash
# Create research namespace
mem namespace create research
```

### Storing Research Notes

```bash
# Paper summaries
mem store -n research -t paper -m '{"authors": "Smith et al.", "year": 2024, "venue": "ICML"}' \
  "Attention mechanisms allow models to focus on relevant parts of input. Key innovation: softmax over query-key pairs to compute attention weights. Enables better handling of long sequences."

mem store -n research -t paper -m '{"authors": "Johnson et al.", "year": 2023, "venue": "NeurIPS"}' \
  "Vector databases enable efficient similarity search. Use HNSW indexing for O(log n) lookup. Critical for RAG applications at scale."

# Concept definitions
mem store -n research -t concept "RAG (Retrieval-Augmented Generation): Combines retrieval of relevant documents with generation. Improves factual accuracy and reduces hallucination in LLMs."

mem store -n research -t concept "Embedding: Dense vector representation of text. Captures semantic meaning. Similar concepts have similar embeddings (high cosine similarity)."

# Experiment results
mem store -n research -t experiment -m '{"date": "2026-02-07", "model": "gpt-4"}' \
  "Experiment: Prompt engineering for code generation. Result: Few-shot examples with test cases improved pass@1 from 45% to 62%. Chain-of-thought not helpful for code."

# Literature review notes
mem store -n research -t review "Recent trends in LLM: 1) Larger context windows (128k+ tokens), 2) Multimodal capabilities, 3) Tool use/function calling, 4) Efficient fine-tuning (LoRA, QLoRA)."
```

### Querying Research

```bash
# Find papers about attention
mem query -n research "attention mechanism"

# Find experiment results
mem list -n research -t experiment

# Find RAG-related information
mem query -n research "retrieval augmented generation"

# Look up concept definitions
mem query -n research "what is embedding"

# Find all papers from 2024
mem list -n research -o json | jq '.[] | select(.memory.metadata.year == 2024)'
```

## Project Documentation

### Team Knowledge Base

```bash
# Create project namespace
mem namespace create project-docs
```

### Storing Project Documentation

```bash
# Onboarding
mem store -n project-docs -t onboarding "Dev environment setup: Clone repo, run 'docker-compose up' to start services. DB runs on port 5432, API on 8080. Install dependencies with 'npm install'. Run tests with 'npm test'."

mem store -n project-docs -t onboarding "Code review process: Create PR, get 2 approvals, pass CI checks. Required checks: lint, tests, security scan. Merge via squash commits."

# Deployment
mem store -n project-docs -t deployment "Production deployment: 1) Merge to main, 2) CI builds Docker image, 3) Push to registry, 4) ArgoCD syncs to prod cluster. Downtime < 30s via rolling updates."

mem store -n project-docs -t deployment "Rollback procedure: kubectl rollout undo deployment/app-name. Or use ArgoCD UI to revert to previous revision. Monitor logs via kubectl logs -f."

# Team contacts
mem store -n project-docs -t team -m '{"role": "backend-lead"}' "Backend team lead: Alice (alice@company.com), expert in distributed systems"

mem store -n project-docs -t team -m '{"role": "devops"}' "DevOps: Bob (bob@company.com), manages Kubernetes and CI/CD"

# Incident response
mem store -n project-docs -t incident -m '{"severity": "p1"}' "P1 Incident Response: 1) Page on-call, 2) Create Slack channel, 3) Post updates every 15 mins, 4) Write post-mortem within 24 hours."

# API documentation snippets
mem store -n project-docs -t api "POST /api/users - Create user. Body: {name, email, password}. Returns: {id, name, email, created_at}. Rate limit: 10/minute."

mem store -n project-docs -t api "GET /api/users/:id - Get user by ID. Returns 404 if not found. Cached for 5 minutes."
```

### Querying Project Docs

```bash
# New developer onboarding
mem query -n project-docs "how to set up development environment"

# Deployment questions
mem query -n project-docs "how to deploy to production"

# Find team contacts
mem list -n project-docs -t team

# Incident procedures
mem query -n project-docs "p1 incident response"

# API reference
mem query -n project-docs "create user api endpoint"
mem list -n project-docs -t api
```

## Backup & Migration

### Regular Backups

```bash
# Daily backup script
cat > ~/backup-mem.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="$HOME/backups/mem"
DATE=$(date +%Y%m%d)

mkdir -p "$BACKUP_DIR"

# Export all memories
mem export "$BACKUP_DIR/mem-$DATE.json"

# Keep last 30 days
find "$BACKUP_DIR" -name "mem-*.json" -mtime +30 -delete

echo "Backup complete: $BACKUP_DIR/mem-$DATE.json"
EOF

chmod +x ~/backup-mem.sh

# Add to crontab for daily backups
crontab -e
# Add: 0 2 * * * ~/backup-mem.sh
```

### Migration Between Machines

```bash
# On source machine
cd /data/jbutler/git/jbutlerdev/mem
mem export migration.json

# Copy to destination
scp migration.json user@destination:/tmp/

# On destination machine
mem import /tmp/migration.json

# Verify
mem list
```

### Namespace Migration

```bash
# Export specific namespace
mem export -n work work-backup.json

# Import to different namespace
mem import -n work-archive work-backup.json

# Verify
mem list -n work-archive
```

### Bulk Operations

```bash
# Bulk store from file
cat memories.txt | while read line; do
  mem store "$line"
done

# Bulk delete by tag
for id in $(mem list -t old -o json | jq -r '.[].memory.id'); do
  mem delete --force $id
done

# Bulk update tags
for id in $(mem list -t todo -o json | jq -r '.[].memory.id'); do
  mem update $id -t in-progress
done
```

## Advanced Workflows

### Integration with Other Tools

```bash
# Store links from browser
# Browser bookmarklet:
# javascript:location.href='mem://store?url='+encodeURIComponent(location.href)+'&title='+encodeURIComponent(document.title)

# Store from notes app
# Create a script that watches a file and adds lines to mem
tail -f notes.txt | while read line; do
  mem store "$line"
done

# Store from Slack/Discord
# Use webhook to trigger mem store
curl -X POST http://localhost:8080/store \
  -H "Content-Type: application/json" \
  -d '{"content": "Important announcement", "namespace": "work"}'
```

### Periodic Review

```bash
# Daily review script
cat > ~/daily-review.sh << 'EOF'
#!/bin/bash
echo "=== Memories added today ==="
mem list -o json | jq -r '.[] | select(.memory.created_at | startswith("'$(date +%Y-%m-%d)'")) | .memory.content'

echo ""
echo "=== High priority items ==="
mem list -o json | jq '.[] | select(.memory.metadata.priority == "high")'

echo ""
echo "=== In-progress items ==="
mem list -t in-progress
EOF
```

### Tag Management

```bash
# Rename a tag
for id in $(mem list -t oldtag -o json | jq -r '.[].memory.id'); do
  # Get current tags
  tags=$(mem list -o json | jq -r '.[] | select(.memory.id == "'$id'") | .memory.tags | join(",")')
  # Remove old tag, add new one
  newtags=$(echo $tags | sed 's/oldtag/newtag/g')
  mem update $id -t $newtags
done
```