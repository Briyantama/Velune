# Legacy API route inventory

Strangler entrypoint for clients is **api-gateway**. The `legacy-api` service exists for transitional catch-all routing and compose defaults.

## Previously exposed (removed in shrink)

| Route pattern | Domain | Status | Replacement |
| ------------- | ------ | ------ | ----------- |
| `GET /health` | ops | **active** (shell) | Same on `legacy-api` shell |
| `/api/v1/budgets/*` | budgets | **removed** | Gateway → `budget-service` (`BUDGET_SERVICE_URL`) |
| `/api/v1/reports/monthly` (never wired in legacy router) | reports | **dead code** (handlers existed, not mounted) | `report-service` (`GET /api/v1/reports/monthly`) |
| Auth, accounts, transactions, categories, recurring | various | **dead code** (unwired) | Respective microservices via gateway |

## Gateway routing (source of truth)

| Prefix | Upstream |
| ------ | -------- |
| `/api/v1/auth` | auth-service |
| `/api/v1/transactions`, `/accounts`, `/recurring`, `/categories` | transaction-service |
| `/api/v1/budgets` | budget-service |
| `/api/v1/notifications` | notification-service |
| `/api/v1/reports` | report-service (+ optional fallback; legacy reports retired) |
| other `/api/v1/*` | `LEGACY_API_URL` if set |

## Legacy shell goal state

- **Health** and observability endpoints only (`/health`, `/metrics`).
- **No** business API on legacy host; clients must use the gateway.
- **Report fallback**: gateway does not rely on legacy for `/api/v1/reports/*`; if report-service fails, response is an error JSON (not a silent legacy 404).

## Deprecated direct access

Calling `legacy-api:8090` directly for budgets or reports is **deprecated** and unsupported after this shrink.
