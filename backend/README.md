# Velune backend

Multi-service **Expense Tracker / Money Manager** backend: independent deployable services, shared contracts, and a **strangler API gateway** for gradual migration.

## Layout

| Path | Purpose |
| ---- | ------- |
| `go.work` | Go workspace — open this folder in the IDE; build/test per module |
| `shared/` | Common libraries: `config`, `logger`, `errors` (errs), `jwt`, `pagination`, `money`, `contracts` |
| `services/auth-service` | Identity, JWT access/refresh, password hashing ([docs](services/auth-service/README.md)) |
| `services/transaction-service` | Ledger source of truth: accounts, categories, transactions, recurring, summaries ([docs](services/transaction-service/README.md)) |
| `services/category-service` | Categories (legacy scaffold; `/api/v1/categories` now routed to transaction-service) |
| `services/budget-service` | Budgets + budget usage/overspend via transaction contracts ([docs](services/budget-service/README.md)) |
| `services/report-service` | Read-only analytics service for `/api/v1/reports/*`; consumes transaction contracts ([docs](services/report-service/README.md)) |
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

4. **Services:** run from `cmd/api` for extracted services, for example:
   - `go run ./services/auth-service/cmd/api`
   - `go run ./services/transaction-service/cmd/api`
   - `go run ./services/budget-service/cmd/api`

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

Current split ownership:

- `/api/v1/auth/*` -> `auth-service`
- `/api/v1/transactions/*`, `/api/v1/accounts/*`, `/api/v1/categories/*`, `/api/v1/recurring/*` -> `transaction-service`
- `/api/v1/budgets/*` -> `budget-service`
- `/api/v1/reports/*` -> `report-service` (safe fallback to `legacy-api` on report-service 404/5xx during parity)

See `deploy/README.md` for the service matrix and ports.
