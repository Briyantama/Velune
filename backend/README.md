# Velune backend

Multi-service **Expense Tracker / Money Manager** backend: independent deployable services, shared contracts, and a **strangler API gateway** for gradual migration.

## Layout

| Path | Purpose |
|------|---------|
| `go.work` | Go workspace — open this folder in the IDE; build/test per module |
| `shared/` | Common libraries: `config`, `logger`, `errors` (errs), `jwt`, `pagination`, `money`, `contracts` |
| `services/auth-service` | Identity, JWT access/refresh, password hashing ([docs](services/auth-service/README.md)) |
| `services/transaction-service` | Ledger, accounts, transactions, recurring (scaffold) |
| `services/category-service` | Categories (scaffold) |
| `services/budget-service` | Budgets & limits (scaffold) |
| `services/report-service` | Aggregations; calls peers via HTTP only (scaffold) |
| `services/notification-service` | Alerts / reminders / future channels (scaffold) |
| `services/api-gateway` | HTTP entry: routes `/api/v1/*` to upstreams or `LEGACY_API_URL` |
| `services/legacy-api` | **Current full API** (clean architecture) until domains are extracted |
| `infra/` | Docker Compose, Postgres init scripts |
| `deploy/` | Deployment notes |

## Principles

- **No business logic in HTTP handlers** — use `domain` → `usecase` → `repository` → `infrastructure`.
- **No cross-service database access** — integrate with IDs and HTTP contracts.
- **Configuration only from environment variables** — see `env.example`.

## Local development

1. Install Go **1.22+**.
2. From **`backend/`** (workspace root): `go work sync` (optional).
3. **Legacy API (full feature set today):**

   ```bash
   cd services/legacy-api
   export DATABASE_URL=postgres://postgres:postgres@localhost:5432/velune_legacy?sslmode=disable
   export JWT_SECRET=dev-secret
   go run ./cmd/api
   ```

4. **Scaffold/services:** auth-service uses `cmd/api` — e.g. `go run ./services/auth-service/cmd/api` from `backend/` (set `HTTP_PORT`).

5. **Docker:** `docker compose -f infra/docker-compose.yml up legacy-api postgres redis` — Postgres + legacy API on port **8090** (see compose file).

6. **Split stack + gateway:** `docker compose -f infra/docker-compose.yml --profile split up` — builds all images; gateway on **8080** (see `deploy/README.md`).

## Testing

```bash
cd services/legacy-api && go test ./...
```

Add unit tests per service under `internal/usecase` / `domain` as features are extracted.

## Migration (strangler)

1. Keep **`legacy-api`** stable for clients.
2. Implement a bounded context in its dedicated service + DB schema.
3. Point **`api-gateway`** env vars for that path prefix to the new service.
4. When all paths are migrated, remove `LEGACY_API_URL` and retire `legacy-api`.

See `deploy/README.md` for the service matrix and ports.
