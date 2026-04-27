# Proposal: Proyecto Neto (MVP)

## Intent
Neto es una plataforma open-source de finanzas personales, self-hosteable y multi-usuario. Resuelve la fricción en el registro y análisis de gastos mediante una interfaz conversacional impulsada por IA (Claude), permitiendo la gestión eficiente del patrimonio en múltiples monedas y plataformas (Web y CLI).

## Scope

### In Scope
- Autenticación multi-usuario (Supabase RLS).
- Monorepo pnpm con Clean Architecture.
- Plataformas: Web (Next.js 15) y CLI (Ink).
- Interfaz conversacional con Anthropic Claude (registro y consultas en lenguaje natural usando Tool Use).
- Gestión de Cuentas, Transacciones y Categorías en árbol.
- Soporte multi-moneda (LatAm, USD, EUR) con seed data base COP y actualización de exchange rates vía API externa.
- Presupuestos, deudas y metas.
- Flujo de cierre de mes con insights.

### Out of Scope
- Aplicación Mobile nativa o PWA (queda explícitamente para v2).
- Sincronización automática bancaria vía APIs de bancos (Plaid/Prometeo).
- Otros proveedores de IA (solo Anthropic en MVP).

## Capabilities

### New Capabilities
- `core-domain`: Entidades core y casos de uso en Clean Architecture (agnósticos).
- `auth-tenancy`: Autenticación y aislamiento multi-tenant con Supabase RLS.
- `ai-chat-interface`: Interfaz conversacional y orquestación de Tool Use con Claude.
- `currency-manager`: Seed y actualización de tipos de cambio (API externa).
- `web-client`: App Next.js 15.
- `cli-client`: App de terminal interactiva con Ink.

### Modified Capabilities
- Ninguna (proyecto greenfield).

## Approach
Se utilizará un **monorepo pnpm**, estructurando el dominio y los casos de uso (Clean Architecture) como paquetes internos consumibles por los clientes (Web Next.js y CLI Ink). La DB será Supabase, delegando el aislamiento robusto de usuarios a Row Level Security (RLS). El registro de transacciones priorizará la naturalidad del lenguaje: parseando la intención del usuario con Claude y ejecutando Tools para mutar la base de datos estructurada.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `packages/core` | New | Dominio puro y casos de uso (agnósticos) |
| `packages/infrastructure` | New | Repositorios, clientes Supabase y Claude |
| `apps/web` | New | Frontend Next.js 15 |
| `apps/cli` | New | Interfaz de consola con TypeScript + Ink |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Falsos positivos IA al mutar DB | Med | Claude Tool Use estricto + UI para confirmación de transacciones antes de commit |
| Límites de rate API Exchange | Low | Cronjob diario para Exchange Rates (Supabase Edge Functions), no por request |
| Leak de datos entre tenants | Low | Tests rigurosos y exclusivos sobre las políticas RLS de Supabase |

## Rollback Plan
Al ser el MVP inicial, en desarrollo se descarta el entorno local. En producción post-lanzamiento, el rollback consta de revertir el PR en la rama `main` y restaurar los backups diarios automáticos (PITR) de Supabase.

## Success Criteria
- [ ] La CLI y la Web logran reutilizar la lógica de negocio consumiendo un `packages/core` compartido.
- [ ] Un usuario se registra, crea su cuenta y su data queda aislada (comprobado vía RLS).
- [ ] Claude procesa exitosamente un prompt ("Gasté 50k COP en luz") ejecutando el tool correspondiente de creación de transacción.
- [ ] El sistema maneja múltiples monedas con el tipo de cambio correcto aplicado a un reporte mensual.