# AGENTS.md - Development Guidelines for Go Kegelmaster

This file contains essential information for agentic coding agents working in this repository.

## Build, Test, and Development Commands

### Backend Commands
- **Run server**: `make backend-run` (starts Go API server)
- **Run tests**: `make backend-test` (runs all Go tests)
- **Run single test**: `cd backend && go test ./internal/club -run TestRepository_Create`
- **Run single test verbose**: `cd backend && go test -v ./internal/club -run TestRepository_Create`
- **Run with coverage**: `cd backend && go test -cover ./...`
- **List all tests**: `cd backend && go test -list=. ./internal/club`
- **Build binary**: `make backend-build`
- **Format code**: `make fmt` (uses gofmt on cmd/ and internal/)
- **Generate SQL code**: `make sqlc-generate` (runs sqlc for type-safe queries)

### Database Commands
- **Migrate up**: `make migrate-up` (applies pending migrations)
- **Migrate down**: `make migrate-down` (rolls back migrations)
- **Create migration**: `migrate create -ext sql -dir backend/migrations migration_name`
- **Database URL**: Set via `DATABASE_URL` env var (defaults to postgres://kegelmaster:kegelmaster@localhost:5432/kegelmaster?sslmode=disable)

### Frontend Commands
- **Dev server**: `make frontend-dev` or `cd frontend && npm run dev` (starts Vite dev server)
- **Build**: `make frontend-build` or `cd frontend && npm run build`
- **Lint**: `cd frontend && npm run lint` (ESLint for TypeScript/React)
- **Preview**: `cd frontend && npm run preview` (preview production build)

### Docker Commands
- **Start all services**: `make compose-up` (PostgreSQL + backend + frontend)
- **Stop all services**: `make compose-down` (removes volumes with -v flag)

## Project Structure

```
backend/
├── cmd/api/              # Application entry point (main.go)
├── internal/             # Private application code
│   ├── auth/            # Authentication service
│   ├── club/            # Club entity & repository
│   ├── config/          # Configuration management
│   ├── database/        # Database connection
│   ├── db/              # SQLc generated code (DO NOT EDIT)
│   ├── handlers/        # HTTP handlers
│   ├── permission/      # Permission system
│   ├── player/          # Player entity & repository
│   ├── role/            # Role entity & repository
│   └── server/          # Server setup
├── migrations/          # Database migrations
├── queries/             # SQL queries for SQLc
└── schema/              # Database schema definitions
```

## Code Style Guidelines

### Go Conventions
- Follow standard Go project layout with `internal/` for private code
- Use PascalCase for exported types/functions, camelCase for unexported
- Package names are lowercase and descriptive (auth, club, player, role)
- Context-aware functions: first parameter should be `ctx context.Context`

### Import Organization
```go
import (
    "context"          // Standard library
    "fmt"
    "time"

    "github.com/gofiber/fiber/v3"  // Third-party
    "github.com/google/uuid"

    "github.com/schnurbus/go-kegelmaster/backend/internal/auth"  // Internal
)
```

### Entity Structure Pattern
```go
type Entity struct {
    ID        string     `json:"id"`       // UUID as string
    Name      string     `json:"name"`
    UserID    string     `json:"user_id"`  // UUID as string
    CreatedAt time.Time  `json:"created_at"`
    UpdatedAt time.Time  `json:"updated_at"`
}
```

### Repository Pattern
- Interface in `repository.go`
- Implementation in `repository.go`
- Tests in `repository_test.go` with sqlmock
- Use SQLc generated queries in `internal/db/`

### Error Handling
- Use custom error types: `var ErrNotFound = errors.New("entity not found")`
- Return errors from functions: `return nil, ErrNotFound`
- Log errors with structured logging (slog)

### Database Patterns
- Use SQLc for type-safe SQL queries
- All tables have UUID primary keys
- Include `created_at` and `updated_at` timestamps
- Use foreign key constraints with `ON DELETE CASCADE`
- User-owned resources reference `users(id)`

## Testing Guidelines

### Test Structure
- Test files: `*_test.go`
- Use sqlmock for database testing
- Test both success and error cases
- Verify mock expectations: `if err := mock.ExpectationsWereMet(); err != nil`

### Test Example Pattern
```go
func TestRepository_Create(t *testing.T) {
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()

    // Setup expectations
    mock.ExpectQuery(`INSERT INTO clubs`).
        WithArgs(...).
        WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
            AddRow(uuid.New(), time.Now()))

    // Execute test
    repo := NewRepository(db)
    result, err := repo.Create(ctx, params)
    
    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

### Running Tests
- All tests: `make backend-test`
- Single test: `cd backend && go test ./internal/club -run TestRepository_Create`
- With coverage: `cd backend && go test -cover ./...`

## API Development

### Handler Pattern
- Use Fiber v3 framework
- Handlers in `internal/handlers/`
- Middleware for auth, CORS, CSRF
- Return JSON responses with consistent structure

### Authentication
- JWT-based authentication
- CSRF protection with double-submit tokens
- User context available in handlers
- Owner-based authorization for clubs

## Configuration

### Environment Variables
- `DATABASE_URL`: PostgreSQL connection string
- `JWT_SECRET`: JWT signing secret
- `JWT_TTL`: Token time-to-live

### Config Structure
- Configuration in `internal/config/`
- Use struct tags for env var mapping
- Provide sensible defaults

## Database Development

### Migrations
- Create new migrations: `migrate create -ext sql -dir backend/migrations migration_name`
- Use UUID primary keys
- Include proper foreign key constraints
- Add indexes for performance

### SQLc Workflow
1. Write SQL queries in `backend/queries/`
2. Define schema in `backend/schema/`
3. Run `make sqlc-generate` to create Go code
4. Use generated code in repositories

## Security Guidelines

- Never commit secrets or API keys
- Use environment variables for configuration
- Implement proper authentication and authorization
- Validate all input data
- Use parameterized queries (SQLc handles this)

## Frontend Integration

- Backend serves API at `/api`
- Frontend built with React, TypeScript, Vite
- Uses Shadcn UI components
- Follows React Router v7 patterns

## Development Workflow

1. Always run tests before committing: `make backend-test`
2. Format code: `make fmt`
3. Check SQLc generation: `make sqlc-generate`
4. Update PROJECT.md with progress (see .cursor/rules/project-md.mdc)
5. Test with Docker: `make compose-up`

## Important Notes

- Never edit files in `backend/internal/db/` (SQLc generated)
- Always use context.Context for database operations
- Follow the repository pattern for data access
- Maintain high test coverage (current: 84-93%)
- Use UUIDs for all entity IDs
- Implement proper error handling and logging