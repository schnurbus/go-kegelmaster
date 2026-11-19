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

Getrennt davon gibt es die Rolle **Instanz-Admin** (Betreiber, `users.is_instance_admin`). Sie gilt nicht für einen Klub, sondern für die Instanz, und gibt Zugriff auf die Betreiber-Übersicht (`/api/admin`) sowie auf gezielte Eingriffe (Benutzer sperren, Sessions/API-Tokens widerrufen, Klub-Owner übertragen, Benutzer/Klubs löschen). Ein Instanz-Admin wird dadurch weder Owner noch Mitglied eines Klubs und sieht keine Fachdaten der Klubs.

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
- Elemente: Registrierungs-Formular (Email, Passwort); bei `REGISTRATION_MODE=closed` nur ein Hinweis, bei `invite_only` nur mit Einladungslink (E-Mail aus der Einladung vorausgefüllt)
- Interaktionen: Registrierung, automatischer Login, Weiterleitung zu Dashboard

**Screen:** Dashboard
- URL: /app
- Elemente: Sidebar mit Club Switcher, Navigation, Hauptinhalt
- Interaktionen: Club-Auswahl, Navigation zwischen Bereichen

**Screen:** Mein Account
- URL: /app/account (Benutzermenü unten links → „Account“)
- Elemente: Kontodaten (E-Mail, registriert am, Passwort zuletzt geändert), Liste der eigenen Clubs (Eigentümer/Mitglied), Benachrichtigungen (Schalter „E-Mail, wenn ein Spieltag abgerechnet wurde“, standardmäßig aus), Formular „Passwort ändern“ (aktuelles, neues, Bestätigung), Passkeys (Liste mit Name, synchronisiert/gerätegebunden, hinzugefügt, zuletzt genutzt), Link zu den API-Tokens
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
| POST | `/api/auth/register` | Neuen User anlegen & automatisch einloggen (abhängig von `REGISTRATION_MODE`) | `{ email, password, invite_token? }` | `201` + User JSON, `403` wenn nicht erlaubt |
| POST | `/api/auth/login` | Login mit E-Mail & Passwort (setzt HTTP-Only Cookie) | `{ email, password }` | `200` + User JSON |
| POST | `/api/auth/logout` | Session-Cookie löschen | – | `204` |
| GET | `/api/auth/csrf-token` | Double-Submit Token erzeugen (Cookie + JSON) | – | `{ csrf_token }` |
| GET | `/api/auth/me` | Authentifizierten User zurückgeben | – | User JSON |
| PATCH | `/api/auth/me` | Eigene Einstellungen ändern (fehlende Felder bleiben unverändert) | `{ notify_gameday_settled }` | `200` + User JSON |
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
- Instanz-Admin-Bereich `/api/admin`: nur Browser-Session (kein API-Token, kein MCP); alle anderen erhalten 404; Vergabe der Rolle nur per CLI bzw. `INSTANCE_ADMIN_EMAILS`; jeder Zugriff wird mit `admin=true` und Client-IP über `slog` auditiert; Eingriffe zusätzlich mit CSRF-Schutz und Bestätigungsdialog
- Gesperrte Konten (`users.disabled_at`): Login (Passwort/Passkey), Browser-Session und API-Tokens werden abgelehnt; Sessions lassen sich per `users.sessions_valid_after` widerrufen
- Admin-Bereich nur mit frischer Anmeldung: Liegt der letzte Passwort- oder Passkey-Login länger als `ADMIN_REAUTH_MINUTES` (Default 15) zurück, muss der Instanz-Admin seine Anmeldung bestätigen (Claim `auth_time` im Session-Cookie)
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
- Jeder Push auf `main` baut `ghcr.io/schnurbus/go-kegelmaster` mit den Tags `<sha>` und `main-<YYYYMMDDHHmmss>-<sha>` (UTC, sortierbar); Releases (Release Please) zusätzlich mit SemVer-Tags.

### Umgebungen
- **Produktion** `https://kegelmaster.schnurbus.net`: Kubernetes (FluxCD, Repo `talos-fluxcd`, `clusters/talos/apps/kegelmaster`), folgt nur SemVer-Releases.
- **Test** `https://kegelmaster-test.schnurbus.net`: gleiches Helm-Chart in Namespace `kegelmaster-test` (`clusters/talos/apps/kegelmaster-test`), eigene Datenbank. Flux' Image-Automation übernimmt automatisch das neueste `main-*`-Image, d. h. jeder Merge auf `main` ist nach wenigen Minuten testbar – ohne Release. Kein E-Mail-Versand (kein `RESEND_API_KEY`), Registrierung `invite_only`.

### Infrastruktur & Developer Experience
- `docker-compose.yaml` orchestriert Postgres, Redis, Backend & Frontend
- Persistenz via benannten Volumes (`pg-data`, `redis-data`)
- Makefile stellt Shortcuts für Builds, Tests, Compose und Migrationen bereit
- Renovate (`renovate.json`) hält Go-, npm-, Docker- und GitHub-Actions-Abhängigkeiten aktuell (wöchentliche, gebündelte PRs; Sicherheits-Updates sofort). Erfordert die installierte Renovate GitHub App (Details in `AGENTS.md`).

---

## 11. Implementierungsfortschritt

### Abgeschlossen ✅

#### Backend
- ✅ User-Authentifizierung (Register, Login, Logout, Session-Management)
- ✅ CSRF-Schutz mit Double-Submit Token
- ✅ Persönliche API-Tokens (Bearer-Auth für Skripte/Agenten)
- ✅ Account-Seite mit Passwort ändern
- ✅ Registrierung per `REGISTRATION_MODE` steuerbar (`open`/`invite_only`/`closed`)
- ✅ Passkeys als zusätzliche Anmeldemethode (inkl. Autofill)
- ✅ E-Mail-Benachrichtigung „Spieltag abgerechnet“ (Opt-in in den Account-Einstellungen)
- ✅ Instanz-Admin und read-only Admin-API `/api/admin` (Betreiber-Übersicht Phase 1, Backend)
- ✅ Erneute Anmeldung für den Admin-Bereich (Re-Auth per Passwort oder Passkey nach `ADMIN_REAUTH_MINUTES`)
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
- ✅ Passwortfelder mit Augen-Symbol zum Anzeigen/Verbergen der Eingabe (`components/ui/password-input.tsx`, in allen Passwortfeldern)
- ✅ Administrationsbereich für Betreiber `/admin` (Dashboard, Benutzer, Klubs, CSV-Export; nur für Instanz-Admins)
- ✅ Admin-Eingriffe: Benutzer sperren/entsperren, Sessions und API-Tokens widerrufen, Klub-Owner übertragen, Benutzer/Klubs löschen

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

## Letzte Implementierung: E-Mail-Benachrichtigung „Spieltag abgerechnet“ (19. November 2026)

User können sich per E-Mail benachrichtigen lassen, sobald ein neuer Spieltag abgerechnet wurde (KEG-4).

- **Einstellung:** `users.notify_gameday_settled` (Migration `000029`, Standard `false` für bestehende und neue User). Änderbar über `PATCH /api/auth/me` mit `{ "notify_gameday_settled": bool }`, im Frontend auf `/app/account` in der Karte „Benachrichtigungen“. Audit-Eintrag `user_settings_updated`.
- **Auslöser:** Wechsel eines Spieltags von Entwurf auf veröffentlicht (oder direkt veröffentlicht angelegt) – also genau dann, wenn Grundgebühren und Strafen gebucht werden. Das Bearbeiten eines bereits veröffentlichten Spieltags verschickt nichts. Auch über MCP (`spieltag_veroeffentlichen`) veröffentlichte Spieltage lösen die E-Mail aus.
- **Nur einmal pro Spieltag:** Tabelle `game_day_notifications` merkt sich benachrichtigte Spieltage (`INSERT … ON CONFLICT DO NOTHING`). Wird ein Spieltag wieder zum Entwurf gemacht und erneut veröffentlicht, gibt es keine zweite E-Mail. Bereits veröffentlichte Spieltage werden bei der Migration als benachrichtigt eingetragen.
- **Empfänger:** Alle User mit aktivierter Einstellung, die einen Spieler im Klub haben; inaktive Spieler nur, wenn sie an diesem Spieltag Buchungen haben. Hat ein User mehrere Spieler im Klub, bekommt er eine E-Mail mit allen.
- **Inhalt (deutsch):** Klubname, Datum, je Spieler die Summe der Buchungen des Spieltags und der aktuelle Saldo, Link zur App (`PUBLIC_BASE_URL/app`).
- **Versand:** Paket `internal/notification`, asynchron nach dem Commit der Buchungen (eigene Goroutine, 2 Minuten Timeout); Fehler werden nur geloggt und blockieren oder verwerfen die Abrechnung nie. Ohne `RESEND_API_KEY` wird mit einer Log-Zeile übersprungen. Beim Herunterfahren wartet der Server kurz auf laufende Benachrichtigungen.
- **Keine neuen Umgebungsvariablen:** genutzt werden `RESEND_API_KEY`, `RESEND_FROM_EMAIL`, `PUBLIC_BASE_URL`.

## Letzte Implementierung: Transaktionen bearbeiten (19. November 2026)

Fehleingaben bei manuellen Buchungen lassen sich korrigieren, ohne die Transaktion zu löschen und neu anzulegen (KEG-3).

- **API:** `PUT /api/clubs/:clubId/transactions/:id` mit `transaction_type` (deposit/tip/expense), `player_id` (Pflicht bei deposit/tip), `amount` (Cent, positiv), `description`, `transaction_date` (optional, leer = unverändert). Benötigt die Berechtigung `transactions/update` (Rollen-Berechtigungstabelle, Spalte „Bearbeiten“); Audit-Eintrag `transaction_updated`.
- **Bearbeitbar:** nur manuelle Buchungen (Einzahlung, Trinkgeld, Ausgabe) – auch Auto-Tip-Einträge und Paar-Modus-Teilbuchungen, da jede Zeile für sich im Ledger steht. Grund- und Strafgebühren (vom Spieltag erzeugt) werden mit 400 abgelehnt; sie werden über den Spieltag korrigiert, sonst laufen Strafzeilen und Ledger auseinander.
- **Salden:** In einer DB-Transaktion wird die Buchung gesperrt, die betroffenen Spieler (alt und neu, in ID-Reihenfolge) gesperrt, die Wirkung der alten Buchung zurückgenommen und die neue angewendet; der Kassenstand ändert sich atomar um die Differenz. Damit bleibt Saldo = Startsaldo + Summe der Buchungen (wie bei recalculate-balance). Die Saldo-Snapshots der bearbeiteten Zeile werden neu gesetzt (als ob jetzt gebucht), andere Zeilen bleiben unverändert (wie beim Löschen).
- **Keine Neuaufteilung:** Beim Bearbeiten gibt es keine Auto-Tip-Aufteilung und keinen Paar-Modus; der Betrag wird genau so gebucht.
- **Frontend:** Auf der Transaktionsseite Stift-Symbol für manuelle Buchungen (nur mit Bearbeiten-Berechtigung), Dialog `EditTransactionDialog`.
- **Hinweis:** Bestehende Rollen erhalten die Bearbeiten-Berechtigung nicht automatisch; der Klub-Owner kann sie in der Rollenverwaltung setzen.

## Letzte Implementierung: Abwesende Teilnehmer (18. November 2026)

Sonderfall: Ein Spieler fehlt am Spieltag, bekommt aber trotzdem eine Strafe (z. B. „Puppe vergessen“).

- **Datenmodell:** `game_day_participants.attended` (Migration `000028`, Standard `true`). Bestehende Einträge bleiben anwesend.
- **API:** `PATCH /api/clubs/:clubId/gamedays/:id/participants/:playerId` mit `{ "attended": bool }`. Beim Markieren als abwesend werden Wettbewerbs-Werte gelöscht; für Abwesende werden keine Wettbewerbs-Werte angenommen. Strafen und Buchungen bleiben unverändert.
- **Frontend:** In der Strafen-Tabelle des Spieltags hat jeder Teilnehmer einen Haken „anwesend“ (standardmäßig gesetzt). Wer abwesend ist, muss aktiv abgehakt werden. Abwesende sind in den Tabellen als „abwesend“ markiert, ihre Wettbewerbsfelder sind gesperrt.
- **Statistik:** Abwesende zählen nicht zur Teilnahme (Quote, Serien), nicht zu Wettbewerben und nicht zur Teilnehmerzahl in der Spieltagsübersicht. Ihre Strafen zählen zu den Strafsummen (`absent_amount` zeigt den Anteil); Durchschnitte und Ranglisten teilen durch die anwesenden Spieltage.
- **MCP:** Spieltag-Details enthalten `abwesend: true` für abwesende Teilnehmer.
- **Spieltage ohne Teilnehmer:** Spieltage ohne anwesende Teilnehmer (z. B. nur zum Einfordern der Grundgebühr) fließen nicht in die Statistik ein (keine leeren Abende in Teilnahmequote, Serien, Diagrammen und Durchschnitten). Die Grundgebühr bleibt im Kontoverlauf enthalten.

## Letzte Implementierung: Statistik-Seite (18. November 2026)

### Implementierte Komponenten

- **Backend:** Paket `internal/statistics` (Queries in `queries/statistics.sql`, reine Berechnung in `Compute`, Unit-Tests). Endpunkt `GET /api/clubs/:clubId/players/me/statistics?from=&to=` (beide optional; ohne `from` = gesamter Zeitraum, `to` = heute). Berücksichtigt nur veröffentlichte Spieltage.
- **Inhalte:** Teilnahmequote und Serien, Strafen gesamt / Ø pro Spieltag mit Klubvergleich, Platz im Klub (Ø-Strafen), strafenfreie Abende, teuerster/günstigster Abend, Strafen nach Typ (eigener Ø vs. Klub-Ø), Wettbewerbe (Ø, Bestwert, Siege, letzte Plätze, Platz nach Ø, Verlauf), Kontoverlauf mit Einzahlungen/Strafen/Grundgebühren.
- **MCP:** Tool `meine_statistik` (von/bis).
- **Frontend:** Route `/app/statistics` (Sidebar „Statistiken“), Zeitraum-Auswahl (Presets + benutzerdefiniert) im URL-Query gespeichert.

## Letzte Implementierung: Registrierung per REGISTRATION_MODE (19. November 2026)

Phase 0 der Betreiber-Übersicht (KEG-6, Konzept `docs/konzept-betreiber-uebersicht.md` §7).

- **Konfiguration:** `REGISTRATION_MODE=open|invite_only|closed`, Default `open` (unverändertes Verhalten). Ungültige Werte verhindern den Start. Ausgeliefert als `registration_mode` unter `GET /api/config/public`.
- **API:** `POST /api/auth/register` antwortet bei `closed` immer mit 403. Bei `invite_only` ist `invite_token` Pflicht: Die Einladung muss existieren, darf nicht abgelaufen oder angenommen sein, und die E-Mail muss passen (sonst 403). Die Einladung wird danach wie bisher über `POST /api/invitations/:token/accept` angenommen. Login, Passkeys (nur für bestehende Konten) und das Annehmen von Einladungen mit bestehendem Konto bleiben in allen Modi möglich. Keine Schemaänderung.
- **Frontend:** Hook `usePublicConfig`. Registrierungslinks (Header, Landingpage, Login) nur bei `open`; die Landingpage verweist sonst auf den Login. `/register` zeigt bei `closed` einen Hinweis, bei `invite_only` ohne Einladungslink ebenfalls; mit Einladungslink ist die E-Mail aus der Einladung vorausgefüllt und gesperrt. Die Einladungsseite blendet „Registrieren“ bei `closed` aus.
- **Tests:** `internal/server/registration_mode_handler_test.go` (alle drei Modi, Default ohne Variable, abgelaufene/angenommene/unbekannte Einladung, falsche E-Mail, `config/public`), `internal/config/registration_test.go`.

## Letzte Implementierung: Auto-Tip pro Einzahlung abschaltbar (19. November 2026)

Auch wenn ein Klub Auto-Tip aktiviert hat, kann es für eine einzelne Einzahlung abgeschaltet werden (KEG-2).

- **API:** `POST /api/clubs/:clubId/transactions` akzeptiert optional `"skip_auto_tip": true` (nur bei `deposit` wirksam). Dann wird der volle Betrag als Einzahlung gebucht, ein Überschuss bleibt als Guthaben beim Spieler; im Paar-Modus wird der Überschuss wie bei deaktiviertem Auto-Tip auf die Spieler verteilt. Standardverhalten unverändert, keine Schemaänderung.
- **Frontend:** Im Transaktionsdialog erscheint bei Einzahlungen (nur wenn der Klub Auto-Tip aktiviert hat) der Haken „Auto-Tip für diese Einzahlung anwenden“ (standardmäßig gesetzt). Abgehakt entfällt die Auto-Tip-Vorschau.
- **MCP:** `buchung_erfassen` hat den optionalen Parameter `ohne_auto_trinkgeld` (nur für `einzahlung`).

## Letzte Implementierung: Instanz-Admin und Admin-API, read-only (19. November 2026)

Phase 1 (Backend) der Betreiber-Übersicht (KEG-7, Konzept `docs/konzept-betreiber-uebersicht.md`). Die Admin-UI ist im nächsten Abschnitt beschrieben (KEG-8).

- **Datenmodell:** Migration `000030_instance_admin`: `users.is_instance_admin` (Standard `false`), `users.last_login_at` (NULL = unbekannt, Bestandskonten bis zum nächsten Login), Index auf `users.created_at`. `last_login_at` wird bei Passwort- und Passkey-Login gesetzt (Fehler werden nur geloggt, der Login läuft weiter).
- **Rolle vergeben:** nur per CLI im Haupt-Binary (`kegelmaster admin grant <email>`, `admin revoke <email>`, `admin list`; im Container per `kubectl exec`/`docker compose exec`) oder per `INSTANCE_ADMIN_EMAILS` (kommagetrennt, setzt beim Start das Flag für bestehende Konten, entzieht nie). **Keine** Vergabe über die Weboberfläche.
- **API** (`internal/admin`, Queries in `queries/admin.sql`, Handler in `handlers/admin.go`): `GET /api/admin/stats`, `/users` (Suche, Filter `orphaned`/`admin`, Sortierung, Paginierung), `/users/:id`, `/clubs` (Suche, Filter `inactive`, Sortierung, Paginierung), `/clubs/:id`, `/export/memberships.csv`. `GET /api/auth/me` liefert zusätzlich `is_instance_admin`. Keine Fachdaten der Klubs (Salden, Transaktionen, Strafen), nur aggregierte Anzahlen.
- **Definitionen:** verwaist = ohne eigenen Klub und ohne Spieler-Verknüpfung, älter als 30 Tage; aktiver Klub = Spieltag oder Transaktion in den letzten 90 Tagen; inaktiver Klub = keine Aktivität seit 180 Tagen; Registrierungsspitze = mehr als 10 Registrierungen an einem Tag (letzte 30 Tage); Tokens ungenutzt = aktiv, aber seit 90 Tagen nicht benutzt.
- **Sicherheit:** Middleware `RequireInstanceAdmin` (unabhängig vom Klub-`PermissionChecker`): nur Cookie-Session; Anfragen mit `Authorization: Bearer` (API-Token, MCP), ohne Anmeldung oder ohne Admin-Flag erhalten **404** wie unbekannte Routen. Antworten mit `Cache-Control: no-store`. CSV-Export mit Schutz vor CSV-Injection.
- **Audit:** nur über `slog` (keine DB-Tabelle): `admin_stats_viewed`, `admin_users_listed`, `admin_user_viewed`, `admin_clubs_listed`, `admin_club_viewed`, `admin_export` mit `admin=true`; CLI-Änderungen als `admin_role_granted`/`admin_role_revoked`. Alle Audit-Einträge enthalten jetzt die Client-IP (`ip`, berücksichtigt `PROXY_HEADER`/`TRUSTED_PROXIES`).
- **Datenschutz:** Die Datenschutzerklärung der Instanz muss ergänzt werden: Der Betreiber sieht Konto-Metadaten (E-Mail, Registrierungs- und Login-Zeitpunkte, Klubzugehörigkeiten, Anzahl/Namen von API-Tokens und Passkeys) zum Betrieb und zur Missbrauchsabwehr (Art. 6 Abs. 1 lit. f DSGVO); `last_login_at` wird mit dem Konto gelöscht; Aufbewahrung der Audit-Logs laut Log-Stack.
- **Tests:** `internal/server/admin_handler_test.go` (ohne Anmeldung/Nicht-Admin/Bearer-Token → 404, Admin → 200 für alle Endpunkte, ungültige Parameter → 400), `internal/admin` (Parameter, CLI, Env-Bootstrap), Audit (`admin=true`, IP).

## Letzte Implementierung: Administrationsbereich für Betreiber (19. November 2026)

Phase 1 (Frontend) der Betreiber-Übersicht (KEG-8, Konzept `docs/konzept-betreiber-uebersicht.md` §4, §9), baut auf der Admin-API (KEG-7) auf. Rein lesend, keine Backend-Änderungen.

- **Seiten** (`frontend/src/routes/admin/`, Layout `DashboardLayout` + `AdminLayout`): `/admin` (Kennzahlen-Kacheln, Liste „Auffälligkeiten“ mit Links auf die gefilterten Listen), `/admin/users` (E-Mail, registriert, letzter Login, Owner von / Spieler in n Klubs, Passkeys/Tokens, Status; Suche, Filter verwaist/Admin, Sortierung per Spaltenkopf, serverseitige Paginierung), `/admin/users/:id` (Stammdaten, eigene Klubs, Spieler-Verknüpfungen mit Rolle, API-Tokens, Passkeys), `/admin/clubs` (Name, Owner, Spieler gesamt/verknüpft, angelegt, letzte Aktivität, Spieltage; Suche, Filter inaktiv), `/admin/clubs/:id` (Owner, Spieler mit verknüpftem Benutzer bzw. „nicht verknüpft“, Rolle, offene Einladungen, Rollen mit Anzahl). Benutzer und Klubs sind gegenseitig verlinkt. Such-, Filter-, Sortier- und Seitenzustand stehen in der URL (verlinkbar, z. B. `/admin/users?filter=orphaned`).
- **CSV-Export:** Button „Verknüpfungen als CSV“ auf allen Admin-Seiten (`/api/admin/export/memberships.csv`).
- **Navigation:** Menüeintrag „Administration“ in der Sidebar nur bei `is_instance_admin` aus `/api/auth/me`, unabhängig vom ausgewählten Klub; der Bereich liest den `ClubContext` nicht. Nicht-Admins sehen auf `/admin*` die Fehlerseite „Seite nicht gefunden“ (ohne Admin-Request), unbekannte IDs ebenso (API 404).
- **Komponenten:** Shadcn `Table`, `Badge` (Admin, verwaist, Owner, verknüpft / nicht verknüpft, inaktiv), `Card`, `Select`; gemeinsame Bausteine in `components/admin/admin-ui.tsx`, Daten-Hook `useAdminFetch` und URL-Listenzustand `useAdminListParams` in `lib/admin.ts`, Typen in `types/admin.ts`. Mobil: Listen als Kartenliste statt Tabelle, Detailtabellen horizontal scrollbar.
- **Tests:** Playwright `frontend/e2e/admin.spec.ts`: Admin (Rolle per CLI `go run ./cmd/api admin grant`) sieht Benutzer und Klubs inkl. Querverweisen und CSV; normaler Benutzer sieht keinen Menüeintrag und bekommt auf `/admin` die Fehlerseite.

## Letzte Implementierung: Admin-Eingriffe (19. November 2026)

Phase 3 der Betreiber-Übersicht (KEG-10, Konzept `docs/konzept-betreiber-uebersicht.md` §7). Baut auf Instanz-Admin, Admin-API und Admin-UI auf.

- **Datenmodell:** Migration `000031_admin_actions`: `users.disabled_at` (gesperrt seit, NULL = aktiv) und `users.sessions_valid_after` (Browser-Sessions, die davor ausgestellt wurden, sind ungültig). In Produktion läuft die Migration über `AUTO_MIGRATE`.
- **Sperren:** `POST /api/admin/users/:id/disable|enable`. Gesperrte Konten: Passwort- und Passkey-Login → 403 „Dieses Konto ist gesperrt …“ (erst nach korrektem Passwort, damit gesperrte Konten nicht ausgekundschaftet werden können), `UserFromCookie` und `userFromAPIToken` lehnen ab (401); damit verliert auch ein gesperrter Instanz-Admin sofort den Admin-Zugriff. Beim Sperren werden die Browser-Sessions mit widerrufen, damit ein späteres Entsperren sie nicht wiederbelebt; API-Tokens gelten nach dem Entsperren wieder.
- **Sessions widerrufen:** `POST /api/admin/users/:id/revoke-sessions` setzt `sessions_valid_after`; `User.SessionValid` prüft es zusätzlich zu `password_changed_at`. API-Tokens sind davon nicht betroffen und werden einzeln widerrufen: `DELETE /api/admin/users/:id/tokens/:tokenId`.
- **Owner übertragen:** `POST /api/admin/clubs/:id/transfer-owner` (`new_owner_email`), gleiche Logik wie `HandleTransferClubOwner` (gemeinsame Funktion `transferClubOwner`), aber ohne Owner- und Passwortprüfung. Gesperrte Konten können nicht (mehr) Owner werden – auch nicht bei der normalen Übertragung durch den Owner. Der Admin wird nicht Mitglied des Klubs.
- **Löschen:** `DELETE /api/admin/users/:id` mit `confirm_email`, `DELETE /api/admin/clubs/:id` mit `confirm_name` (exakter Klubname). Ein Benutzer, der noch Owner ist, wird mit 409 abgelehnt (atomar in der Delete-Query geprüft), sonst würde die Kaskade seine Klubs löschen. Verknüpfte Spieler bleiben in ihren Klubs (`players.user_id` → NULL), Salden und Buchungen ändern sich nicht. Beim Klub-Löschen kaskadieren alle Fachdaten des Klubs.
- **Schutzregeln:** Das eigene Konto kann nicht gesperrt, gelöscht oder per Sessions-Widerruf abgemeldet werden (400). Andere Instanz-Admins können über die Weboberfläche weder gesperrt noch gelöscht werden (409): wie die Vergabe bleibt das Entziehen der Rolle der CLI vorbehalten (`kegelmaster admin revoke`). So kann ein kompromittiertes Admin-Konto keine anderen Admins aussperren, und der letzte Admin kann nie über die Weboberfläche entfernt werden.
- **Audit:** `admin_user_disabled`, `admin_user_enabled`, `admin_sessions_revoked`, `admin_token_revoked`, `admin_club_owner_transferred`, `admin_user_deleted`, `admin_club_deleted` mit `admin=true`; abgelehnte Logins gesperrter Konten als `user_login_rejected`.
- **Übersicht:** Benutzerliste mit Filter „Gesperrt“ und Badge „gesperrt“, Dashboard zählt gesperrte Konten (`users.disabled` in `/api/admin/stats`), Benutzer-Detail zeigt „Gesperrt seit“ und „Sessions widerrufen“.
- **Frontend:** Karte „Eingriffe“ auf `/admin/users/:id` (Sperren/Entsperren, Sessions widerrufen, Benutzer löschen mit Eingabe der E-Mail; Löschen erst, wenn keine eigenen Klubs mehr bestehen) und `/admin/clubs/:id` (Owner übertragen, Klub löschen mit Eingabe des Namens); „Widerrufen“ je API-Token. Alle Eingriffe mit Bestätigungsdialog (`components/admin/admin-action-dialog.tsx`), Requests über `adminMutation` (CSRF-Header) in `lib/admin.ts`.
- **Tests:** `internal/server/admin_actions_handler_test.go` (gesperrter Benutzer: Login 403, Cookie-Session und API-Token 401; widerrufene Session 401; gesperrter Admin → 404; alle Eingriffe: Nicht-Admin/Bearer → 404, ohne CSRF → 403, Erfolg, eigenes Konto, anderer Admin, falsche Bestätigung, Owner muss erst übertragen, gesperrter neuer Owner), `internal/user` (`SessionValid` mit `sessions_valid_after`), Playwright `frontend/e2e/admin.spec.ts` (Sperren: Session ungültig, Login mit Hinweis abgelehnt; Entsperren: Login wieder möglich).

## Letzte Implementierung: Erneute Anmeldung für den Admin-Bereich (19. November 2026)

Phase 2 der Betreiber-Übersicht (KEG-9, Konzept `docs/konzept-betreiber-uebersicht.md` §3 „Absicherung“).

- **Entscheidung:** „Immer frische Session“. Der Admin-Bereich ist nur nutzbar, wenn die letzte echte Anmeldung (Passwort- oder Passkey-Login bzw. eine erneute Bestätigung) jünger als `ADMIN_REAUTH_MINUTES` ist – unabhängig davon, ob das Konto einen Passkey hat. Ein Passkey zählt als Bestätigung genauso wie das Passwort. Keine Passkey-Pflicht.
- **Session:** Das Session-JWT enthält den Claim `auth_time`. Login, Registrierung, Passwortwechsel und erneute Bestätigung setzen ihn auf „jetzt“; ein künftiges Verlängern der Session ohne Anmeldung muss ihn übernehmen (`auth.Service.GenerateTokenWithAuthTime`). Cookies ohne `auth_time` (vor dieser Änderung ausgestellt) gelten als veraltet. Kein Datenbank-Schema geändert.
- **Konfiguration:** `ADMIN_REAUTH_MINUTES` (Ganzzahl 1–1440, Default 15); ungültige Werte verhindern den Start. In Helm `config.ADMIN_REAUTH_MINUTES`.
- **API:** `RequireInstanceAdmin` antwortet einem Instanz-Admin mit veralteter Anmeldung mit `401 {"error":"reauth_required","message":…}`; Nicht-Admins, Anfragen ohne Anmeldung und Bearer-Tokens erhalten weiterhin 404. Bestätigen: `POST /api/auth/reauth` (`password`) bzw. `POST /api/auth/passkeys/reauth` (Zeremonie über `POST /api/auth/passkeys/login/begin`; der Passkey muss zum angemeldeten Konto gehören). Beide nur mit Browser-Session (Bearer → 403), CSRF-geschützt, Rate-Limit wie Login; sie stellen das Cookie für den aktuellen Benutzer neu aus (Remember-Me bleibt, neues CSRF-Cookie).
- **Audit:** `user_reauth` und `user_reauth_failed` (mit `method` = `password`/`passkey`), `admin_reauth_required` (mit `admin=true`).
- **Frontend:** `useAdminFetch` meldet `reauth_required` an `lib/admin-reauth.ts`; der Dialog „Anmeldung bestätigen“ (`components/admin/admin-reauth-dialog.tsx`, in `AdminLayout`) fragt das Passwort ab bzw. bietet „Mit Passkey bestätigen“, wenn der Browser Passkeys unterstützt. Danach lädt die Admin-Seite ihre Daten neu; „Abbrechen“ führt zum Dashboard.
- **Tests:** `internal/server/admin_reauth_handler_test.go` (frische, veraltete, fehlende und zukünftige `auth_time`, konfiguriertes Zeitfenster, Nicht-Admin/Bearer weiterhin 404, Re-Auth per Passwort: Erfolg, falsches Passwort, ohne Anmeldung/CSRF, Bearer; per Passkey mit Software-Authenticator: Erfolg, Passkey eines anderen Kontos, ohne Zeremonie), `internal/auth` (`auth_time` gesetzt bzw. beim Neuausstellen übernommen), `internal/config` (`ADMIN_REAUTH_MINUTES`), Playwright `frontend/e2e/admin-reauth.spec.ts` (veraltete Session → Dialog → falsches/richtiges Passwort bzw. virtueller Passkey → Admin-Bereich).
