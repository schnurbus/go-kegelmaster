# Kegelmaster Helm Chart

Helm-4-Chart zum Deployen der Kegelmaster-App (Go-API mit eingebettetem Frontend) auf Kubernetes.

Postgres und Redis werden **nicht** mitinstalliert. Die App erwartet ein bestehendes PostgreSQL; Redis ist optional.

Das Chart nutzt Chart-API **v2** (stabil unter Helm 4). Charts v3 sind experimentell und werden nicht verwendet.

## Voraussetzungen

- Helm 4
- Kubernetes >= 1.28
- Externes PostgreSQL
- Container-Image `ghcr.io/schnurbus/go-kegelmaster` (ggf. `imagePullSecrets` für GHCR)

## Installation

Pflicht-Secrets: `JWT_SECRET` und `DATABASE_URL`. In Production kein Default-Secret verwenden.

```bash
helm upgrade --install kegelmaster ./helm \
  --namespace kegelmaster --create-namespace \
  --set secrets.JWT_SECRET='change-me' \
  --set secrets.DATABASE_URL='postgres://user:pass@postgres:5432/kegelmaster?sslmode=disable' \
  --set config.PUBLIC_BASE_URL='https://kegelmaster.example.com' \
  --set config.CORS_ALLOW_ORIGINS='https://kegelmaster.example.com'
```

Oder ein vorhandenes Secret referenzieren (gleiche Keys: mindestens `JWT_SECRET`, `DATABASE_URL`):

```bash
kubectl create secret generic kegelmaster-secrets \
  --namespace kegelmaster \
  --from-literal=JWT_SECRET='change-me' \
  --from-literal=DATABASE_URL='postgres://user:pass@postgres:5432/kegelmaster?sslmode=disable'

helm upgrade --install kegelmaster ./helm \
  --namespace kegelmaster --create-namespace \
  --set secrets.existingSecret=kegelmaster-secrets
```

## Konfiguration

| Quelle | Inhalt |
| --- | --- |
| ConfigMap | Alle nicht-sensiblen Env-Vars aus `config.*` (`APP_ENV`, `PUBLIC_BASE_URL`, `AUTO_MIGRATE`, …) |
| Secret | `JWT_SECRET`, `DATABASE_URL`; optional `REDIS_URL`, `RESEND_API_KEY`, `GEMINI_API_KEY` |

Bei Änderungen an ConfigMap oder vom Chart erzeugtem Secret starten die Pods neu (`checksum`-Annotations).

Standardmäßig ist `AUTO_MIGRATE=true`. Dabei `replicaCount` auf 1 lassen (Autoscaling aus), damit Migrationen beim Start nicht parallel laufen.

Wer ein Konto anlegen darf, steuert `config.REGISTRATION_MODE`: `open` (Default, jeder), `invite_only` (nur über einen gültigen Einladungslink für die eingeladene E-Mail) oder `closed` (keine Registrierung). Ein ungültiger Wert verhindert den Start.

Ingress und HTTPRoute (Gateway API) sind standardmäßig deaktiviert. Nach dem Aktivieren `config.PUBLIC_BASE_URL` und `config.CORS_ALLOW_ORIGINS` auf denselben öffentlichen Origin setzen. Passkeys werden an den Host aus `config.PUBLIC_BASE_URL` gebunden; ohne korrekten Wert funktionieren sie nicht (abweichend: `config.WEBAUTHN_RP_ID`, `config.WEBAUTHN_ORIGINS`).

Health-Check: `GET /healthz` auf Port 8080.

### Instanz-Admin (Betreiber-Übersicht)

Instanz-Admins sehen unter `/api/admin` eine read-only Übersicht über Benutzer, Klubs und Spieler-Verknüpfungen. Die Rolle wird nie über die Weboberfläche vergeben, sondern:

- per CLI im laufenden Pod (nutzt `DATABASE_URL` aus dem Secret):

  ```bash
  kubectl -n kegelmaster exec deploy/kegelmaster -- kegelmaster admin grant betreiber@example.com
  kubectl -n kegelmaster exec deploy/kegelmaster -- kegelmaster admin list
  kubectl -n kegelmaster exec deploy/kegelmaster -- kegelmaster admin revoke betreiber@example.com
  ```

- oder per `config.INSTANCE_ADMIN_EMAILS` (kommagetrennt): beim Start erhalten diese bestehenden Konten das Flag; es wird nie etwas entzogen. Zum Entziehen die E-Mail aus der Liste entfernen **und** `admin revoke` ausführen.

## Umgebungen (FluxCD)

Das Chart wird per FluxCD aus dem Repo `talos-fluxcd` deployt (Chart-Quelle: `helm/` auf `main`):

| Umgebung | Flux-Pfad | Image-Tags | Hostname |
| --- | --- | --- | --- |
| Produktion | `clusters/talos/apps/kegelmaster` | SemVer (`0.9.0`, nur Releases) | `kegelmaster.schnurbus.net` |
| Test | `clusters/talos/apps/kegelmaster-test` | `main-<YYYYMMDDHHmmss>-<sha>` (jeder Push auf `main`) | `kegelmaster-test.schnurbus.net` |

Die `main-*`-Tags erzeugt `.github/workflows/docker-publish.yaml`; Flux' `ImagePolicy` wählt per Zeitstempel das neueste. Die Test-Umgebung hat eine eigene Datenbank, verschickt keine E-Mails (kein `RESEND_API_KEY`; Einladungslinks stehen im Pod-Log) und läuft mit `replicaCount: 1`.

## Prüfen ohne Cluster

```bash
helm lint ./helm
helm template kegelmaster ./helm \
  --set secrets.JWT_SECRET=test \
  --set secrets.DATABASE_URL='postgres://user:pass@db:5432/kegelmaster'
```
