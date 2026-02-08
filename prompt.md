# Task: Implement Next Phase

You are an expert coding agent working on the `mem` (Semantic Memory CLI) project. Your goal is to incrementally build the system based on the specification and implementation plan.

## Instructions

1. **Context & Status**:
   - Read `spec.md` to understand the system architecture, data models, CLI requirements, and configuration.
   - Read `plan.md` to identify the current progress of the project.

2. **Identify Task**:
   - Find the **first** incomplete phase or task in `plan.md` (look for unchecked items like `- [ ]` or sections that haven't been started).
   - If the plan doesn't use checkboxes, look for the first phase that hasn't been implemented yet based on the file structure.

3. **Implementation**:
   - Implement **only** that specific phase or task.
   - Follow the architectural guidelines in `spec.md`.
   - Create necessary directories, files, and code.
   - Ensure the code compiles (Go) or runs without immediate errors.

4. **Key Requirements**:
   - **Language**: Go (Golang)
   - **Vector Database**: chromem-go (primary), PostgreSQL + pgvector (provisional)
   - **Embeddings**: OpenAI-compatible API
   - **Development Endpoint**: http://10.10.199.29:8080 (LM Studio)
   - **Embedding Model**: `text-embedding-qwen3-embedding-8b`
   - **Reranker Model**: `qwen3-reranker-8b`
   - **Config Files**: `~/.mem/config.yaml` (global), `./.memconfig` (local override)

5. **Update Progress**:
   - Once the implementation is complete and verified, **edit `plan.md`**.
   - Mark the completed task/phase as done (e.g., change `- [ ]` to `- [x]`).

6. **Stop**:
   - Do not proceed to the next phase. Return control so the next agent can verify the work and pick up the next task.

## Project Structure Reference

```
mem/
├── cmd/
│   └── mem/
│       └── main.go
├── internal/
│   ├── config/
│   ├── storage/
│   ├── embeddings/
│   ├── reranker/
│   ├── cli/
│   └── models/
├── pkg/
│   └── api/
├── examples/
├── spec.md
├── plan.md
└── README.md
```

## Important Notes

- Always test against LM Studio at http://10.10.199.29:8080 during development
- Use the exact model names specified in the spec
- Follow the configuration precedence: CLI flags > .memconfig > ~/.mem/config.yaml > defaults
- Phase 7 (End-to-End Integration Test) must run against real chromem-go database with LM Studio models