# Migraciones — Neto API

Usando [goose](https://github.com/pressly/goose) con SQL puro.

## Setup local

1. Instalar goose:
   ```bash
   go install github.com/pressly/goose/v3/cmd/goose@latest
   ```

2. Copiar y completar el archivo de entorno:
   ```bash
   cp migrations/goose.env.example migrations/goose.env
   # Editar goose.env con tu GOOSE_DBSTRING de Supabase
   ```

   Obtener la connection string en: Supabase → Settings → Database → Connection string → URI

3. El archivo `goose.env` está en `.gitignore` — nunca lo commitees.

## Comandos

Desde el directorio `api/`:

```bash
# Aplicar todas las migraciones pendientes
goose -dir migrations postgres "$GOOSE_DBSTRING" up

# Revertir la última migración
goose -dir migrations postgres "$GOOSE_DBSTRING" down

# Ver estado de migraciones
goose -dir migrations postgres "$GOOSE_DBSTRING" status

# Crear nueva migración
goose -dir migrations create nombre_descriptivo sql
```

O via Makefile desde la raíz del proyecto:

```bash
make migrate-up
make migrate-down
make migrate-status
```

Para usar los targets del Makefile, exportar la variable:
```bash
export GOOSE_DBSTRING="postgres://..."
make migrate-up
```

## Convención de nombres

```
YYYYMMDDHHMMSS_descripcion_en_snake_case.sql
```

Ejemplo: `20260428000001_create_extensions.sql`

## Orden de aplicación

El orden importa por dependencias de foreign keys:

1. `extensions` — uuid-ossp
2. `currencies` — tabla base sin FKs
3. `exchange_rates` → currencies
4. `accounts` → currencies
5. `categories` → categories (self-ref)
6. `transactions` → accounts, categories, currencies
7. `budgets` → categories, currencies
8. `debts` → currencies
9. `goals` → currencies
10. `ai_conversations` — sin FKs externas
11. `ai_messages` → ai_conversations
12. `idempotency_keys` — sin FKs externas
13. `enable_rls` — políticas RLS (requiere todas las tablas)
14. `seed_currencies` — datos iniciales

## En CI

Las migraciones se corren automáticamente en el pipeline usando `GOOSE_DBSTRING` desde secrets:

```yaml
- name: Run migrations
  run: |
    cd api
    goose -dir migrations postgres "$GOOSE_DBSTRING" up
  env:
    GOOSE_DBSTRING: ${{ secrets.SUPABASE_DB_URL }}
```

## Convenciones

- **SQL puro** — no Go migrations
- **Idempotente**: usar `IF NOT EXISTS`, `IF EXISTS`, `ON CONFLICT DO NOTHING`
- **Down siempre presente**: toda migración debe tener su sección `-- +goose Down`
- **Un concern por archivo**: no mezclar tablas no relacionadas en una migración
