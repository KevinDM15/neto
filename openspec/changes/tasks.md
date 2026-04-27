# Tasks: Proyecto Neto (MVP)

## Phase 1: Setup

- [x] 1.1 `NETO-001` Crear estructura monorepo `api/`, `tui/`, `web/`, `shared/` y entrypoints mínimos. *(S)*
- [x] 1.2 `NETO-002` Inicializar módulos de `api/` y `tui/` más `web/package.json` y `shared/openapi.yaml` placeholder. *(S)*
- [x] 1.3 `NETO-003` Agregar configuración base de entorno, `.env.example` y `.gitignore` por paquete. *(S)*

## Phase 2: Core Domain

- [ ] 2.1 `NETO-004` Crear value objects de dinero, moneda y tipos financieros en `api/internal/domain/valueobject/`. *(M)*
- [ ] 2.2 `NETO-005` Crear entidades `Account`, `Category`, `Transaction` y reglas de balance/categorización. *(M)*
- [ ] 2.3 `NETO-006` Crear entidades `Budget`, `Debt`, `Goal`, `AIConversation`, `AIMessage`, `IdempotencyKey`. *(M)*
- [ ] 2.4 `NETO-007` Definir interfaces de repositorio y servicios externos del dominio. *(S)*
- [ ] 2.5 `NETO-008` Escribir tests unitarios para balance, overspend y progreso de deuda/meta. *(M)*

## Phase 3: DB + Supabase

- [ ] 3.1 `NETO-009` Crear migración inicial con 11 tablas, constraints e índices definidos en diseño. *(L)*
- [ ] 3.2 `NETO-010` Agregar políticas RLS para tablas con `user_id` y lectura autenticada de `exchange_rates`. *(M)*
- [ ] 3.3 `NETO-011` Crear seed SQL de monedas base LatAm + USD/EUR. *(S)*
- [ ] 3.4 `NETO-012` Configurar ejecución de migraciones `goose` desde `api/`. *(S)*
- [ ] 3.5 `NETO-013` Escribir pruebas SQL de aislamiento RLS entre usuarios y visibilidad de categorías default. *(M)*

## Phase 4: API

- [ ] 4.1 `NETO-014` Crear bootstrap HTTP en `api/cmd/server/main.go` con router Chi y healthcheck. *(S)*
- [ ] 4.2 `NETO-015` Implementar carga de config y cliente Supabase/JWKS en infraestructura. *(M)*
- [ ] 4.3 `NETO-016` Implementar middleware JWT con contexto de usuario autenticado. *(M)*
- [ ] 4.4 `NETO-017` Implementar middleware de idempotencia respaldado por `idempotency_keys`. *(M)*
- [ ] 4.5 `NETO-018` Implementar repositorios Supabase para cuentas, transacciones, categorías, presupuestos, deudas y metas. *(L)*
- [ ] 4.6 `NETO-019` Implementar use cases y handlers REST para CRUD base y resumen mensual. *(L)*
- [ ] 4.7 `NETO-020` Publicar contrato inicial en `shared/openapi.yaml` alineado a endpoints reales. *(M)*
- [ ] 4.8 `NETO-021` Escribir tests de integración de handlers, JWT e idempotencia. *(M)*

## Phase 5: AI Agent

- [ ] 5.1 `NETO-022` Implementar cliente Anthropic sin streaming y límite de 5 iteraciones. *(M)*
- [ ] 5.2 `NETO-023` Definir catálogo de tools financieras y esquema de inputs/outputs. *(M)*
- [ ] 5.3 `NETO-024` Implementar tool loop server-side con confirmaciones pendientes para mutaciones. *(L)*
- [ ] 5.4 `NETO-025` Persistir conversaciones y mensajes AI en Supabase. *(M)*
- [ ] 5.5 `NETO-026` Exponer endpoint de chat integrado con auth, tools y errores de negocio. *(M)*
- [ ] 5.6 `NETO-027` Escribir tests de registro de gasto, consulta y confirmación destructiva. *(M)*

## Phase 6: TUI

- [ ] 6.1 `NETO-028` Crear bootstrap Bubbletea con modelo raíz, viewport e input principal. *(M)*
- [ ] 6.2 `NETO-029` Implementar manejo de config local en `~/.config/neto/` con permisos `0600`. *(M)*
- [ ] 6.3 `NETO-030` Implementar flujo de login inicial y persistencia local de JWT. *(M)*
- [ ] 6.4 `NETO-031` Implementar cliente HTTP del TUI para auth, chat y endpoints de lectura. *(M)*
- [ ] 6.5 `NETO-032` Implementar vista de chat con historial, spinner y estados de error/éxito. *(M)*
- [ ] 6.6 `NETO-033` Implementar UI de confirmación de acciones AI y navegación básica ayuda/configuración. *(M)*
- [ ] 6.7 `NETO-034` Escribir tests TUI para primer uso, chat exitoso y confirmación de mutación. *(M)*

## Phase 7: Web

- [ ] 7.1 `NETO-035` Inicializar app Vite + React cliente-only con estructura `src/`. *(S)*
- [ ] 7.2 `NETO-036` Implementar cliente API tipado desde `shared/openapi.yaml` y manejo de sesión JWT. *(M)*
- [ ] 7.3 `NETO-037` Implementar layout base, login y vista principal de chat web. *(M)*
- [ ] 7.4 `NETO-038` Implementar componentes de confirmación y feedback financiero enriquecido. *(M)*
- [ ] 7.5 `NETO-039` Escribir tests UI para login, chat y render de respuestas/resúmenes. *(M)*
