# Neto

![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)
![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)
![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)
![Status](https://img.shields.io/badge/status-WIP-yellow)

> Personal finance manager designed for Latin America — multi-currency, budget tracking, debt/goal management, and an AI assistant.

## Structure

```
neto/
├── api/        # Go REST API — Chi router, Clean Architecture, Supabase
├── tui/        # Go terminal UI — Bubbletea
├── web/        # React web client — Vite
├── shared/     # OpenAPI spec and shared contracts
└── openspec/   # SDD artifacts (specs, design, tasks)
```

## Quick Start

```bash
# Install dependencies
make dev

# Run API
make api

# Run TUI
make tui

# Run web
make web
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| API | Go 1.22+, Chi, goose, pgx |
| Auth | Supabase JWT (JWKS) |
| DB | PostgreSQL via Supabase (RLS) |
| AI | Claude (Anthropic) — Tool Use |
| TUI | Bubbletea |
| Web | Vite + React 19 |

## License

MIT © Neto Team
