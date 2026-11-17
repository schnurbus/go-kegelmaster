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

Ingress und HTTPRoute (Gateway API) sind standardmäßig deaktiviert. Nach dem Aktivieren `config.PUBLIC_BASE_URL` und `config.CORS_ALLOW_ORIGINS` auf denselben öffentlichen Origin setzen.

Health-Check: `GET /healthz` auf Port 8080.

## Prüfen ohne Cluster

```bash
helm lint ./helm
helm template kegelmaster ./helm \
  --set secrets.JWT_SECRET=test \
  --set secrets.DATABASE_URL='postgres://user:pass@db:5432/kegelmaster'
```
