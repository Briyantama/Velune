# budget-service — contract checklist

## API

Source: `services/budget-service/internal/delivery/http/router.go`.

| Method | Path                                            |
| ------ | ----------------------------------------------- | ------------------------- |
| GET    | `/health`                                       |
| GET    | `/metrics`                                      |
| POST   | `/internal/admin/reconcile/budget`              |
| \*     | `/api/v1/budgets`, `/api/v1/budgets/{id}/usage` | JWT; verbs in `router.go` |

OpenAPI: TBD.

## Events

| Produced                           | Notes                                                          |
| ---------------------------------- | -------------------------------------------------------------- |
| `notification.overspend.requested` | Emitted when overspend evaluated (`budget_repo.go` / use case) |
| `budget.usage.evaluated`           | Where implemented                                              |
| `budget.mismatch.detected`         | Reconciliation (`internal/reconciliation/summary.go`)          |

Published via transactional outbox + dispatcher in `cmd/api`; Rabbit envelope in `shared/contracts`.

| Consumed | Notes |
| -------- | ----- |
| None     |       |

## Change events

N/A in budget DB (no `change_events` table in budget migrations).

## Metrics

Outbox / DLQ / overspend-related metrics from `shared/metrics` as referenced in service.

## Observability

Correlation; zap; OTEL HTTP + pgx (`otelx`); outbound HTTP to transaction-service uses `otelx.TracedHTTPClient` in `internal/infrastructure/transactions/client.go`.

## Testing

`go test ./...` including `tests/usecase`.

## Data ownership

Migrations: `services/budget-service/migrations/` (budgets, outbox, alert state, audit logs).
