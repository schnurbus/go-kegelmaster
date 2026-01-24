DATABASE_URL ?= postgres://kegelmaster:kegelmaster@localhost:5432/kegelmaster?sslmode=disable

.PHONY: backend-run backend-test backend-build frontend-dev frontend-build compose-up compose-down fmt migrate-up migrate-down sqlc-generate

backend-run:
	@cd backend && go run ./cmd/api

backend-test:
	@cd backend && go test ./...

backend-build:
	@cd backend && go build ./cmd/api

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

