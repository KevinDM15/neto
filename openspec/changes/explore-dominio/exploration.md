# Exploración: Dominio Finanzas Personales con IA — Neto

> Fecha: 2026-04-27  
> Agente: sdd-explore  
> Estado: completed

---

## 1. Mapa del Dominio

### Entidades Clave

```
USUARIO
├── Perfil (nombre, moneda base, zona horaria)
├── Configuración IA (límites de gasto, alertas, goals)
└── Suscripción (si aplica modelo SaaS)

CUENTA
├── id, nombre, tipo (efectivo | débito | crédito | inversión | cripto)
├── moneda, saldo_actual
├── institución (banco, fintech, etc.)
└── activa: boolean

TRANSACCIÓN  ← entidad central
├── id, cuenta_id, fecha, monto, moneda
├── tipo (ingreso | egreso | transferencia)
├── categoría_id, subcategoría_id
├── descripción, notas
├── recurrente: boolean, patrón_recurrencia
├── etiquetas: string[]
├── adjuntos (foto de ticket, PDF)
└── origen (manual | importación | IA-inferido)

CATEGORÍA
├── id, nombre, ícono, color
├── tipo (ingreso | egreso)
├── padre_id (para subcategorías)
└── sistema: boolean (predefinida vs custom)

PRESUPUESTO
├── id, categoría_id, periodo (mensual | semanal | anual)
├── monto_limite, monto_gastado
├── alerta_porcentaje (ej: alertar al 80%)
└── activo: boolean

DEUDA
├── id, nombre, acreedor/deudor
├── tipo (me_deben | debo_yo)
├── monto_original, monto_pendiente
├── tasa_interés (opcional), fecha_vencimiento
├── cuotas (si es en cuotas)
└── estado (activa | parcial | cancelada)

META_FINANCIERA
├── id, nombre, descripción
├── monto_objetivo, monto_acumulado
├── fecha_objetivo
├── cuenta_destino_id (opcional, para ahorro dedicado)
└── estado (en_curso | alcanzada | cancelada)

INFORME / SNAPSHOT
├── periodo, generado_en
├── ingresos_totales, egresos_totales, balance
├── breakdown_por_categoría
└── insights_ia: string[] (generado por IA)

AGENTE_IA (entidad lógica)
├── contexto_usuario (historial financiero resumido)
├── reglas_personalizadas (alertas, limits)
├── historial_conversaciones
└── acciones_tomadas (categorizó, alertó, sugirió)
```

### Relaciones

```
Usuario 1──n Cuenta
Cuenta  1──n Transacción
Transacción n──1 Categoría
Categoría   1──n Categoría (self-ref, subcategorías)
Usuario 1──n Presupuesto
Presupuesto n──1 Categoría
Usuario 1──n Deuda
Usuario 1──n Meta_Financiera
Meta_Financiera n──1 Cuenta (opcional)
Transacción n──n Etiqueta
```

---

## 2. Flows Principales

### Flow 1: Registro de Gasto (más frecuente)
```
Usuario → CLI/Web/Mobile
  → Ingresa: "gasté $1500 en almuerzo hoy"
  → IA infiere: categoría=Comida, subcategoría=Restaurantes, fecha=hoy
  → Confirma o corrige usuario
  → Persiste Transacción
  → Actualiza saldo Cuenta
  → Evalúa Presupuesto → si supera umbral, alerta
```

### Flow 2: Consulta con IA
```
Usuario → "¿cuánto gasté en delivery este mes?"
  → IA consulta transacciones filtradas
  → Retorna resumen + gráfico (si es web/mobile)
  → Opcionalmente sugiere acción ("vas a pasarte del presupuesto")
```

### Flow 3: Gestión de Deudas
```
Usuario → "le presté $5000 a Juan"
  → IA crea Deuda (tipo=me_deben, deudor=Juan, monto=5000)
  → Registra como transacción saliente (categoría=Préstamo)
  → Programa recordatorio si hay fecha acordada
```

### Flow 4: Cierre de Mes / Informe
```
Trigger: automático o manual al fin de mes
  → Agrega todas las transacciones del período
  → Calcula: balance, gastos por categoría, vs presupuestos
  → IA genera insights ("Gastaste 40% más en entretenimiento que el mes pasado")
  → Genera Informe persistido
  → Notifica al usuario (push/email/CLI output)
```

### Flow 5: Meta de Ahorro
```
Usuario crea Meta → "quiero juntar $200,000 para vacaciones en 6 meses"
  → IA calcula: "necesitás ahorrar $33,333/mes"
  → Cada mes, IA verifica progreso vs meta
  → Alerta si el ritmo no alcanza
```

### Flow 6: Importación de Extracto
```
Usuario sube PDF/CSV de banco
  → Parser extrae transacciones
  → IA categoriza automáticamente
  → Usuario revisa y confirma batch
  → Deduplicación con transacciones ya existentes
```

---

## 3. Opciones Técnicas con Tradeoffs

### 3.1 Base de Datos

| Opción | Pros | Contras | Fit Neto |
|--------|------|---------|----------|
| **Supabase** | Auth built-in, Realtime, Storage (adjuntos), edge functions, UI dashboard para admin | Vendor lock-in, costo a escala, datos financieros en la nube (privacidad), latencia si offline | ★★★★☆ para MVP rápido |
| **Postgres local** | Control total, sin costo, datos on-premise | Dev-ops complejo, no resuelve sync multi-device, auth hay que construirla | ★★☆☆☆ |
| **SQLite (local)** | Funciona offline, cero latencia, ideal para CLI y mobile local-first | No multi-device sin sync layer adicional (CR-SQLite/Turso) | ★★★☆☆ para offline-first |
| **Turso (libSQL)** | SQLite-compatible + sync edge, local-first con réplicas, bajo costo | Ecosistema más chico, menos maduro | ★★★★☆ para local-first serio |
| **PocketBase** | Self-hosted Supabase-like, SQLite bajo, auth, realtime | Auto-host = infra propia, comunidad más pequeña | ★★★☆☆ si quieren self-host |

**Recomendación**: **Supabase para MVP** — resuelve auth, storage, realtime y sync en una sola herramienta. Si la privacidad es requisito hard, evaluar **Turso** (local-first + sync edge sin datos en terceros).

---

### 3.2 CLI Framework

| Opción | Lenguaje | Pros | Contras | Fit |
|--------|----------|------|---------|-----|
| **Ink + Commander** | Node/TS | Comparte código con web (TypeScript unificado), React para TUI | Perf inferior, startup más lento | ★★★★☆ si el stack es TS |
| **Cobra + BubbleTea** | Go | Binario único, perf excelente, TUI espectacular, distribución fácil | Lenguaje separado si el resto es TS | ★★★★★ para CLI standalone |
| **Typer + Rich** | Python | Productividad alta, libs IA en Python (LangChain nativo) | Distribución compleja, más lento | ★★★☆☆ si hay mucho Python IA |
| **Clack + Prompts** | Node/TS | Simple, bonito, bajo overhead | No TUI compleja | ★★★☆☆ para CLI simple |

**Recomendación**: **Ink/Commander (TypeScript)** si el monorepo es TS. **Cobra+BubbleTea (Go)** si quieren una CLI de primera clase distribuible como binario.

---

### 3.3 Web Framework

| Opción | Pros | Contras | Fit |
|--------|------|---------|-----|
| **Next.js 15** | Full-stack en uno, App Router, Server Actions, ecosistema enorme, fácil deploy Vercel | Bundle size, over-engineering para app simple, opinionated | ★★★★★ |
| **SvelteKit** | Menos JS, más rápido, sintaxis más simple, stores reactivos | Ecosistema más chico, menos devs disponibles | ★★★★☆ |
| **Astro + React islands** | Cero JS por defecto, ideal para contenido estático | No ideal para app altamente interactiva (dashboard financiero vivo) | ★★☆☆☆ para este caso |
| **Remix** | Full-stack, loader/action pattern limpio, UX de primera | Menos ecosystem que Next, curva si no conocen | ★★★★☆ |

**Recomendación**: **Next.js 15 (App Router)** — máximo ecosistema, Server Actions para operaciones financieras, deploy trivial en Vercel, y compatibilidad natural con Supabase.

---

### 3.4 Mobile

| Opción | Pros | Contras | Fit |
|--------|------|---------|-----|
| **Expo (React Native)** | Código compartido con web (si es React), OTA updates, push notifications, cámara para tickets | Perf inferior a nativo, bundle grande | ★★★★★ si web es React |
| **PWA** | Un solo codebase web+mobile, cero app store | Sin push notifications reales en iOS, sin acceso a cámara nativo, experiencia limitada | ★★★☆☆ para MVP rápido |
| **Flutter** | UI nativa excelente, un codebase iOS+Android | Lenguaje Dart separado, no comparte con web | ★★★☆☆ si priorizan mobile |
| **Capacitor (Ionic)** | Wrapper web-to-native, reutiliza web app | UX inferior a nativo, limitaciones de acceso nativo | ★★★☆☆ |

**Recomendación**: **Expo (React Native)** para compartir lógica con Next.js via monorepo. O **PWA** para MVP si quieren velocidad máxima (aceptando limitaciones iOS).

---

### 3.5 Inteligencia Artificial

| Opción | Pros | Contras | Fit |
|--------|------|---------|-----|
| **OpenAI API (GPT-4o)** | Mejor calidad general, function calling maduro, visión (para tickets), streaming | Costo por token, datos financieros van a OpenAI, dependencia | ★★★★★ calidad |
| **Anthropic (Claude)** | Excelente razonamiento, mejor en tareas largas, tool use maduro | Similar costo, misma preocupación de privacidad | ★★★★★ calidad alt |
| **Ollama (local)** | Privacidad total, sin costo, offline | Requiere hardware del usuario, calidad inferior en razonamiento financiero complejo | ★★★☆☆ privacidad |
| **Groq API** | Ultra-rápido, bajo costo, modelos open | Modelos menos capaces, no siempre disponible | ★★★☆☆ velocidad |
| **LangChain/LangGraph** | Orquestación de agentes, memory, tools | Complejidad alta, overhead, cambia mucho la API | ★★★☆☆ si necesitan agente complejo |
| **Vercel AI SDK** | Integración natural con Next.js, streaming built-in, multi-provider | Solo viable si el web stack es Next/Vercel | ★★★★☆ para Next.js |

**Recomendación**: **Anthropic Claude** como proveedor principal + **Vercel AI SDK** si van con Next.js (abstrae provider, maneja streaming, tool use). Fallback a Groq para operaciones simples por costo.

---

## 4. Riesgos Identificados

### Riesgo 1 (ALTO): Sincronización multi-plataforma
**Problema**: El usuario registra un gasto en CLI, lo ve en web, y lo edita en mobile. Conflictos de estado, offline edits, y race conditions son complejos.  
**Mitigación**: Supabase Realtime resuelve esto parcialmente. Si van local-first, necesitan CRDT o estrategia de merge. Definir esta decisión ANTES de codear.

### Riesgo 2 (ALTO): Privacidad de datos financieros
**Problema**: Los datos financieros son extremadamente sensibles. Enviarlos a OpenAI/Anthropic implica que el proveedor los ve. Regulaciones (GDPR-like) pueden aplicar.  
**Mitigación**: Anonimizar/resumir antes de enviar a IA (no mandar saldos exactos, sino patrones). O usar modelos locales (Ollama) para datos sensibles. Documentar política de privacidad desde el inicio.

### Riesgo 3 (MEDIO): Scope creep por IA
**Problema**: "El agente IA" puede significar mil cosas. Sin límites claros, se termina construyendo un sistema de agentes autónomos cuando lo que se necesita es categorización automática + consultas en lenguaje natural.  
**Mitigación**: Definir en la propuesta: ¿qué hace el agente? ¿es conversacional? ¿toma acciones autónomas? ¿solo categoriza? Limitar MVP de IA a: (1) categorización automática, (2) consultas en lenguaje natural, (3) alertas de presupuesto.

### Riesgo 4 (MEDIO): Arquitectura multi-plataforma prematura
**Problema**: Intentar construir CLI + Web + Mobile desde el día 1 fragmenta el equipo y demora el valor.  
**Mitigación**: Definir una plataforma primaria para el MVP (recomendación: Web primero), y compartir dominio/lógica de negocio via monorepo desde el inicio.

### Riesgo 5 (BAJO): Importación de extractos bancarios
**Problema**: Cada banco tiene formato de PDF/CSV diferente. Parsear PDFs es doloroso y los bancos cambian el formato.  
**Mitigación**: Para MVP, solo soportar entrada manual + IA conversacional. Importación como feature de v2 con parsers específicos por banco.

---

## 5. Arquitectura Base Propuesta

### Estructura: Monorepo con Shared Domain

```
neto/
├── packages/
│   ├── domain/          # Entidades, casos de uso, puertos (framework-agnóstico, TypeScript)
│   │   ├── entities/    # Account, Transaction, Budget, Debt, Goal
│   │   ├── usecases/    # RegisterTransaction, GetMonthlyReport, AlertBudget
│   │   └── ports/       # ITransactionRepo, INotifier, IAIService
│   ├── infra/           # Implementaciones concretas
│   │   ├── supabase/    # SupabaseTransactionRepo, SupabaseAuthService
│   │   ├── ai/          # AnthropicService, OpenAIService
│   │   └── parsers/     # BankStatementParser (v2)
│   └── shared/          # Types, utils, formatters, validaciones
├── apps/
│   ├── web/             # Next.js 15 App Router
│   ├── cli/             # Ink + Commander (TypeScript)
│   └── mobile/          # Expo (React Native)
└── openspec/            # Specs SDD
```

### Principio Guía: Clean/Hexagonal Architecture
- `packages/domain` no depende de ningún framework ni proveedor
- Las apps (`web`, `cli`, `mobile`) son delivery mechanisms
- La IA es un puerto más (`IAIService`), intercambiable entre proveedores
- Supabase es una implementación del puerto `IRepository`, reemplazable

### Flujo de datos
```
App (web/cli/mobile)
  → UseCase (domain)
    → Port (abstracción)
      → Infra (Supabase / AI provider / etc.)
```

---

## 6. Preguntas Abiertas (para propuesta)

1. **¿Cuál es la plataforma primaria del MVP?** — ¿Empezamos por web, CLI o mobile? Esto define el stack inicial.
2. **¿Los datos financieros pueden ir a la nube?** — Supabase (Postgres gestionado) vs local-first (Turso/SQLite). Esto es una decisión de privacidad.
3. **¿El agente IA es conversacional o solo asistente?** — ¿Hay un chat? ¿O solo categoriza y alerta?
4. **¿Hay modelo de suscripción real o es personal/internal tool?** — Si es SaaS real, cambia todo el modelo de auth y multi-tenancy.
5. **¿Multi-moneda desde el inicio?** — ARS + USD + crypto? O solo ARS para MVP.
6. **¿El CLI es para el mismo usuario que la web o es un tool de power-user separado?**

---

## Recomendación Final

**Stack MVP recomendado:**
- **Web**: Next.js 15 (App Router) + Supabase + Vercel AI SDK + Anthropic Claude
- **CLI**: Ink + Commander (TypeScript, mismo monorepo)
- **Mobile**: Expo — diferido a v2, o PWA como puente
- **Arquitectura**: Monorepo (pnpm workspaces) con Clean Architecture en `packages/domain`
- **IA MVP scope**: categorización automática + consultas en lenguaje natural + alertas de presupuesto

**Plataforma primaria sugerida**: Web (más rápido de iterar, mejor para mostrar dashboards, y comparte lógica con CLI/mobile después).

---

## Ready for Proposal
**Sí** — con clarificación de las 6 preguntas abiertas (especialmente plataforma primaria y política de privacidad de datos). La propuesta puede avanzar con supuestos documentados si el usuario quiere moverse rápido.
