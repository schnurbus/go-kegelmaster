DATABASE_URL ?= postgres://kegelmaster:kegelmaster@localhost:5432/kegelmaster?sslmode=disable

# E2E Test Database
E2E_DB_PORT=5433
E2E_DB_USER=kegelmaster
E2E_DB_PASSWORD=e2e_test_password
E2E_DB_NAME=e2e_testdb
E2E_DB_URL="postgres://$(E2E_DB_USER):$(E2E_DB_PASSWORD)@localhost:$(E2E_DB_PORT)/$(E2E_DB_NAME)?sslmode=disable"
E2E_CONTAINER_NAME=postgres_e2e

# Lokale Test-Instanz (make dev). Eigener Postgres-Container mit eigenem Volume,
# damit weder die docker-compose-DB (:5432) noch die E2E-DB (:5433) berührt werden.
DEV_CONTAINER_NAME ?= kegelmaster-dev-postgres
DEV_VOLUME         ?= kegelmaster-dev-pgdata
DEV_DB_PORT        ?= 5434
DEV_BACKEND_PORT   ?= 8080
DEV_FRONTEND_PORT  ?= 5173
DEV_DB_URL          = postgres://kegelmaster:kegelmaster@localhost:$(DEV_DB_PORT)/kegelmaster?sslmode=disable
# FRESH=1: Datenbank vor dem Start verwerfen. SEED=1: Beispiel-Daten laden.
FRESH ?= 0
SEED  ?= 0

.PHONY: dev dev-fresh dev-db-up dev-db-down dev-db-reset dev-seed dev-psql seed-build
.PHONY: backend-run backend-test backend-build backend-build-embed build import-build migrate-old-build frontend-dev frontend-build compose-up compose-down fmt migrate-up migrate-down sqlc-generate test-e2e

backend-run:
	@cd backend && go run ./cmd/api

backend-test:
	@cd backend && go test ./...

backend-build:
	@cd backend && go build ./cmd/api

# Build CLI import tool (CSV → Spieltage).
import-build:
	@mkdir -p bin && cd backend && go build -o ../bin/import ./cmd/import

# Build CLI seed (Beispiel-Daten für lokale Test-Instanzen).
seed-build:
	@mkdir -p bin && cd backend && go build -o ../bin/seed ./cmd/seed

# Build CLI migrate-old (alte DB → neue App-DB).
migrate-old-build:
	@mkdir -p bin && cd backend && go build -o ../bin/migrate-old ./cmd/migrate-old

# Build frontend, copy into backend for embed, then build single binary with embedded frontend.
build: frontend-build
	@mkdir -p backend/internal/server/web && cp -r frontend/dist/* backend/internal/server/web/
	@mkdir -p bin && cd backend && go build -tags embed -o ../bin/kegelmaster ./cmd/api

# Build single binary only (assumes frontend/dist already exists from frontend-build).
backend-build-embed:
	@mkdir -p backend/internal/server/web && cp -r frontend/dist/* backend/internal/server/web/
	@mkdir -p bin && cd backend && go build -tags embed -o ../bin/kegelmaster ./cmd/api

frontend-dev:
	@cd frontend && npm run dev

frontend-build:
	@cd frontend && npm run build

compose-up:
	docker compose up --build

compose-down:
	docker compose down -v

fmt:
	@cd backend && gofmt -w ./cmd ./internal

migrate-up:
	migrate -path backend/migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path backend/migrations -database "$(DATABASE_URL)" down

sqlc-generate:
	@cd backend && sqlc generate

test-e2e:
	@echo "Starte dedizierte E2E-Datenbank..."
	docker rm -f $(E2E_CONTAINER_NAME) || true
	
	docker run --name $(E2E_CONTAINER_NAME) \
		-e POSTGRES_USER=$(E2E_DB_USER) \
		-e POSTGRES_PASSWORD=$(E2E_DB_PASSWORD) \
		-e POSTGRES_DB=$(E2E_DB_NAME) \
		-p $(E2E_DB_PORT):5432 \
		-d postgres:16-alpine
	
	@until docker exec $(E2E_CONTAINER_NAME) pg_isready; do sleep 1; done
	
	@echo "Führe Migrationen auf E2E-DB aus..."
	migrate -path backend/migrations -database $(E2E_DB_URL) up
	
	@echo "Starte Tests..."
	cd frontend && npx playwright test --config playwright.local.config.ts $(args); \
	EXIT_CODE=$$?; \
	docker rm -f $(E2E_CONTAINER_NAME); \
	exit $$EXIT_CODE

# ---------------------------------------------------------------------------
# Lokale Test-Instanz aus dem aktuellen Branch
#   make dev                 Postgres starten, Backend (:8080) + Vite (:5173) starten
#   make dev SEED=1          zusätzlich Beispiel-Daten laden (übersprungen, wenn vorhanden)
#   make dev FRESH=1         mit leerer Datenbank starten
#   make dev-fresh           = make dev FRESH=1 SEED=1
# Login nach SEED: demo@kegelmaster.local / kegelmaster (Owner),
#                  mitglied@kegelmaster.local / kegelmaster (nur Lesen)
# Beenden mit Ctrl-C; die Datenbank bleibt erhalten (make dev-db-down stoppt sie).
# ---------------------------------------------------------------------------

dev:
ifeq ($(FRESH),1)
	@$(MAKE) --no-print-directory dev-db-reset
else
	@$(MAKE) --no-print-directory dev-db-up
endif
ifeq ($(SEED),1)
	@$(MAKE) --no-print-directory dev-seed
endif
	@test -d frontend/node_modules || (echo "Installiere Frontend-Abhängigkeiten..." && cd frontend && npm ci)
	@echo "Baue Backend (Branch $$(git rev-parse --abbrev-ref HEAD))..."
	@mkdir -p bin && cd backend && go build -o ../bin/kegelmaster-dev ./cmd/api
	@echo ""
	@echo "  Frontend: http://localhost:$(DEV_FRONTEND_PORT)"
	@echo "  Backend:  http://localhost:$(DEV_BACKEND_PORT)  (API-Doku: /api/docs)"
	@echo "  DB:       $(DEV_DB_URL)"
	@echo ""
	@trap 'kill $$BACKEND_PID $$FRONTEND_PID 2>/dev/null; wait' INT TERM EXIT; \
	APP_ENV=development AUTO_MIGRATE=true \
	DATABASE_URL="$(DEV_DB_URL)" BACKEND_PORT=$(DEV_BACKEND_PORT) \
	CORS_ALLOW_ORIGINS=http://localhost:$(DEV_FRONTEND_PORT) \
	PUBLIC_BASE_URL=http://localhost:$(DEV_FRONTEND_PORT) \
		./bin/kegelmaster-dev & BACKEND_PID=$$!; \
	cd frontend; \
	VITE_BACKEND_URL=http://localhost:$(DEV_BACKEND_PORT) \
		./node_modules/.bin/vite --port $(DEV_FRONTEND_PORT) --strictPort & FRONTEND_PID=$$!; \
	while kill -0 $$BACKEND_PID 2>/dev/null && kill -0 $$FRONTEND_PID 2>/dev/null; do sleep 1; done

dev-fresh:
	@$(MAKE) --no-print-directory dev FRESH=1 SEED=1

# Startet (oder legt an) den Dev-Postgres und wartet, bis er Verbindungen annimmt.
dev-db-up:
	@if [ -z "$$(docker ps -q -f name=^$(DEV_CONTAINER_NAME)$$)" ]; then \
		if [ -n "$$(docker ps -aq -f name=^$(DEV_CONTAINER_NAME)$$)" ]; then \
			docker start $(DEV_CONTAINER_NAME) >/dev/null; \
		else \
			docker run -d --name $(DEV_CONTAINER_NAME) \
				-e POSTGRES_USER=kegelmaster -e POSTGRES_PASSWORD=kegelmaster -e POSTGRES_DB=kegelmaster \
				-p $(DEV_DB_PORT):5432 -v $(DEV_VOLUME):/var/lib/postgresql/data \
				postgres:16-alpine >/dev/null; \
		fi; \
	fi
	@echo "Warte auf Postgres ($(DEV_CONTAINER_NAME), Port $(DEV_DB_PORT))..."
	@until docker exec $(DEV_CONTAINER_NAME) pg_isready -h 127.0.0.1 -U kegelmaster -q; do sleep 1; done

# Stoppt den Dev-Postgres, Daten bleiben im Volume erhalten.
dev-db-down:
	@docker stop $(DEV_CONTAINER_NAME) >/dev/null 2>&1 || true

# Verwirft Container und Volume und startet mit einer leeren Datenbank.
dev-db-reset:
	@echo "Verwerfe Dev-Datenbank..."
	@docker rm -f $(DEV_CONTAINER_NAME) >/dev/null 2>&1 || true
	@docker volume rm $(DEV_VOLUME) >/dev/null 2>&1 || true
	@$(MAKE) --no-print-directory dev-db-up

# Führt Migrationen aus und lädt Beispiel-Daten (übersprungen, wenn schon vorhanden).
dev-seed: dev-db-up
	@cd backend && go run ./cmd/seed -database-url "$(DEV_DB_URL)"

dev-psql:
	@docker exec -it $(DEV_CONTAINER_NAME) psql -U kegelmaster kegelmaster