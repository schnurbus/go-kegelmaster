# Web-App Projektspezifikation (Go Backend + React Router v7 Frontend)

---

## 1. Projektübersicht

**Projektname:** go-kegelmaster

**Kurzbeschreibung:**
Diese App dient zur Nachverfolgung der Spielergebnisse mehrere Kegelklubs.

Jeder Kegelklub het mehrere Mitglieder, die sich regelmäßig zum Kegeln treffen. Bei diesen Treffen werde Spiele gespielt, bei denen es zu Strafen kommen kann. Jede Strafe hat einen Gebührenwert. Die Art und der Name der Strafen, sowie die Gebühren, können bei jedem Kegelklub unterschiedlich sein. Ebenfalls gibt es in jedem Kegelklub verschieden Rollen (z.B. Präsident, Kassenwart, Mitglied, Gast, etc.). Diese Rollen sollen verschiedene Rechte haben, die ebenfalls bei jedem Klub unterschiedlich sein können.

Insgesamt soll es für die Mitglieder möglich sein, einen Spielabend eintragen zu können, inkl. der Informationen, welche Spieler anwesend waren, und wieviel Strafen sie pro Strafentyp angesammelt haben. Daraus muss sich ein Saldo-Wert für jeden Spieler ergeben. Ebenfalls soll es möglich sein, Einzahlungen von Spielern einzutragen, um diesen Saldo-Wert auszugleichen, und dafür den Saldo des Klubs zu erhöhen.

Ebenso soll es für die Spieler möglich sein, sich mit ihrem Account anzumelden, und Statistiken zu ihren vergangenen Wert und zum Klub anzuschauen.

**Zielgruppen / Nutzer:**
- Primäre Nutzer: Normale Nutzer ohne technischen Hintergrund
- Sekundäre Nutzer:
- Administratoren: Betreiber der Applikation

**Ziele & Erfolgskriterien:**
-
-

---

## 2. Rollen & Berechtigungen

Die Rollen und Berechtigungen müssen dynamisch pro Klub konfigurierbar sein. Zum Beispiel:
- Spieler X hat Rolle "Kassenwart"
- Rolle "Kassenwart" darf Gebühren anlegen und Ändern
- Rolle "Kassenwart" darf Spielabende und Strafen eintragen
- Rolle "Kassenwart" darf nicht Grundgebühr des Klubs ändern

---

## 3. Funktionen (Featureliste)

### 3.1 Muss-Funktionen
- ✅ Feature 1: Multi-Tenancy (Grundlage: User-Authentifizierung)
- ✅ Feature 2: Auswahl des Klubs (Club Switcher in Sidebar, localStorage-Persistierung)
- ✅ Feature 3: Anlegen von Strafen mit Gebührenwert (Penalty Types mit Display Order, Soft Delete, Replace-Mechanismus)
- ✅ Feature 4: Anlegen von Spielern mit Start-Saldo (Player Management mit Rollen-Zuordnung)
- ✅ Feature 5: Anlegen von Spieltagen mit Datum und Notizen (GameDay Entity mit Participants)
- ⏳ Feature 6: Eintragen von Anzahl der Strafen pro Spieler und Spieltag (API komplett, UI-Erweiterung offen)
- ⏳ Feature 7: Verknüpfung von Spielern und Usern
- Feature 8: 

### 3.2 Soll-Funktionen
- Feature 1: Übersicht über alle Strafen und Saldo-Änderungen
- Feature 2: Übersicht aller Statistiken pro Spieler im Spieler-Profil

### 3.3 Kann-Funktionen (Nice-to-Have)
- Feature 1: 
- Feature 2:

Weitere Features kommen mit der Zeit dazu

---

## 4. Systemarchitektur

### Backend (Go)
- Framework: GoFiber
- Authentifizierung: lokal & Social Login
- Datenbank: PostgreSQL
- ORM/Query Builder: sqlc
- Logging: slog
- Caching: redis
- Migrationstool: go migrate
- Deployment: Helm-4-Chart unter `helm/` (nur App; Postgres/Redis extern)

**Aktueller Stand (25.11.2025)**
- Fiber v3 Boilerplate mit `/healthz` und `/api` Platzhaltern
- Konfigurations-Layer `internal/config` inkl. Defaults für DB/Redis
- Graceful Shutdown & Logging (`cmd/api/main.go`)
- Dockerfile für Container-Build hinterlegt
- Auth-Service (`internal/auth`) inkl. Passwort-Hashing & JWT-Ausstellung
- User-Repository (`internal/user`) mit Postgres-Persistenz
- Endpunkte `/api/auth/register`, `/api/auth/login`, `/api/auth/logout`
- Session-/Profil-Endpunkte `/api/auth/me`, `/api/auth/csrf-token`
- CORS + Double-Submit CSRF-Token inklusive Cookie-/Header-Validierung
- Persönliche API-Tokens (`internal/apitoken`) für Skripte/Agenten: `Authorization: Bearer kgm_…`, nur SHA-256-Hash gespeichert, kein CSRF nötig
- Passkeys (`internal/passkey`, WebAuthn via `go-webauthn`) als zusätzliche Anmeldemethode: discoverable Credentials, Challenge einmalig in `webauthn_sessions`
- MCP-Server (`internal/mcpserver`) unter `/mcp` für KI-Agenten: Streamable HTTP (zustandslos), Auth per API-Token; Tools rufen die REST-API in-process mit dem Token des Aufrufers auf
- Migration `0001_create_users` angelegt
- **Club Entity & Repository** (`internal/club`) mit vollständigem CRUD
- **Club API Endpunkte**: `GET /api/clubs`, `GET /api/clubs/:id`, `POST /api/clubs`, `PUT /api/clubs/:id`, `DELETE /api/clubs/:id`
- **Owner-Check**: Nur der Owner kann Clubs aktualisieren oder löschen
- **Migration `0002_create_clubs`** mit Foreign Key zu Users
- **Unit-Tests**: Club Repository (93.3% Coverage) und Server Handler (84.2% Coverage)

### Frontend (React)
- React Router v7
- State Management:
- UI Framework: Shadcn UI
- CSS Setup: Tailwind v4
- Build Tool: Vite

**Aktueller Stand (25.11.2025)**
- Vite (React + SWC + TS) Scaffold
- React Router v7 Setup mit Landing- und Dashboard-Routen
- AuthContext mit CSRF-Handshake, Login/Logout/Register Actions
- Seiten `/login` & `/register`, Weiterleitung nach erfolgreicher Anmeldung
- Protected Routes via `RequireAuth` für Dashboard & künftige private Views
- Dockerfile für Dev-Server vorhanden
- **Shadcn UI Komponenten**: Sidebar, Dialog, Dropdown, Button, Input, Label, etc.
- **Club Switcher** in Sidebar: Zeigt alle Clubs des Users, ermöglicht Club-Auswahl
- **Create Club Dialog**: Dialog zum Erstellen neuer Clubs mit Validierung
- **ClubContext**: Globaler State für Clubs mit localStorage-Persistierung
- **AppLayout** mit Sidebar und Header für Dashboard-Bereich

### API
- REST oder GraphQL: REST
- JSON-Format:

--- 

## 5. Datenmodell

### Entitäten (Models)

**Entity:** User
- id (uuid)
- email (string)
- password_hash (string)
- created_at (timestamp)
- updated_at (timestamp)

Relationen:
- id => Klub.owner (one:many)

**Entity:** Club
- id (uuid)
- name (string)
- balance (integer, Cent-Betrag)
- base_fee (integer, Cent-Betrag)
- user_id (uuid, Owner)
- created_at (timestamp)
- updated_at (timestamp)

Relationen:
- user_id => user.id (many:one, Foreign Key mit CASCADE DELETE)

**Entity:** Player
- id (uuid)
- name (string)
- balance (number)
- user_id (uuid)
- club_id (uuid)
- created_at (timestamp)
- updated_at (timestamp)

Relationen:
- user_id => user.id (one:one)
- club_id => club.id (one:one)

Weitere Entitäten kommen mit der Zeit dazu


---

## 6. Workflows / Use Cases

**Use Case Beispiel:**
- Klub anlegen
  1. User loggt sich eine
  2. User klickt auf "Neuer Klub"
  3. User gibt Namen an
  4. User gibt den Wert einer Grundgebühr pro Spielabend an
  5. User klickt auf "Klub anlegen"
- Spielabend eintragen
  1. User logt sich ein
  2. Sofern der User Spieler in mehreren Klubs hat, wählt er den aktuellen Klub über ein Select Element aus
  3. User Erstellt neuen Spielabend mit einem Datum
  4. User fügt vorhandene Spieler aus dem Klub zum Spielabend hinzu
  5. User gibt für jeden Spieler und aktiven Strafentyp die Anzahl sein
  6. User spiechert Spielabend ab
  7. Saldo der Spieler wird automatisch aktualisiert

*Weitere Workflows: Registrierung, Login, CRUD, Admin-Aktionen etc.*

---

## 7. Benutzeroberfläche (UI/UX)

### Screens / Seiten

**Screen:** Landing Page
- URL: /
- Elemente: Willkommensseite, Links zu Login/Register
- Interaktionen: Navigation zu Login/Register

**Screen:** Login
- URL: /login
- Elemente: Login-Formular (Email, Passwort)
- Interaktionen: Login, Weiterleitung zu Dashboard

**Screen:** Register
- URL: /register
- Elemente: Registrierungs-Formular (Email, Passwort)
- Interaktionen: Registrierung, automatischer Login, Weiterleitung zu Dashboard

**Screen:** Dashboard
- URL: /app
- Elemente: Sidebar mit Club Switcher, Navigation, Hauptinhalt
- Interaktionen: Club-Auswahl, Navigation zwischen Bereichen

**Screen:** Mein Account
- URL: /app/account (Benutzermenü unten links → „Account“)
- Elemente: Kontodaten (E-Mail, registriert am, Passwort zuletzt geändert), Liste der eigenen Clubs (Eigentümer/Mitglied), Formular „Passwort ändern“ (aktuelles, neues, Bestätigung), Passkeys (Liste mit Name, synchronisiert/gerätegebunden, hinzugefügt, zuletzt genutzt), Link zu den API-Tokens
- Interaktionen: Passwort ändern (meldet alle anderen Sitzungen ab und widerruft alle API-Tokens; der aktuelle Browser bleibt angemeldet); Passkey hinzufügen (aktuelles Passwort erforderlich), umbenennen, entfernen

**Screen:** Login
- URL: /login
- Elemente: E-Mail, Passwort, „Angemeldet bleiben“, Button „Mit Passkey anmelden“ (nur wenn der Browser Passkeys unterstützt)
- Interaktionen: Anmeldung per Passwort oder Passkey; unterstützt der Browser Autofill für Passkeys, werden sie direkt im E-Mail-Feld angeboten

**Screen:** API-Tokens
- URL: /app/api-tokens (Benutzermenü unten links → „API-Tokens“)
- Elemente: Tabelle der eigenen Tokens (Name, Präfix, erstellt, zuletzt genutzt, Ablauf), Dialog zum Erstellen (Name, Gültigkeit 30/90/365 Tage oder unbegrenzt)
- Interaktionen: Token erstellen (Klartext wird einmalig angezeigt, kopierbar), Token widerrufen

### Navigationsstruktur
- Public Routes: `/`, `/login`, `/register`
- Private Routes: `/app` (Dashboard)
- Sidebar Navigation: Dashboard, Spielabende, Statistiken, Klubs, Spieler

---

## 8. API-Spezifikation

### Dokumentation
- **OpenAPI Spec:** `backend/openapi/openapi.yaml` (manuell gepflegt). Wird unter `GET /api/openapi.yaml` ausgeliefert.
- **Swagger UI:** Interaktive API-Dokumentation unter `GET /api/docs` (lädt die Spec von `/api/openapi.yaml`). Für technische Nutzer/Entwickler.

### Auth
- Login
- Logout
- Token Refresh
- Registrieren

### Endpunkte

#### Auth
| Methode | Endpoint | Beschreibung | Request Body | Response |
|--------|----------|--------------|--------------|---------|
| POST | `/api/auth/register` | Neuen User anlegen & automatisch einloggen | `{ email, password }` | `201` + User JSON |
| POST | `/api/auth/login` | Login mit E-Mail & Passwort (setzt HTTP-Only Cookie) | `{ email, password }` | `200` + User JSON |
| POST | `/api/auth/logout` | Session-Cookie löschen | – | `204` |
| GET | `/api/auth/csrf-token` | Double-Submit Token erzeugen (Cookie + JSON) | – | `{ csrf_token }` |
| GET | `/api/auth/me` | Authentifizierten User zurückgeben | – | User JSON |
| POST | `/api/auth/change-password` | Eigenes Passwort ändern (nur Browser-Session; andere Sitzungen und API-Tokens werden ungültig, Cookie wird neu ausgestellt) | `{ current_password, new_password }` | `200` + User JSON |
| GET | `/api/auth/tokens` | Eigene API-Tokens auflisten (nur Browser-Session) | – | Token-Liste ohne Geheimnis |
| POST | `/api/auth/tokens` | API-Token erstellen (nur Browser-Session, max. 20) | `{ name, expires_in_days? }` | `201` + Token inkl. Klartext (einmalig) |
| DELETE | `/api/auth/tokens/:id` | API-Token widerrufen (nur Browser-Session) | – | `204` |
| POST | `/api/auth/passkeys/login/begin` | Passkey-Anmeldung starten (Rate-Limit, setzt Cookie `passkey_ceremony`) | – | WebAuthn-Optionen `{ publicKey }` |
| POST | `/api/auth/passkeys/login/finish` | Passkey-Anmeldung abschließen (setzt Auth-Cookie) | `{ credential, remember_me }` | `200` + User JSON |
| POST | `/api/auth/passkeys/register/begin` | Passkey hinzufügen starten (nur Browser-Session, max. 20) | `{ current_password }` | WebAuthn-Optionen `{ publicKey }` |
| POST | `/api/auth/passkeys/register/finish` | Passkey speichern | `{ name, credential }` | `201` + Passkey |
| GET | `/api/auth/passkeys` | Eigene Passkeys auflisten (nur Browser-Session) | – | Passkey-Liste ohne Schlüssel |
| PATCH | `/api/auth/passkeys/:id` | Passkey umbenennen | `{ name }` | Passkey |
| DELETE | `/api/auth/passkeys/:id` | Passkey entfernen | – | `204` |

**Authentifizierung per API-Token:** Alle geschützten Endpunkte akzeptieren alternativ zum Cookie den Header `Authorization: Bearer kgm_…`. Das Token hat die Rechte seines Users; ein CSRF-Header ist nicht nötig. Ist der Header gesetzt, wird nur das Token geprüft (kein Rückfall auf das Cookie). Tokens verfallen optional (1–365 Tage) und werden beim Zurücksetzen oder Ändern des Passworts widerrufen. Tokens können nicht mit einem Token verwaltet werden, damit ein geleaktes Token keine weiteren erzeugen kann.

#### Clubs
| Methode | Endpoint | Beschreibung | Request Body | Response | Berechtigung |
|--------|----------|--------------|--------------|---------|--------------|
| GET | `/api/clubs` | Alle Clubs des angemeldeten Users abrufen | – | `200` + Club[] JSON | Authentifiziert |
| GET | `/api/clubs/:id` | Einzelnen Club abrufen | – | `200` + Club JSON | Authentifiziert |
| POST | `/api/clubs` | Neuen Club erstellen | `{ name, balance, base_fee }` | `201` + Club JSON | Authentifiziert (Owner wird automatisch gesetzt) |
| PUT | `/api/clubs/:id` | Club aktualisieren | `{ name, balance, base_fee }` | `200` + Club JSON | Nur Owner |
| DELETE | `/api/clubs/:id` | Club löschen | – | `204` | Nur Owner |


### Fehlerbehandlung
- Error Codes
- Format
- Throttling

---

## 9. Qualität & Anforderungen

### Tests
- Backend: Unit + Integration
  - **Club Repository Tests**: 93.3% Coverage (15 Test-Funktionen)
  - **Server Handler Tests**: 84.2% Coverage (25+ Test-Funktionen für Club-Endpunkte)
  - Abdeckung: Erfolgreiche CRUD-Operationen, Fehlerfälle, Owner-Checks, Validierung
- Frontend: E2E (Cypress/Playwright) - geplant

### Sicherheit
- JWT Sicherheit
- Passwort-Hashing
- Rate Limiting
- CSRF-Schutz (Double-Submit Token) & CORS-Konfiguration
- Persönliche API-Tokens: 256 Bit Zufall, nur Hash in der DB, optionaler Ablauf, Widerruf bei Passwort-Reset
- Passkeys: Hinzufügen nur mit aktuellem Passwort und Browser-Session; Challenges einmalig und 5 Minuten gültig; Signaturzähler-Prüfung gegen geklonte Authenticators; alle Passkeys werden beim Zurücksetzen des Passworts („Passwort vergessen“) gelöscht, beim normalen Passwortwechsel bleiben sie erhalten

### Performance
- API < 200ms
- Pagination, Caching

### Definition of Done
- API implementiert
- Validierung vorhanden
- E2E Tests abgeschlossen
- UX geprüft
- README fertig

---

## 10. Deployment

### Zielumgebung
- Docker
- Hosting: kokal

### CI/CD
- GitHub Actions: Tests → Build

### Infrastruktur & Developer Experience
- `docker-compose.yaml` orchestriert Postgres, Redis, Backend & Frontend
- Persistenz via benannten Volumes (`pg-data`, `redis-data`)
- Makefile stellt Shortcuts für Builds, Tests, Compose und Migrationen bereit

---

## 11. Implementierungsfortschritt

### Abgeschlossen ✅

#### Backend
- ✅ User-Authentifizierung (Register, Login, Logout, Session-Management)
- ✅ CSRF-Schutz mit Double-Submit Token
- ✅ Persönliche API-Tokens (Bearer-Auth für Skripte/Agenten)
- ✅ Account-Seite mit Passwort ändern
- ✅ Passkeys als zusätzliche Anmeldemethode (inkl. Autofill)
- ✅ MCP-Server `/mcp` (Phase 1: 7 Lese-Tools, Spieltag als Entwurf erfassen, veröffentlichen, Buchung erfassen)
- ✅ Club Entity mit vollständigem CRUD
- ✅ Owner-basierte Berechtigungen für Club-Updates/Deletes
- ✅ Unit-Tests für Club-Funktionalität (hohe Coverage)
- ✅ Datenbank-Migrationen (Users, Clubs)

#### Frontend
- ✅ Authentifizierung (Login, Register, Logout)
- ✅ Protected Routes
- ✅ Sidebar mit Navigation
- ✅ Club Switcher mit localStorage-Persistierung
- ✅ Create Club Dialog
- ✅ ClubContext für globales State Management
- ✅ Shadcn UI Komponenten-Integration

### In Arbeit ⏳
- ⏳ Player Entity & CRUD
- ⏳ Game Night Entity & CRUD
- ⏳ Penalty Types Entity & CRUD
- ⏳ Rollen & Berechtigungssystem

### Geplant 📋
- ✅ Statistik-Seite pro Spieler mit wählbarem Zeitraum (siehe unten)
- 📋 Dashboard mit Übersichten
- 📋 E2E Tests

---

## 12. Offene Fragen / Risiken
-
-


---

## Letzte Implementierung: Game Day Feature (24. Januar 2026)

### Implementierte Komponenten

**Backend:**
- ✅ Datenbank-Migration 0007: `game_days`, `game_day_participants`, `game_day_fees` Tabellen
- ✅ SQLc-Queries für alle CRUD-Operationen
- ✅ GameDay-Package mit Entities (GameDay, GameDayParticipant, GameDayFee)
- ✅ Repository mit vollständigen CRUD-Methoden
- ✅ HTTP-Handlers mit Permission-Checks
- ✅ 8 neue API-Endpoints unter `/api/clubs/:clubId/gamedays`
- ✅ Penalty Type Snapshot-Mechanismus (speichert historische Preise/Namen)
- ✅ Integration in Permission-System (EntityTypeGameDays)

**Frontend:**
- ✅ Routes: `/app/gamedays` und `/app/gamedays/:id`
- ✅ Navigation: "Spieltage" Link in Sidebar
- ✅ GameDaysPage: Liste aller Spieltage
- ✅ GameDayDetailPage: Erstellen/Bearbeiten von Spieltagen

**Technische Details:**
- Snapshot-System bewahrt Penalty-Type-Informationen (Name, Beschreibung, Preis) zum Zeitpunkt der Eintragung
- Hard Delete für GameDays (CASCADE zu Participants und Fees)
- sql.NullString für optionale Felder (Notes, Description)
- Mehrere GameDays pro Datum erlaubt

**Offene Punkte:**
- UI für Teilnehmer-Verwaltung (Add/Remove Players)
- UI für Strafen-Eingabe (Grid: Players × Penalty Types)
- Snapshot-Indikator im UI (zeigt historische vs. aktuelle Preise)
- Transaktions-System für Balance-Updates
- Repository-Tests

---

## Letzte Implementierung: Helm-4-Chart (14. September 2026)

### Implementierte Komponenten

- ✅ Chart unter `helm/` (Chart-API v2, Helm 4)
- ✅ Deployment, Service, optionales Ingress/HTTPRoute/HPA
- ✅ Eine ConfigMap für nicht-sensible Env-Vars
- ✅ Ein Secret für `JWT_SECRET`, `DATABASE_URL` (optional Redis/API-Keys); Alternative `secrets.existingSecret`
- ✅ Health-Probes auf `/healthz`
- ✅ README mit Install-Beispiel

Postgres und Redis werden nicht mitdeployt. `AUTO_MIGRATE` ist im Chart-Default aktiv; bei mehreren Replicas können Migrationen konkurrieren.

---

## Letzte Implementierung: Passkeys (18. November 2026)

Passkeys sind eine zusätzliche Anmeldemethode; das Passwort bleibt Pflicht (für „Passwort vergessen“ und das Hinzufügen von Passkeys).

- ✅ Migration `000027_create_webauthn` (Tabellen `webauthn_credentials` und `webauthn_sessions`)
- ✅ Paket `internal/passkey` auf Basis von `github.com/go-webauthn/webauthn` (Go ≥ 1.26, Docker-Images und CI entsprechend angehoben)
- ✅ Nur discoverable Credentials („resident keys“): Anmeldung ohne E-Mail-Eingabe, keine Preisgabe registrierter Adressen
- ✅ Challenge pro Zeremonie in der DB, per `DELETE … RETURNING` einmalig verwendbar; Verweis darauf im HTTP-only-Cookie `passkey_ceremony` (Pfad `/api/auth/passkeys`, 5 Minuten). Funktioniert mit mehreren Replicas.
- ✅ Relying Party aus `PUBLIC_BASE_URL` abgeleitet, überschreibbar mit `WEBAUTHN_RP_ID` / `WEBAUTHN_ORIGINS`
- ✅ Hinzufügen erfordert das aktuelle Passwort; Verwaltung nur mit Browser-Session (nicht per API-Token)
- ✅ Passwort-Reset löscht alle Passkeys; normaler Passwortwechsel behält sie
- ✅ Audit-Events `user_login_passkey`, `passkey_created`, `passkey_renamed`, `passkey_deleted`
- ✅ Frontend: Button „Mit Passkey anmelden“ und Autofill (`autocomplete="username webauthn"`) auf `/login`, Passkey-Verwaltung auf `/app/account`; native Browser-API (`PublicKeyCredential.parse…OptionsFromJSON`)
- ✅ Tests: komplette Zeremonie mit Software-Authenticator (`internal/passkey/service_test.go`), Handler-Tests (`internal/server/passkey_handler_test.go`), E2E mit virtuellem Chromium-Authenticator (`frontend/e2e/passkey.spec.ts`)

### Geplant

- 📋 Passwortlose Konten (Registrierung nur mit Passkey)
- 📋 Unbekannte Passkeys dem Browser melden (`PublicKeyCredential.signalUnknownCredential`), damit gelöschte Passkeys aus dem Autofill verschwinden

---

## Letzte Implementierung: Account-Seite (17. November 2026)

Die Platzhalter „Billing“ und „Notifications“ im Benutzermenü sind entfernt; „Account“ führt zur neuen Seite `/app/account`.

- ✅ Endpunkt `POST /api/auth/change-password` (CSRF, nur Browser-Session, Rate-Limit, aktuelles Passwort erforderlich)
- ✅ Setzt `password_changed_at` → alle anderen Sitzungen ungültig; API-Tokens werden gelöscht; Cookie wird mit gleicher Remember-me-Wahl neu ausgestellt
- ✅ Audit-Event `user_password_changed`
- ✅ User-JSON enthält `password_changed_at`
- ✅ Frontend-Seite `/app/account` (Kontodaten, eigene Clubs, Passwort ändern, Link zu API-Tokens)
- ✅ Handler-Tests in `internal/server/change_password_handler_test.go`, E2E-Test in `frontend/e2e/account.spec.ts`

---

## Letzte Implementierung: Persönliche API-Tokens (17. November 2026)

### Implementierte Komponenten

- ✅ Migration `000026_create_api_tokens` (Tabelle `api_tokens`: Name, SHA-256-Hash, Anzeige-Präfix, `expires_at`, `last_used_at`)
- ✅ Repository `internal/apitoken` (Token-Format `kgm_` + 43 Zeichen base64url)
- ✅ `UserFromCookie` akzeptiert `Authorization: Bearer`; `ValidateCSRF` überspringt Bearer-Requests
- ✅ Endpunkte `GET/POST /api/auth/tokens`, `DELETE /api/auth/tokens/:id` (nur Browser-Session)
- ✅ `last_used_at` höchstens minütlich aktualisiert; Audit-Events `api_token_created` / `api_token_deleted`
- ✅ Frontend-Seite `/app/api-tokens` (Benutzermenü)
- ✅ OpenAPI: Security-Scheme `bearerAuth` für alle geschützten Endpunkte
- ✅ Handler-Tests in `internal/server/api_token_handler_test.go`

---

## Letzte Implementierung: MCP-Server Phase 1 (17. November 2026)

KI-Agenten (Claude Code, Claude Desktop, Agent-SDKs) können Kegelmaster über das Model Context Protocol bedienen.

### Implementierte Komponenten

- ✅ Paket `internal/mcpserver` mit offiziellem Go-SDK (`github.com/modelcontextprotocol/go-sdk`), eingehängt unter `/mcp` (Fiber-Adaptor)
- ✅ Auth: nur persönliche API-Tokens (`Authorization: Bearer kgm_…`), Prüfung über `GET /api/auth/me`; 401 mit `WWW-Authenticate: Bearer`
- ✅ Tools rufen die REST-API in-process (direkt im Fiber-Handler, ohne Netzwerk) mit dem Token des Aufrufers auf – Rechte, Validierung, Audit-Log und Saldo-Invarianten bleiben identisch
- ✅ Lese-Tools: `klubs_auflisten`, `klub_ueberblick`, `spieltage_auflisten`, `spieltag_details` (inkl. Buchungsvorschau für Entwürfe), `spieler_konto`, `transaktionen_suchen`, `mein_spieler`
- ✅ Schreib-Tools: `spieltag_erfassen` (immer Entwurf, bei Fehler wird der Entwurf verworfen), `spieltag_veroeffentlichen` (bucht Grundgebühren und Strafen, `destructiveHint`), `buchung_erfassen` (Einzahlung/Trinkgeld/Ausgabe, `destructiveHint`)
- ✅ Beträge in Cent mit Euro-Text, IDs plus Namen in allen Ausgaben, deutsche Fehlermeldungen der API werden durchgereicht
- ✅ Tests: `internal/mcpserver/server_test.go` (Fake-API), `internal/server/mcp_handler_test.go` (Auth über Fiber)

### Einrichtung (Beispiel Claude Code)

```bash
claude mcp add --transport http kegelmaster https://<host>/mcp --header "Authorization: Bearer kgm_…"
```

### Nebenbei behoben

- `PUT …/participants/:playerId/fees` lehnt Mengen jetzt mit 400 ab, statt sie still zu runden (1,5 bei Strafentypen ohne Dezimalmengen, mehr als zwei Nachkommastellen)

### Geplant (Phase 2)

- 📋 `spieltag_teilnehmer_aktualisieren`, `spieltag_entwurf_loeschen`, `buchung_stornieren`
- 📋 Stammdaten: `spieler_speichern`, `strafentyp_speichern`
- 📋 API-Tokens mit Scope (z. B. nur lesen)
- 📋 OAuth für Custom Connectors auf claude.ai

## Letzte Implementierung: Statistik-Seite (18. November 2026)

### Implementierte Komponenten

- **Backend:** Paket `internal/statistics` (Queries in `queries/statistics.sql`, reine Berechnung in `Compute`, Unit-Tests). Endpunkt `GET /api/clubs/:clubId/players/me/statistics?from=&to=` (beide optional; ohne `from` = gesamter Zeitraum, `to` = heute). Berücksichtigt nur veröffentlichte Spieltage.
- **Inhalte:** Teilnahmequote und Serien, Strafen gesamt / Ø pro Spieltag mit Klubvergleich, Platz im Klub (Ø-Strafen), strafenfreie Abende, teuerster/günstigster Abend, Strafen nach Typ (eigener Ø vs. Klub-Ø), Wettbewerbe (Ø, Bestwert, Siege, letzte Plätze, Platz nach Ø, Verlauf), Kontoverlauf mit Einzahlungen/Strafen/Grundgebühren.
- **MCP:** Tool `meine_statistik` (von/bis).
- **Frontend:** Route `/app/statistics` (Sidebar „Statistiken“), Zeitraum-Auswahl (Presets + benutzerdefiniert) im URL-Query gespeichert.
