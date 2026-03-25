# admin-service — contract checklist

## API

Source: `services/admin-service/internal/delivery/http/handlers.go` (`Routes()`).

| Method | Path | Auth |
|--------|------|------|
| GET | `/health` | Public |
| GET | `/metrics` | Public |
| GET | `/internal/admin/health` | `ADMIN_API_KEY` |
| GET | `/internal/admin/dlq` | |
| POST | `/internal/admin/dlq/replay` | |
| GET | `/internal/admin/outbox` | |
| POST | `/internal/admin/outbox/retry` | |
| GET | `/internal/admin/notifications/jobs` | |
| POST | `/internal/admin/notifications/jobs/retry` | |
| POST | `/internal/admin/reconcile/balance` | |
| POST | `/internal/admin/reconcile/budget` | |
| GET | `/internal/admin/reconcile/logs` | |
| POST | `/internal/admin/events/replay` | |

OpenAPI: TBD.

## Events

| Produced | Notes |
|----------|-------|
| Optional Rabbit republish | DLQ replay / event replay paths via `shared/events.RabbitPublisher` |

| Consumed | Notes |
|----------|-------|
| None | |

## Change events

Read-only inspection across DB pools; does not write `change_events`.

## Metrics

`/metrics` via shared Prometheus handler.

## Observability

Correlation; zap; OTEL HTTP; HTTP client to upstreams uses `otelx.TracedHTTPClient`; pgx pools use `otelx.InstrumentPoolConfig` when tracing enabled.

## Testing

`go test ./...` under `services/admin-service`.

## Data ownership

No standalone migrations; connects to `TRANSACTION_DATABASE_URL`, `BUDGET_DATABASE_URL`, `NOTIFICATION_DATABASE_URL`.
