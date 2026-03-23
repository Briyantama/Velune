# Deployment notes

## Services

| Service | Responsibility | Default port (local) |
|---------|------------------|------------------------|
| `api-gateway` | Strangler HTTP entry, routes to microservices or legacy | 8080 |
| `legacy-api` | Monolithic API during migration (Postgres `velune_legacy`) | 8090 |
| `auth-service` | Identity, JWT access/refresh, password hashing | 8081 |
| `transaction-service` | Ledger, accounts, transactions, recurring postings | 8082 |
| `category-service` | Categories (default + user-defined) | 8083 |
| `budget-service` | Budget limits, usage, exceed detection | 8084 |
| `report-service` | Aggregations; calls other services via HTTP only | 8085 |
| `notification-service` | Alerts, reminders, future email/push/webhooks | 8086 |

## Local compose

- **Legacy only (simplest):** `docker compose -f infra/docker-compose.yml up legacy-api postgres redis`
- **Full split + gateway:** `docker compose -f infra/docker-compose.yml --profile split up`

The gateway forwards by path prefix; unset upstream env vars fall through to `LEGACY_API_URL` when set.

## Configuration

All configuration is environment-driven. See root `env.example`. Never commit secrets.
