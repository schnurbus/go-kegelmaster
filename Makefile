DATABASE_URL ?= postgres://kegelmaster:kegelmaster@localhost:5432/kegelmaster?sslmode=disable

# E2E Test Database
E2E_DB_PORT=5433
E2E_DB_USER=kegelmaster
E2E_DB_PASSWORD=e2e_test_password
E2E_DB_NAME=e2e_testdb
E2E_CONTAINER_NAME=postgres_e2e

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
	# 1. Alten Test-Container entfernen, falls noch vorhanden
	docker rm -f $(E2E_CONTAINER_NAME) || true
	
	# 2. Neuen Container starten
	docker run --name $(E2E_CONTAINER_NAME) \
		-e POSTGRES_USER=$(E2E_DB_USER) \
		-e POSTGRES_PASSWORD=$(E2E_DB_PASSWORD) \
		-e POSTGRES_DB=$(E2E_DB_NAME) \
		-p $(E2E_DB_PORT):5432 \
		-d postgres:16-alpine
	
	# 3. Warten bis Postgres bereit ist
	@until docker exec $(E2E_CONTAINER_NAME) pg_isready; do sleep 1; done
	
	@echo "Führe Migrationen auf E2E-DB aus..."
	# Hier dein Migrations-Tool aufrufen, z.B.:
	migrate -path backend/migrations -database $(E2E_DB_URL) up
	
	@echo "Starte Tests..."
	cd frontend && npx playwright test --config playwright.local.config.ts $(args); \
	EXIT_CODE=$$?; \
	
	# 4. Cleanup: Nur den DB-Container löschen
	docker rm -f postgres_e2e; \
	exit $$EXIT_CODE