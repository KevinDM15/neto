.PHONY: api tui web dev test lint migrate-up migrate-down migrate-status

# Correr la API en modo desarrollo
api:
	go run ./... -C api

# Correr el TUI
tui:
	go run ./... -C tui

# Correr el cliente web (Vite dev server)
web:
	cd web && pnpm dev

# Levantar todo en paralelo (requiere make ≥ 4.3 con --jobs)
dev:
	$(MAKE) -j3 api tui web

# Correr todos los tests Go
test:
	cd api && go test ./... -v
	cd tui && go test ./... -v

# Lint Go (requiere golangci-lint) y ESLint para el web
lint:
	cd api && golangci-lint run ./...
	cd tui && golangci-lint run ./...
	cd web && pnpm lint

# Migraciones (requiere GOOSE_DBSTRING exportada o en api/.env)
migrate-up:
	cd api && goose -dir migrations postgres "$$GOOSE_DBSTRING" up

migrate-down:
	cd api && goose -dir migrations postgres "$$GOOSE_DBSTRING" down

migrate-status:
	cd api && goose -dir migrations postgres "$$GOOSE_DBSTRING" status
