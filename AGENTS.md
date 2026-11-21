# AGENTS.md - Development Guidelines for Go Kegelmaster

This file contains essential information for agentic coding agents working in this repository.

## Build, Test, and Development Commands

### Lokale Test-Instanz (aktueller Branch)
- **Start**: `make dev` – startet einen eigenen Postgres-Container (`kegelmaster-dev-postgres`, Port 5434, Volume `kegelmaster-dev-pgdata`), baut das Backend aus dem aktuellen Checkout, migriert automatisch und startet Backend (:8080) + Vite (:5173). Beenden mit Ctrl-C; die Daten bleiben erhalten.
- **Optionen**: `FRESH=1` verwirft die Datenbank vorher, `SEED=1` lädt Beispiel-Daten; `make dev-fresh` = beides. Ports über `DEV_DB_PORT`, `DEV_BACKEND_PORT`, `DEV_FRONTEND_PORT` änderbar.
- **Einzelschritte**: `make dev-db-up`, `make dev-db-down` (stoppt, Daten bleiben), `make dev-db-reset` (leere DB), `make dev-seed`, `make dev-psql`.
- **Beispiel-Daten** (`backend/cmd/seed`): Club „KC Alle Neune“ mit Rollen (Gast ohne Rollen-Leserecht), 9 Spielern, Strafentypen, Wettbewerben, 7 veröffentlichten Spieltagen + 1 Entwurf, Einzahlungen und einer Ausgabe. Buchungen laufen über die Repositories, Salden sind also konsistent. Logins: `demo@kegelmaster.local` (Owner) und `mitglied@kegelmaster.local` (nur Lesen), Passwort jeweils `kegelmaster`; beide gelten als „Rundgang gesehen“. Dazu eine offene Einladung für den Spieler Frank (Rolle Gast) unter `/invite/e2e-einladung-kc-alle-neune` (E-Mail `eingeladen@kegelmaster.local`, für den E2E-Test; nur einmal annehmbar, danach `make dev-fresh`). Der Seed überspringt sich, wenn der Demo-User schon existiert.

### Backend Commands
- **Run server**: `make backend-run` (starts Go API server)
- **Run tests**: `make backend-test` (runs all Go tests)
- **Run single test**: `cd backend && go test ./internal/club -run TestRepository_Create`
- **Run single test verbose**: `cd backend && go test -v ./internal/club -run TestRepository_Create`
- **Run with coverage**: `cd backend && go test -cover ./...`
- **List all tests**: `cd backend && go test -list=. ./internal/club`
- **Build binary**: `make backend-build`
- **Build import CLI**: `make import-build` (builds `bin/import` for CSV-Spieltag-Import)
- **Format code**: `make fmt` (uses gofmt on cmd/ and internal/)
- **Generate SQL code**: `make sqlc-generate` (runs sqlc for type-safe queries)

### Database Commands
- **Migrate up**: `make migrate-up` (applies pending migrations)
- **Migrate down**: `make migrate-down` (rolls back migrations)
- **Create migration**: `migrate create -ext sql -dir backend/migrations migration_name`
- **Database URL**: Set via `DATABASE_URL` env var (defaults to postgres://kegelmaster:kegelmaster@localhost:5432/kegelmaster?sslmode=disable)
- **Auto-migration**: Set `AUTO_MIGRATE=true` (or `1`, `on`) to run pending migrations on application startup. Migration files are embedded in the binary. When not set (default: false), use `make migrate-up` for manual migration.

### CLI Import Tool (Spieltag-CSV)
- **Build**: `make import-build` → binary `bin/import`
- **Run**: `./bin/import --club "Club-Name" --file pfad/zur.csv [--database-url URL] [--dry-run]`
- **Flags**: `--club` (Pflicht), `--file` (Pflicht), `--database-url` (optional, sonst `DATABASE_URL`), `--dry-run` (nur anzeigen, keine DB-Änderungen)
- **CSV**: Semikolon-getrennt; erste Zeile Header: Datum, Name, dann Strafen-/Wettbewerbsnamen; Datenzeilen: DD.MM.YYYY, Spielername, Zahlen. Unbekannte Spieler oder Spalten führen zu Fehlerabbruch. Ausführung ist idempotent.

### CLI Migrate-Old (alte DB → neue App-DB)
- **Build**: `make migrate-old-build` → binary `bin/migrate-old`
- **Run**: `./bin/migrate-old --source "postgres://..." --target "postgres://..." [--dry-run]`
- **Flags**: `--source` (Pflicht, URL der alten PostgreSQL-DB), `--target` (Pflicht, URL der neuen App-DB), `--dry-run` (nur lesen, keine Schreibzugriffe auf die Ziel-DB)
- **Idempotenz**: Deterministische UUIDs aus alten bigint-IDs; `INSERT ... ON CONFLICT (id) DO NOTHING`. Mehrfaches Ausführen erzeugt keine Duplikate.

### CLI Admin (Instanz-Admin vergeben)
- Subcommand im Haupt-Binary, damit es im Container verfügbar ist: `kegelmaster admin [--database-url URL] grant <email> | revoke <email> | list` (lokal: `cd backend && go run ./cmd/api admin list`).
- Im Cluster: `kubectl exec deploy/kegelmaster -- kegelmaster admin grant betreiber@example.com` (nutzt `DATABASE_URL` aus der Umgebung); mit Docker Compose: `docker compose exec backend kegelmaster admin list`.
- Führt keine Migrationen aus. Exit-Codes: 0 ok, 1 Fehler/Konto nicht gefunden, 2 falsche Verwendung. Änderungen werden als Audit `admin_role_granted`/`admin_role_revoked` (stderr) geloggt.
- Die Rolle wird bewusst **nie** über die Weboberfläche vergeben (siehe auch `INSTANCE_ADMIN_EMAILS`).

### Frontend Commands
- **Dev server**: `make frontend-dev` or `cd frontend && npm run dev` (starts Vite dev server)
- **Build**: `make frontend-build` or `cd frontend && npm run build`
- **Lint**: `cd frontend && npm run lint` (ESLint for TypeScript/React)
- **Preview**: `cd frontend && npm run preview` (preview production build)
- **E2E (Playwright)**: `cd frontend && npm run e2e` (Backend + Frontend müssen laufen; BASE_URL optional, Default localhost:5173). Die Admin-Specs brauchen `E2E_DATABASE_URL` (DB der laufenden Instanz), `onboarding.spec.ts` die Seed-Einladung (sonst wird dieser Test übersprungen).

### Docker Commands
- **Start all services**: `make compose-up` (PostgreSQL + backend + frontend)
- **Stop all services**: `make compose-down` (removes volumes with -v flag)

## Project Structure

```
backend/
├── cmd/
│   ├── api/             # Application entry point (main.go)
│   ├── import/          # CLI Import-Tool (Spieltag-CSV)
│   └── migrate-old/     # CLI Migration alter DB → neue App-DB
├── internal/            # Private application code
│   ├── admin/           # Instanz-Admin: Betreiber-Übersicht und Eingriffe (Queries in queries/admin.sql), CLI, Env-Bootstrap
│   ├── auth/            # Authentication service
│   ├── automigrate/     # Auto-run DB migrations on startup (embedded, optional)
│   ├── club/            # Club entity & repository
│   ├── config/          # Configuration management
│   ├── database/        # Database connection
│   ├── db/              # SQLc generated code (DO NOT EDIT)
│   ├── gamedayimport/   # CSV parsing, resolve, apply for import CLI
│   ├── oldmigrate/      # Migration alter DB-Schema → neues Schema (deterministische UUIDs, ON CONFLICT DO NOTHING)
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
- Use foreign key constraints with `ON DELETE CASCADE` (exception: `players.role_id` is `ON DELETE RESTRICT`, every player keeps exactly one role)
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
- CSRF protection with double-submit tokens (skipped for `Authorization: Bearer` requests)
- Personal API tokens (`internal/apitoken`, table `api_tokens`) for scripts/agents: `Authorization: Bearer kgm_…`. `UserFromCookie` checks the bearer header first and never falls back to the cookie when it is present. Tokens are managed via `/api/auth/tokens` (browser session only).
- MCP server (`internal/mcpserver`, mounted at `/mcp`): Streamable HTTP, stateless, auth with personal API tokens or OAuth access tokens (see next bullet). Tools never touch repositories directly; they call the REST API in-process via `InProcessAPI` (Fiber request handler) with the caller's token, so new tools get permissions, validation, audit and balance invariants for free. Tool names/descriptions are German; amounts are cents. Register tools in `tools.go`, test them in `server_test.go` against `fakeAPI`.
- OAuth 2.1 for MCP clients (KEG-12, `internal/oauth`, `handlers/oauth.go`): Kegelmaster is its own authorization server, issuer = `PUBLIC_BASE_URL`. `/mcp` answers 401 with `WWW-Authenticate: Bearer resource_metadata="<base>/.well-known/oauth-protected-resource"`; metadata at `/.well-known/oauth-protected-resource[/mcp]` and `/.well-known/oauth-authorization-server[/mcp]`; `POST /oauth/register` (RFC 7591, public clients only, redirect URIs https or loopback http), `GET /oauth/authorize` (code + PKCE S256 only, forwards to the frontend consent page `/oauth/consent`), `POST /oauth/token` (authorization_code, refresh_token), `POST /oauth/revoke`. Consent and connected apps under `/api/oauth/*` (browser session only, CSRF, audit). Tables `oauth_clients`, `oauth_grants` (one approved connection = one refresh chain), `oauth_authorization_codes`, `oauth_tokens`; hashes only, prefixes `kgc_`/`kgo_`/`kgr_`. Codes 10 min single use (reuse revokes the grant), access tokens 1 h, refresh tokens 30 d rotating (reuse revokes the grant). Access tokens are bound to `/mcp`: `UserFromCookie` accepts `kgo_…` only on requests the MCP server dispatches in-process (`InProcessAPI` sets the fasthttp user value `oauth.InProcessKey{}`), so they never work against the REST API directly; grants created before `password_changed_at`/`sessions_valid_after` are rejected like sessions.
- Account settings: `PATCH /api/auth/me` (e.g. `notify_gameday_settled`, opt-in, default false; `tour_seen` / `setup_hint_seen` set the Rundgang flags `tour_seen_at` / `setup_hint_seen_at` once)
- Rundgang und Ersteinrichtung (GLOSSARY.md): The Rundgang is driver.js wrapped in `frontend/src/components/tour/` (step catalog with `data-tour` targets and visibility conditions, start logic in `TourAutoStart` on the dashboard, triggers via `lib/onboarding.ts`). The Ersteinrichtung is computed on every request by `GET /api/clubs/{id}/setup` (owner only, never stored); `POST /api/clubs/{id}/setup/dismiss` sets `clubs.setup_dismissed_at` (belongs to the club, survives an ownership transfer). `POST /api/clubs/{clubId}/roles/suggestion` creates Präsident, Kassenwart, Mitglied and Gast (409 if the club has roles). Creating a player with `link_me: true` links the caller („Das bin ich“). Pages open their create dialog with `?neu=1` (`hooks/use-open-from-query.ts`).
- Every player has exactly one role: `players.role_id` is `NOT NULL` with `ON DELETE RESTRICT`; deleting a role that still has players answers 409.
- E-mail notifications (`internal/notification`): when a game day goes from draft to published (create with `is_draft: false` or update draft → published), the game day handler calls `GameDayNotifier.GameDaySettled` after the bookings were committed. The notifier runs in a goroutine with its own timeout, claims the game day once in `game_day_notifications` (re-publishing sends nothing), and e-mails every user with `notify_gameday_settled` who has a player in the club (inactive players only if they were charged), including the day's bookings and current balance. Mail errors are only logged.
- Instance admin (operator) area `/api/admin` (`internal/admin`, `handlers/admin.go`): read-only overview of users, clubs, owners and player↔user links (stats, users, clubs, details, CSV export). Guarded by `Handler.RequireInstanceAdmin` (route group middleware), which is independent of the club `PermissionChecker`: cookie session only, requests with `Authorization: Bearer` (API tokens, hence MCP) and non-admins get **404** (not 401/403). Never expose club business data there (balances, transactions, penalties of players), only aggregated counts. Every access is audited via `audit.LogAdminAudit` (adds `admin=true`; every audit entry carries the client `ip`). `users.is_instance_admin` is set only via CLI `kegelmaster admin …` or `INSTANCE_ADMIN_EMAILS`, never via the web UI. `GET /api/auth/me` returns `is_instance_admin`; `users.last_login_at` is set on password and passkey login. Interventions (`handlers/admin_actions.go`): disable/enable (`users.disabled_at`; login → 403, cookie sessions and API tokens → 401 in `UserFromCookie`/`userFromAPIToken`), revoke sessions (`users.sessions_valid_after`, checked by `User.SessionValid` besides `password_changed_at`; API tokens unaffected), revoke one API token, transfer club owner (shared `transferClubOwner` with `HandleTransferClubOwner`), delete user (`confirm_email`; 409 while the user still owns clubs, since `clubs.user_id` cascades) and delete club (`confirm_name`). All validate CSRF and are audited with `admin=true`. The acting admin cannot disable/delete themselves, and other instance admins cannot be disabled/deleted via the web (409) – removing an admin stays a CLI task. Re-auth (KEG-9, `handlers/admin_reauth.go`): the session JWT carries `auth_time` (time of the last real authentication; `auth.Service.GenerateToken` sets it to now, `GenerateTokenWithAuthTime` keeps it for reissues without authentication, e.g. a future sliding refresh). After the admin checks, `RequireInstanceAdmin` requires `auth_time` within `ADMIN_REAUTH_MINUTES` (default 15), otherwise **401** `{"error":"reauth_required"}` (non-admins still get 404; cookies without `auth_time` count as stale). `POST /api/auth/reauth` (password) and `POST /api/auth/passkeys/reauth` (ceremony from `/passkeys/login/begin`, passkey must belong to the session user) reissue the cookie with a fresh `auth_time` (keeping remember-me, rotating CSRF); cookie session only, CSRF, rate-limited, audited as `user_reauth`/`user_reauth_failed`. Frontend: `lib/admin-reauth.ts` + `components/admin/admin-reauth-dialog.tsx` (mounted in `AdminLayout`).
- User context available in handlers
- Owner-based authorization for clubs

## Configuration

### Environment Variables
- `DATABASE_URL`: PostgreSQL connection string
- `AUTO_MIGRATE`: Optional. When set (e.g. `true`, `1`, `on`), the app runs pending DB migrations on startup using embedded migration files. Default: false.
- `JWT_SECRET`: JWT signing secret
- `JWT_TTL`: Token time-to-live
- `REGISTRATION_MODE`: (optional) Wer ein Konto anlegen darf: `open` (Default, rückwärtskompatibel), `invite_only` (`POST /api/auth/register` nur mit `invite_token` einer gültigen, nicht abgelaufenen, nicht angenommenen Einladung aus `player_invitations`; die E-Mail muss zur Einladung passen) oder `closed` (Registrierung antwortet immer mit 403). Groß-/Kleinschreibung egal; ungültige Werte verhindern den Start. Wird als `registration_mode` unter `GET /api/config/public` ausgeliefert; das Frontend blendet Registrierungsformular und -links entsprechend aus. Bestehende Konten (Login, Passkeys, Einladung annehmen) sind nicht betroffen.
- `GEMINI_API_KEY`: (optional) API-Key für den Chatbot-Assistenten (Hilfe zur App). Wenn nicht gesetzt, antwortet `POST /api/chat` mit 503 und `GET /api/config/public` liefert `chat_enabled: false`, sodass das Chat-Widget im Frontend ausgeblendet wird.
- `GEMINI_MODEL`: (optional) Gemini-Modellname (z.B. `gemini-1.5-flash`, `gemini-2.0-flash`). Default: `gemini-1.5-flash`.
- `CHAT_RATE_LIMIT_MAX`: (optional) Max. Chat-Requests pro User pro Fenster. Default: 30.
- `CHAT_RATE_LIMIT_WINDOW_MINUTES`: (optional) Fenster in Minuten für das Chat-Rate-Limit. Default: 1.
- `LEGAL_IMPRESSUM_PATH`: (optional) Absoluter Dateipfad zur Impressum-HTML-Datei. Wenn gesetzt, liefert `GET /api/legal/impressum` den Dateiinhalt.
- `LEGAL_DATENSCHUTZ_PATH`: (optional) Absoluter Dateipfad zur Datenschutz-HTML-Datei. Wenn gesetzt, liefert `GET /api/legal/datenschutz` den Dateiinhalt.
- `IMPRESSUM_HTML`: (optional) Inhalt für das Impressum (HTML-String). Überschreibt Datei, wenn gesetzt.
- `DATENSCHUTZ_HTML`: (optional) Inhalt für die Datenschutzerklärung (HTML-String). Überschreibt Datei, wenn gesetzt.
  - **Datenschutz-Hinweis für Betreiber:** Mit der Betreiber-Übersicht (`/api/admin`) sieht der Instanz-Admin Konto-Metadaten (E-Mail, Registrierungs- und Login-Zeitpunkte, Klubzugehörigkeiten, Namen/Präfixe von API-Tokens und Passkeys) zum Betrieb und zur Missbrauchsabwehr (Art. 6 Abs. 1 lit. f DSGVO). Das gehört in die Datenschutzerklärung der Instanz, ebenso die Aufbewahrungsfrist der Audit-Logs (im Log-Stack). Fachdaten der Klubs sind dort nicht sichtbar.
- `BUY_ME_A_COFFEE_URL`: (optional) URL für „Buy Me a Coffee“. Wenn gesetzt, wird im App-Footer ein entsprechender Link angezeigt; ausgegeben unter `GET /api/config/public` als `buy_me_a_coffee_url`.
- `PROXY_HEADER`: (optional) Header mit der echten Client-IP hinter einem Reverse Proxy (z. B. `X-Forwarded-For`). Ohne diese Einstellung sehen Rate-Limits (Login, Passwort-Reset, Chat) nur die Proxy-IP und gelten damit für alle Nutzer gemeinsam.
- `TRUSTED_PROXIES`: (optional, nur mit `PROXY_HEADER`) Kommagetrennte IPs/CIDRs der vertrauenswürdigen Proxies; zusätzlich `private`, `loopback`, `linklocal`. Nur von diesen Adressen wird `PROXY_HEADER` ausgewertet.
- `WEBAUTHN_RP_ID`: (optional) Domain, an die Passkeys gebunden sind (WebAuthn Relying Party ID). Default: Host aus `PUBLIC_BASE_URL`. Achtung: Ein späterer Wechsel macht alle registrierten Passkeys unbrauchbar.
- `WEBAUTHN_ORIGINS`: (optional) Kommagetrennte Origins (`https://host[:port]`), die Passkey-Anmeldungen ausführen dürfen. Default: Origin aus `PUBLIC_BASE_URL` plus `CORS_ALLOW_ORIGINS` (damit funktioniert der Vite-Dev-Server auf `:5173`). Ist die Konfiguration ungültig, startet der Server trotzdem; die Passkey-Endpunkte antworten dann mit `503`.
- `RESEND_API_KEY`: (optional) API-Key für den E-Mail-Versand über Resend (Einladungen, Passwort-Reset, Benachrichtigung „Spieltag abgerechnet“). Ohne Key werden keine E-Mails verschickt; Spieltag-Benachrichtigungen werden dann mit einer Log-Zeile übersprungen.
- `RESEND_FROM_EMAIL`: (optional) Absenderadresse für E-Mails. Default: `noreply@example.com`.
- `INSTANCE_ADMIN_EMAILS`: (optional) Kommagetrennte E-Mails, die beim Start das Instanz-Admin-Flag erhalten (Betreiber-Übersicht `/api/admin`). Nur bestehende Konten (unbekannte E-Mails werden geloggt); es wird **nie** etwas entzogen. Zum Entziehen die E-Mail entfernen **und** `kegelmaster admin revoke <email>` ausführen, sonst setzt der nächste Start das Flag wieder.
- `ADMIN_REAUTH_MINUTES`: (optional) Minuten seit der letzten echten Anmeldung (Passwort- oder Passkey-Login bzw. erneute Bestätigung), in denen der Admin-Bereich `/api/admin` nutzbar ist. Danach antworten die Admin-Endpunkte einem Instanz-Admin mit `401 {"error":"reauth_required"}`, und die Admin-UI fragt Passwort oder Passkey ab (`POST /api/auth/reauth` bzw. `POST /api/auth/passkeys/reauth`). Gilt unabhängig davon, ob ein Passkey registriert ist. Ganzzahl 1–1440, Default `15`; ungültige Werte verhindern den Start.
- `LOG_LEVEL`: (optional) Log-Level (`debug`, `info`, `warn`, `error`). Bei `debug` werden auch Leseanfragen (GET/HEAD) im Request-Log erfasst; sonst nur Mutationen und Fehler. Default: `info`.

### GitHub Actions / Release Please
- **RELEASE_PLEASE_TOKEN:** (Repository-Secret) Personal Access Token (PAT) mit mindestens Scope `repo`. Wird vom Workflow `.github/workflows/release-please.yaml` verwendet. Ohne dieses Secret nutzt Release Please den Standard-`GITHUB_TOKEN`; dann lösen von Release Please erzeugte Events (z. B. „release published“) keine weiteren Workflows aus – der Workflow `docker-release` würde nach einem Release nicht starten. Secret unter Settings → Secrets and variables → Actions anlegen (z. B. Name `RELEASE_PLEASE_TOKEN`).

### Abhängigkeits-Updates (Renovate)
- Konfiguration in `renovate.json` (Repo-Root). Voraussetzung: Ein Repo-Admin muss die [Renovate GitHub App](https://github.com/apps/renovate) für `schnurbus/go-kegelmaster` installieren; danach legt Renovate ein „Dependency Dashboard“-Issue an.
- Abgedeckt: Go-Module (`backend/go.mod`, mit `go mod tidy`), npm (`frontend/`), Dockerfiles, `docker-compose.yaml`, GitHub Actions.
- Zeitplan: montags vor 6 Uhr (Europe/Berlin), max. 5 offene PRs / 2 pro Stunde. Minor/Patch werden je Ökosystem gebündelt (Go-Abhängigkeiten, Frontend-Abhängigkeiten, GitHub Actions, Docker-Images), Majors kommen einzeln. Kein Automerge.
- PR-Titel: `chore(deps): ...` (kein Release). Sicherheits-Updates (GitHub-Alerts/OSV) kommen sofort, mit Label `security` und als `fix(deps): ...`, damit Release Please ein Patch-Release erzeugt.
- Majors von `postgres`/`redis`-Images (Datenmigration nötig) erst nach Freigabe im Dependency Dashboard.

### Security / Production
- **JWT_SECRET:** In Production zwingend setzen (kein Default „dev-secret-change-me“ verwenden). Fehlt das Secret, beendet sich die App beim Start.
- **PUBLIC_BASE_URL:** Muss auf die vertrauenswürdige App-Origin zeigen (z. B. Frontend-URL). Wird u. a. für Passwort-Reset-Links und den App-Link in Spieltag-Benachrichtigungen (`PUBLIC_BASE_URL/app`) verwendet und ist der OAuth-Issuer für MCP-Clients (Metadaten, `resource` = `PUBLIC_BASE_URL/mcp`); dafür muss die URL öffentlich per https erreichbar sein und `/mcp`, `/oauth/*` und `/.well-known/*` an das Backend leiten. Bei Deployment prüfen, dass keine fremde/Phishing-Domain konfiguriert ist.

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

### Conventional Commits

Use [Conventional Commits](https://www.conventionalcommits.org/) for all commit messages:

- **Format:** `type(scope): description` (lowercase, no period at end of subject)
- **Types:** `feat` (feature), `fix` (bug fix), `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`
- **Scope** (optional): affected area, e.g. `transactions`, `auth`, `frontend`
- **Body** (optional): detailed explanation; separate from subject by blank line
- **Breaking change:** add `!` after type/scope or footer `BREAKING CHANGE: ...`

Examples:
- `feat(transactions): allow transaction date for manual transactions`
- `fix(transactions): require player for tip transactions`
- `docs: update AGENTS.md with conventional commits`

## Important Notes

- Never edit files in `backend/internal/db/` (SQLc generated)
- Always use context.Context for database operations
- Follow the repository pattern for data access
- Maintain high test coverage (current: 84-93%)
- Use UUIDs for all entity IDs
- Implement proper error handling and logging
- **OpenAPI Spec:** Bei neuen oder geänderten API-Endpunkten die OpenAPI-Spezifikation in `backend/openapi/openapi.yaml` anpassen und ggf. Schemas ergänzen. Die Spec wird unter `/api/openapi.yaml` ausgeliefert; Swagger UI unter `/api/docs`.