# Design: Proyecto Neto (MVP)

## Technical Approach

Go monorepo con Clean Architecture. La API Go sirve REST+WebSocket para el TUI (Bubbletea) y el Web (Next.js 15). Supabase como DB con RLS por `user_id`. Claude Tool Use orquestado server-side en Go. Toda mutación es idempotente vía middleware + tabla `idempotency_keys`.

## 1. Estructura del Monorepo

```
neto/
├── api/                          # Go API server
│   ├── cmd/server/main.go        # Entrypoint
│   ├── internal/
│   │   ├── handler/              # HTTP handlers (adapters in)
│   │   ├── middleware/           # Auth JWT, idempotency, logging
│   │   ├── usecase/              # Application services
│   │   ├── domain/               # Entities, value objects, interfaces
│   │   │   ├── entity/
│   │   │   ├── repository/       # Port interfaces
│   │   │   └── valueobject/
│   │   ├── infrastructure/       # Adapters out
│   │   │   ├── supabase/         # Repository implementations
│   │   │   ├── anthropic/        # Claude client + tool defs
│   │   │   └── exchangerate/     # External rate API client
│   │   └── config/               # Env vars, app config
│   ├── migrations/               # SQL migrations (goose)
│   └── go.mod
├── tui/                          # Go TUI client (Bubbletea)
│   ├── cmd/neto/main.go          # Entrypoint
│   ├── internal/
│   │   ├── ui/                   # Bubbletea models (screens)
│   │   ├── client/               # HTTP client to api/
│   │   └── config/               # TUI config (API URL, auth token)
│   └── go.mod
├── web/                          # Next.js 15 frontend
│   ├── src/app/                  # App Router
│   ├── src/lib/api-client.ts     # Typed client from OpenAPI
│   └── package.json
├── shared/                       # Contratos compartidos
│   ├── openapi.yaml              # OpenAPI 3.1 spec (source of truth)
│   └── proto/                    # Futuro: gRPC si se necesita
└── openspec/                     # SDD artifacts
```

Go modules: `api/` y `tui/` son módulos Go independientes (no Go workspace por simplicidad). El TUI consume la API vía HTTP — NO importa paquetes de `api/` directamente.

## 2. Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|-------------|-----------|
| HTTP framework | **Chi** | Gin, Echo | Stdlib-compatible (`net/http`), zero-magic, composable middleware. Gin agrega abstractions innecesarias para una API que es thin adapter. |
| DB migrations | **goose** | golang-migrate, atlas | SQL puro, sin DSL. Supabase-friendly. |
| Auth | **Supabase JWT validated in Go middleware** | Auth0, custom JWT | Ya usamos Supabase DB — evitar vendor extra. El middleware Go valida el JWT con la JWKS pública de Supabase. |
| Idempotency | **Middleware + tabla `idempotency_keys`** | Application-level per-endpoint | Cross-cutting concern → middleware. Respuesta cacheada 24h, luego purgada. |
| AI orchestration | **Server-side tool use loop en Go** | Client-side, edge functions | Control total del loop, logging, rate limiting. El TUI y Web son thin clients. |
| OpenAPI contract | **Spec-first en `shared/openapi.yaml`** | Code-first con swag | Single source of truth para Go server + TS client + TUI client. |

## 3. Schema de DB (3NF) — Supabase/PostgreSQL

### `users` (Supabase Auth — tabla `auth.users`)
Managed by Supabase. Referenciamos `auth.users.id` como FK.

### `currencies`
```sql
CREATE TABLE currencies (
  code       TEXT PRIMARY KEY,           -- 'COP', 'USD', 'EUR'
  name       TEXT NOT NULL,
  symbol     TEXT NOT NULL,              -- '$', '€'
  decimals   SMALLINT NOT NULL DEFAULT 2
);
-- Seed: COP, ARS, MXN, BRL, CLP, PEN, UYU, BOB, PYG, USD, EUR
```

### `exchange_rates`
```sql
CREATE TABLE exchange_rates (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  base_code    TEXT NOT NULL REFERENCES currencies(code),
  target_code  TEXT NOT NULL REFERENCES currencies(code),
  rate         NUMERIC(18,8) NOT NULL,
  fetched_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(base_code, target_code, fetched_at)
);
CREATE INDEX idx_exchange_rates_pair ON exchange_rates(base_code, target_code, fetched_at DESC);
-- RLS: READ para cualquier usuario autenticado (datos públicos). No write desde client.
```

### `accounts`
```sql
CREATE TABLE accounts (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  type          TEXT NOT NULL CHECK (type IN ('checking','savings','cash','credit_card','investment','crypto')),
  currency_code TEXT NOT NULL REFERENCES currencies(code),
  balance       NUMERIC(18,2) NOT NULL DEFAULT 0,
  is_active     BOOLEAN NOT NULL DEFAULT true,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_accounts_user ON accounts(user_id);
-- RLS: ALL ops WHERE user_id = auth.uid()
```

### `categories`
```sql
CREATE TABLE categories (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID REFERENCES auth.users(id) ON DELETE CASCADE, -- NULL = system default
  parent_id  UUID REFERENCES categories(id) ON DELETE CASCADE,
  name       TEXT NOT NULL,
  icon       TEXT,
  type       TEXT NOT NULL CHECK (type IN ('expense','income')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id, parent_id, name)
);
CREATE INDEX idx_categories_user ON categories(user_id);
CREATE INDEX idx_categories_parent ON categories(parent_id);
-- RLS: WHERE user_id = auth.uid() OR user_id IS NULL (system defaults visible)
```

### `transactions`
```sql
CREATE TABLE transactions (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  account_id      UUID NOT NULL REFERENCES accounts(id) ON DELETE RESTRICT,
  category_id     UUID REFERENCES categories(id) ON DELETE SET NULL,
  type            TEXT NOT NULL CHECK (type IN ('expense','income','transfer')),
  amount          NUMERIC(18,2) NOT NULL CHECK (amount > 0),
  currency_code   TEXT NOT NULL REFERENCES currencies(code),
  description     TEXT,
  transaction_date DATE NOT NULL DEFAULT CURRENT_DATE,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_tx_user_date ON transactions(user_id, transaction_date DESC);
CREATE INDEX idx_tx_account ON transactions(account_id);
CREATE INDEX idx_tx_category ON transactions(category_id);
-- RLS: ALL ops WHERE user_id = auth.uid()
```

### `budgets`
```sql
CREATE TABLE budgets (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  category_id   UUID NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  amount_limit  NUMERIC(18,2) NOT NULL CHECK (amount_limit > 0),
  currency_code TEXT NOT NULL REFERENCES currencies(code),
  period_type   TEXT NOT NULL CHECK (period_type IN ('monthly','weekly','yearly')),
  period_start  DATE NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id, category_id, period_type, period_start)
);
CREATE INDEX idx_budgets_user_period ON budgets(user_id, period_start);
-- RLS: ALL ops WHERE user_id = auth.uid()
```

### `debts`
```sql
CREATE TABLE debts (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  name            TEXT NOT NULL,
  type            TEXT NOT NULL CHECK (type IN ('owe','owed')),  -- yo debo / me deben
  counterparty    TEXT NOT NULL,
  original_amount NUMERIC(18,2) NOT NULL CHECK (original_amount > 0),
  remaining       NUMERIC(18,2) NOT NULL CHECK (remaining >= 0),
  currency_code   TEXT NOT NULL REFERENCES currencies(code),
  due_date        DATE,
  is_settled      BOOLEAN NOT NULL DEFAULT false,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_debts_user ON debts(user_id);
-- RLS: ALL ops WHERE user_id = auth.uid()
```

### `goals`
```sql
CREATE TABLE goals (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  name            TEXT NOT NULL,
  target_amount   NUMERIC(18,2) NOT NULL CHECK (target_amount > 0),
  current_amount  NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (current_amount >= 0),
  currency_code   TEXT NOT NULL REFERENCES currencies(code),
  target_date     DATE,
  is_completed    BOOLEAN NOT NULL DEFAULT false,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_goals_user ON goals(user_id);
-- RLS: ALL ops WHERE user_id = auth.uid()
```

### `idempotency_keys`
```sql
CREATE TABLE idempotency_keys (
  key           UUID PRIMARY KEY,
  user_id       UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  endpoint      TEXT NOT NULL,
  status_code   SMALLINT NOT NULL,
  response_body JSONB NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_idemp_cleanup ON idempotency_keys(created_at);
-- Purge: cron diario DELETE WHERE created_at < now() - interval '24 hours'
-- RLS: WHERE user_id = auth.uid()
```

### `ai_conversations`
```sql
CREATE TABLE ai_conversations (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  title      TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ai_messages (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  conversation_id UUID NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
  role            TEXT NOT NULL CHECK (role IN ('user','assistant','tool_use','tool_result')),
  content         JSONB NOT NULL,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_ai_msg_conv ON ai_messages(conversation_id, created_at);
-- RLS: ai_conversations WHERE user_id = auth.uid()
-- RLS: ai_messages via JOIN a ai_conversations.user_id
```

### RLS Pattern (aplicado a TODAS las tablas con `user_id`)
```sql
ALTER TABLE {table} ENABLE ROW LEVEL SECURITY;
CREATE POLICY "{table}_user_isolation" ON {table}
  USING (user_id = auth.uid())
  WITH CHECK (user_id = auth.uid());
```

## 4. Data Flow: TUI → API → Supabase

```
┌──────┐    HTTP/JSON     ┌──────────────────────────────────┐
│ TUI  │ ───────────────→ │ Go API (Chi)                     │
│Bubble│ ←─────────────── │                                  │
│ tea  │                  │  ┌─────────┐  ┌──────────┐       │
└──────┘                  │  │Auth MW  │→ │Idemp MW  │       │
                          │  └─────────┘  └──────────┘       │
┌──────┐    HTTP/JSON     │       ↓                          │
│ Web  │ ───────────────→ │  ┌─────────────┐                 │
│Next  │ ←─────────────── │  │  Handler    │                 │
└──────┘                  │  └──────┬──────┘                 │
                          │         ↓                        │
                          │  ┌─────────────┐  ┌───────────┐  │
                          │  │  Use Case   │→ │ Anthropic  │  │
                          │  └──────┬──────┘  │ Tool Loop  │  │
                          │         ↓         └───────────┘  │
                          │  ┌─────────────┐                 │
                          │  │ Repository  │                 │
                          │  └──────┬──────┘                 │
                          └─────────┼────────────────────────┘
                                    ↓
                          ┌──────────────────┐
                          │   Supabase DB    │
                          │   (RLS active)   │
                          └──────────────────┘
```

### Secuencia: Chat AI con mutación
```
User ──"Gasté 50k en luz"──→ TUI
TUI ──POST /api/chat {message, idempotency_key}──→ API
API: Auth MW valida JWT
API: Idemp MW busca key → miss
API: ChatUseCase.Send(msg)
  → Anthropic API: messages.create(tools=[...], messages=[...])
  ← tool_use: {name: "create_transaction", input: {amount:50000, ...}}
  → API genera confirmación pendiente
  ← Response: "¿Confirmo? 50,000 COP en Servicios/Luz"
API ──response──→ TUI ──muestra confirmación──→ User
User ──"sí"──→ TUI
TUI ──POST /api/chat {message:"sí", idempotency_key_2}──→ API
API: ChatUseCase continúa el tool loop
  → TransactionRepo.Create(tx) con RLS user_id
  → Anthropic: tool_result {success: true}
  ← "Listo, registré $50,000 COP en Servicios/Luz"
API: Idemp MW guarda response
API ──response──→ TUI
```

## 5. Claude Tool Use Integration

### Tools expuestos al agente

| Tool | Tipo | Confirmación |
|------|------|-------------|
| `create_transaction` | mutación | ✅ Siempre |
| `list_transactions` | lectura | ❌ |
| `get_balance` | lectura | ❌ |
| `create_account` | mutación | ✅ Siempre |
| `update_budget` | mutación | ✅ Si diff > 20% |
| `list_categories` | lectura | ❌ |
| `get_monthly_summary` | lectura | ❌ |
| `record_debt` | mutación | ✅ Siempre |
| `update_goal_progress` | mutación | ❌ Baja criticidad |
| `delete_transaction` | mutación destructiva | ✅ Siempre + re-confirm |

### Orquestación del tool loop en Go

```go
// api/internal/infrastructure/anthropic/client.go
func (c *Client) RunToolLoop(ctx context.Context, msgs []Message, tools []Tool) (*Response, error) {
    for {
        resp, err := c.CreateMessage(ctx, msgs, tools)
        if err != nil { return nil, err }

        if resp.StopReason == "end_turn" {
            return resp, nil
        }

        if resp.StopReason == "tool_use" {
            for _, block := range resp.Content {
                if block.Type != "tool_use" { continue }

                if needsConfirmation(block.Name) {
                    return &Response{PendingConfirmation: block}, nil
                }

                result := c.executeTool(ctx, block.Name, block.Input)
                msgs = append(msgs, assistantMsg(resp), toolResultMsg(block.ID, result))
            }
            continue
        }
        return resp, nil
    }
}
```

**Confirmaciones**: Las mutaciones marcadas con confirmación retornan al cliente con `pending_confirmation: true`. El cliente muestra el preview, el usuario acepta/rechaza, y el siguiente request continúa el loop.

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | Domain entities, value objects | `go test` puro, zero deps |
| Unit | Use cases | Mocks de repository interfaces |
| Integration | Handlers + middleware | `httptest` con DB test |
| Integration | Supabase RLS | SQL tests: verificar que user A no lee data de user B |
| E2E | Chat → Transaction flow | TUI test client contra API local |

## Migration / Rollout

Greenfield — no hay migración de datos. Migrations via `goose up` en `api/migrations/`. Seed de `currencies` en primera migración.

## Open Questions

- [ ] ¿Qué API externa para exchange rates? (Open Exchange Rates vs ExchangeRate-API vs Frankfurter — gratis)
- [ ] ¿WebSocket para streaming de respuestas AI o SSE?
- [ ] ¿El TUI almacena JWT en keychain del OS o en archivo config?
