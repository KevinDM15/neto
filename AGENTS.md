# Neto — Contributor & Agent Guide

> Personal finance manager with AI agent. Open source MIT.

This document is the source of truth for anyone (human or AI agent) contributing to Neto.
Read it entirely before writing a single line of code.

---

## Project Overview

Neto is an open source personal finance manager that lets users track expenses, debts, budgets,
and goals through a conversational AI agent (Anthropic Claude). It ships two interfaces:

- **TUI** (`tui/`) — Terminal UI built with Go + Bubbletea. Primary interface for technical users.
- **Web** (`web/`) — Browser UI built with Vite + React + TypeScript. For non-technical users.

Both interfaces talk to a shared REST API built in Go.

**Stack:**

| Layer | Technology |
|-------|-----------|
| API | Go + Chi (Clean Architecture) |
| TUI | Go + Bubbletea |
| Web | Vite + React + TypeScript |
| Database | Supabase (PostgreSQL + Auth + RLS) |
| AI | Anthropic Claude (tool use) |
| Migrations | goose |
| Monorepo | plain directories (no Nx/Turborepo) |

---

## Repository Structure

```
neto/
├── api/                    # Go REST API
│   ├── internal/
│   │   ├── domain/         # Entities, value objects, repository interfaces (no deps)
│   │   ├── usecase/        # Application logic — orchestrates domain
│   │   ├── handler/        # HTTP handlers (Chi routes)
│   │   ├── repository/     # Supabase implementations
│   │   ├── middleware/     # JWT auth, idempotency, logging
│   │   └── ai/             # Claude client + tool use loop
│   ├── migrations/         # goose SQL migration files
│   ├── main.go
│   ├── go.mod
│   └── .env.example
├── tui/                    # Go TUI (Bubbletea)
│   ├── internal/
│   │   ├── ui/             # Bubbletea models and views
│   │   ├── client/         # HTTP client for the API
│   │   └── config/         # Local config (~/.config/neto/)
│   ├── main.go
│   └── go.mod
├── web/                    # Vite + React + TypeScript
│   ├── src/
│   │   ├── components/     # React components
│   │   ├── hooks/          # Custom hooks
│   │   ├── services/       # API client (typed, fetch-based)
│   │   └── main.tsx
│   ├── eslint.config.js
│   ├── tsconfig.json
│   └── package.json
├── shared/
│   └── openapi.yaml        # OpenAPI 3.1 — single source of truth for the API contract
├── openspec/               # SDD artifacts (specs, design, tasks) — do not edit manually
├── .github/
│   └── workflows/
│       └── ci.yml          # CI runs on every PR
├── .golangci.yml           # Go linter config
├── Makefile                # Common dev tasks
├── AGENTS.md               # This file
└── README.md
```

---

## Architecture

Neto follows **Clean Architecture** in the API. Dependencies always point inward:

```
HTTP Request
     │
     ▼
 handler/          ← knows about HTTP, calls use cases
     │
     ▼
 usecase/          ← orchestrates domain logic, calls repository interfaces
     │
     ▼
 domain/           ← pure Go structs, value objects, interfaces — NO external deps
     │
     ▼
 repository/       ← implements domain interfaces, talks to Supabase
```

**Rules:**
- `domain/` must NEVER import from `handler/`, `usecase/`, `repository/`, or any external package except the standard library.
- `usecase/` must NEVER import from `handler/` or `repository/` directly — only through interfaces defined in `domain/`.
- `handler/` is the only layer that knows about HTTP (Chi, `net/http`).

**TUI and Web are pure clients.** They call the API over HTTP. They contain no business logic.

---

## Git Flow

```
main          ← production releases only
  └── develop ← integration branch, always deployable
        ├── feature/xxx   ← new capabilities
        ├── fix/xxx        ← bug fixes on develop
        └── release/x.y.z ← release preparation

hotfix/xxx    ← branches from main, merged back to main AND develop
```

### Branch naming

| Type | Pattern | Example |
|------|---------|---------|
| Feature | `feature/<scope>-<description>` | `feature/core-domain` |
| Bug fix | `fix/<scope>-<description>` | `fix/api-idempotency-key` |
| Hotfix | `hotfix/<description>` | `hotfix/auth-token-leak` |
| Release | `release/<version>` | `release/1.0.0` |

### PR rules

- PRs target `develop` (never directly to `main` unless hotfix).
- CI must pass before merge — no exceptions.
- Squash merge into `develop` to keep history clean.
- `main` receives PRs only from `release/` or `hotfix/` branches.

---

## Commit Conventions

We use [Conventional Commits](https://www.conventionalcommits.org/).

```
<type>(<scope>): <short description>

[optional body]
```

### Types

| Type | When |
|------|------|
| `feat` | New feature |
| `fix` | Bug fix |
| `chore` | Tooling, deps, config (no production code) |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test` | Adding or updating tests |
| `docs` | Documentation only |
| `ci` | CI/CD changes |
| `perf` | Performance improvement |

### Scopes

Use the module name as scope: `api`, `tui`, `web`, `domain`, `db`, `ci`, `deps`.

### Examples

```
feat(api): add idempotency middleware for POST /transactions
fix(domain): correct balance calculation when currency differs
feat(tui): implement chat view with Bubbletea viewport
chore(deps): upgrade golangci-lint to v1.62
test(api): add RLS isolation tests for transactions
ci: add paths filter to skip unchanged modules
```

---

## Development Setup

### Prerequisites

- Go 1.22+
- Node 20+
- pnpm 10+: `npm install -g pnpm`
- [Supabase CLI](https://supabase.com/docs/guides/cli)
- [goose](https://github.com/pressly/goose): `go install github.com/pressly/goose/v3/cmd/goose@latest`
- [golangci-lint](https://golangci-lint.run): `brew install golangci-lint`

### First-time setup

```bash
# 1. Clone the repo
git clone https://github.com/KevinDM15/neto.git
cd neto

# 2. Copy env and fill in your values
cp api/.env.example api/.env

# 3. Run database migrations
cd api && goose -dir migrations postgres "$DATABASE_URL" up

# 4. Install web dependencies
cd ../web && pnpm install

# 5. Start everything
make dev
```

### Makefile targets

| Target | What it does |
|--------|-------------|
| `make api` | Build the API binary |
| `make tui` | Build the TUI binary |
| `make web` | Build the web client |
| `make dev` | Start API + web in watch mode |
| `make test` | Run all tests (Go + web) |
| `make lint` | Run all linters (golangci-lint + ESLint) |

---

## Code Standards

### Go

- Format: `gofmt` + `goimports` (enforced by CI).
- Linter: `golangci-lint` with the config in `.golangci.yml`.
- All exported symbols must have a doc comment.
- Errors must be handled — never `_` an error silently.
- Architecture comments (explaining *why*) may be written in Spanish.
- Code identifiers, variable names, and function names: always English.

### TypeScript / React

- Strict TypeScript (`strict: true` in tsconfig).
- ESLint with `--max-warnings 0` — zero tolerance.
- No `any` types. Use `unknown` and narrow properly.
- Functional components only. No class components.
- Custom hooks for all data fetching.

### General

- No secrets in code. Use environment variables.
- No `console.log` in production code (ESLint warns on it).
- All new API endpoints must be documented in `shared/openapi.yaml` first.

---

## CI/CD

Every PR triggers `.github/workflows/ci.yml`.

| Job | Trigger | Checks |
|-----|---------|--------|
| `go-check (api)` | Changes in `api/**` | vet, golangci-lint, build, test |
| `go-check (tui)` | Changes in `tui/**` | vet, golangci-lint, build, test |
| `web-check` | Changes in `web/**` | ESLint, type-check, build |

**A PR cannot be merged if CI fails.**

Jobs are skipped automatically if no files changed in their module — fast and cheap.

---

## Database

- **Engine**: PostgreSQL via Supabase.
- **Schema**: normalized to 3NF. No data duplication.
- **Migrations**: managed with `goose`. Files live in `api/migrations/`.
- **RLS**: every table has Row Level Security policies. Users can only access their own data.
- **Seed data**: currencies are seeded via `api/migrations/seed_currencies.sql`.

### Migration naming

```
YYYYMMDDHHMMSS_description.sql

20240101000001_create_accounts.sql
20240101000002_create_transactions.sql
```

### Adding a new migration

```bash
cd api
goose -dir migrations create <description> sql
# Edit the generated file
goose -dir migrations postgres "$DATABASE_URL" up
```

---

## AI Agent

The AI agent uses **Anthropic Claude with tool use**. Architecture:

1. User sends a message via TUI or Web → `POST /api/v1/chat`
2. API sends the message to Claude with a system prompt and the tool catalog
3. Claude decides which tool(s) to call (e.g., `create_transaction`, `list_expenses`)
4. API executes the tool (mutates Supabase), sends results back to Claude
5. Claude generates a final response in natural language
6. API returns the response to the client

**Tool use rules:**
- Max 5 iterations per request to prevent infinite loops.
- Destructive operations (delete, bulk update) require explicit user confirmation before execution.
- Each tool call is wrapped in an idempotency key to prevent duplicate mutations on retry.

### Adding a new tool

1. Define the tool schema in `api/internal/ai/tools.go`
2. Implement the handler in `api/internal/ai/handlers.go`
3. Add the corresponding use case in `api/internal/usecase/`
4. Document the tool in `shared/openapi.yaml`
5. Write a test for the tool handler

---

## Contributing

1. Fork the repo and clone locally.
2. Create a branch from `develop`: `git checkout -b feature/your-feature develop`
3. Make your changes following the standards above.
4. Ensure CI passes locally: `make lint && make test`
5. Open a PR targeting `develop` with a clear description.
6. Wait for CI to pass and request a review.

### Reporting bugs

Open a GitHub Issue with:
- What you expected
- What happened instead
- Steps to reproduce
- Environment (OS, Go version, Node version)

### Proposing features

Open a GitHub Issue tagged `proposal` before writing any code.
Describe the problem you're solving, not just the solution.
