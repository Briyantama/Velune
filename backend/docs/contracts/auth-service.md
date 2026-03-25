# auth-service — contract checklist

## API

Composition: outer router in `services/auth-service/cmd/api/main.go` + `internal/delivery/http/router.go`.

| Method | Path                    | Notes        |
| ------ | ----------------------- | ------------ |
| GET    | `/health`               | Outer router |
| GET    | `/metrics`              | Prometheus   |
| POST   | `/api/v1/auth/register` |              |
| POST   | `/api/v1/auth/login`    |              |
| POST   | `/api/v1/auth/refresh`  |              |
| GET    | `/api/v1/auth/me`       | JWT          |

OpenAPI: not centralized (TBD).

## Events

| Role     | Details                         |
| -------- | ------------------------------- |
| Produced | None via outbox in this service |
| Consumed | None                            |

## Change events

N/A (no `change_events` writes in auth-service).

## Metrics

Shared Prometheus handler on `/metrics`. No service-specific counters beyond defaults in `shared/metrics` (gateway/outbox metrics N/A here).

## Observability

- Correlation: `middlewares.CorrelationIDHeader` (outer stack).
- Logging: zap via `shared/logger`.
- Tracing: `shared/otelx` init in `cmd/api/main.go`; HTTP wrapped with `otelx.HTTPHandler`.

## Testing

- `go test ./...` under `services/auth-service`.
- Integration: none required for contracts doc.

## Data ownership

Postgres: `services/auth-service/migrations/` — users, refresh tokens (see `000001_init.up.sql`).
