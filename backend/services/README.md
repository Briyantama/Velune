# Services

Each directory is a **separate Go module** (`go.mod`) with a single responsibility. Use `github.com/moon-eye/velune/shared` for common types and utilities (replace path in `go.mod` for local dev).

| Service | Module | Port (example) |
|---------|--------|------------------|
| `auth-service` | `.../services/auth-service` | 8081 |
| `transaction-service` | `.../services/transaction-service` | 8082 |
| `category-service` | `.../services/category-service` | 8083 |
| `budget-service` | `.../services/budget-service` | 8084 |
| `report-service` | `.../services/report-service` | 8085 |
| `notification-service` | `.../services/notification-service` | 8086 |
| `api-gateway` | `.../services/api-gateway` | 8080 |
| `legacy-api` | `.../services/legacy-api` | 8090 |

Boilerplate services expose `GET /health` and `GET /api/v1/meta`. Extract handlers and persistence from `legacy-api` incrementally.
