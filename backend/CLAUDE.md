# Backend

Go + Gin REST API for the learn_kids educational platform.

## Commands

```bash
go run ./cmd/api                           # Run API server locally (port 8080)
go test ./...                              # Run all tests
go test ./internal/domain/lessons/...     # Run tests for a specific package
```

Integration tests require a `.test.env` file at the project root (one level above `backend/`).

After making changes: run `go test ./...`. If you add or change an endpoint, update `api/openapi.yaml`.

## Architecture

Domain-driven layout under `internal/domain/`. Each domain has handler and repository; business logic lives in handlers (or in repository where it's simple).

```
internal/domain/{domain}/
├── handler.go    # HTTP layer: parse request, validate, call repository, write response
├── repository.go # Database queries (and optional model types)
└── model.go     # (optional) shared structs
```

Domains: `users`, `lessons`, `progress`, `comments`, `reactions`, `chat`

Other packages:
- **Entry point:** `cmd/api/main.go` — connects DB, runs migrations, starts Gin server
- **Routes & middleware:** `internal/server/` — CORS, JWT auth, rate limiting, logging
- **Migrations:** `migrations/` — SQL files applied automatically on startup
- **OpenAPI spec:** `api/openapi.yaml` — embedded in the binary

## Key Conventions

- Follow the handler → repository pattern; keep DB access in repositories, validation and orchestration in handlers
- Auth: sliding JWT sessions, 1-hour TTL, renewed when < 30 minutes remain
- Rate limiting on auth endpoints: 10 req/min per IP
- Sanitize all HTML input with `bluemonday` before writing to DB
- New domain features need: handler + service + repository + tests + migration (if schema changes)
- Do not break existing API contracts without discussion; check `../docs/context/api-contract.md`
