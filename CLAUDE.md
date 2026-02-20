# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Educational platform for teaching robotics (Arduino/Raspberry Pi) to kids. Monorepo with a Go (Gin) backend and React (Vite) frontend. UI and content are in Russian.

## Commands

### Full Stack (Docker)

```bash
docker compose up --build   # Build and run everything
docker compose up -d db     # Start only the database
```

### Environment

Copy `.env.example` to `.env`. Key variables:
- `DATABASE_URL` — PostgreSQL connection string
- `JWT_SECRET` — JWT signing secret
- `FRONTEND_ORIGIN` — CORS origin (e.g. `http://localhost:8888`)
- `GEMINI_API_KEY` — optional, enables lesson AI chat
- `TEACHER_INVITE_CODE` — optional, gates teacher registration

Backend integration tests require a `.test.env` file at the project root (one level above `backend/`).

## Architecture

### Request Flow

```
Browser → Nginx :8888 → [static files OR /api → Go API :8080] → PostgreSQL :5432
```

Nginx serves the built frontend (from `frontend/dist/` copied to `api/web/`) and proxies `/api/*` to the Go backend. In development, Vite's dev server proxies `/api` to `:8080` directly.

### Backend (`backend/`)

Domain-driven layout under `internal/domain/`. Domains: `users`, `lessons`, `progress`, `comments`, `reactions`, `chat`. Each domain follows handler → service → repository pattern. See [backend/CLAUDE.md](backend/CLAUDE.md) for details.

### Frontend (`frontend/src/`)

Feature-based structure under `features/`. React Router v7, i18next (Russian + English), react-markdown for lesson content. See [frontend/CLAUDE.md](frontend/CLAUDE.md) for details.

## Development Rules

**Every new feature must include tests.** This is non-negotiable:

- Backend: add tests in the same package (`*_test.go`) covering the handler and service logic
- Frontend: add Vitest tests for new components and utility functions

Do not consider a feature complete until its tests pass.

## Domain Documentation

`docs/context/` contains detailed documentation for each domain. **Read the relevant file before making changes to a domain, and update it after changing API contracts, request/response formats, or behavior.** Files:

- `lessons.md` — catalog structure (modules → lessons → steps), material types
- `auth.md` — JWT, sliding sessions, roles (student/teacher)
- `progress.md` — lesson completion and checklist tracking
- `chat.md` — Gemini AI integration
- `comments.md` / `reactions.md` — social features
- `parent-dashboard.md` — teacher/parent access to student progress
- `api-contract.md` — full API endpoint summary
- `i18n.md` — translation setup and usage
