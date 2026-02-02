DATABASE_URL ?= postgres://kegelmaster:kegelmaster@localhost:5432/kegelmaster?sslmode=disable

.PHONY: backend-run backend-test backend-build backend-build-embed build import-build frontend-dev frontend-build compose-up compose-down fmt migrate-up migrate-down sqlc-generate

backend-run:
	@cd backend && go run ./cmd/api

backend-test:
	@cd backend && go test ./...

backend-build:
	@cd backend && go build ./cmd/api

# Build CLI import tool (CSV → Spieltage).
import-build:
	@mkdir -p bin && cd backend && go build -o ../bin/import ./cmd/import

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

