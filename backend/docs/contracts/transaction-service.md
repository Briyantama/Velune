# transaction-service — contract checklist

## API

Source: `services/transaction-service/internal/delivery/http/router.go`.

| Method | Path |
|--------|------|
| GET | `/health` |
| GET | `/metrics` |
| POST | `/internal/admin/reconcile/balance` | Admin API key |
| * | `/api/v1/accounts`, `/api/v1/categories`, `/api/v1/transactions`, `/api/v1/recurring` | JWT; CRUD per `router.go` |
| GET | `/api/v1/transactions/summary` |
| GET | `/api/v1/transactions/summary/categories` |

OpenAPI: TBD.

## Events

| Produced | Transport | Insert sites |
|----------|-----------|--------------|
| `transaction.created` / `transaction.updated` / `transaction.deleted` | Outbox → RabbitMQ | `internal/infrastructure/postgres/ledger.go` |
| `balance.mismatch.detected` | Outbox (reconciliation) | `internal/reconciliation/balance.go` |

| Consumed | Notes |
|----------|-------|
| None | |

Routing keys follow `EventType` (topic exchange).

## Change events

**Yes (partial):** ledger records `change_events` for transaction mutations (`ledger.go`). Account, category, recurring mutations do not emit `change_events` today — see `docs/GAP_REPORT.md`.

## Metrics

`shared/metrics`: `OutboxPending`, `OutboxRetryTotal`, `DLQMessagesTotal`, `ReconciliationMismatchTotal`, etc., as used from `cmd/api` dispatch and reconciliation.

## Observability

Correlation middleware on router; zap; OTEL (`otelx` + `otelhttp`); pgx via `otelx.InstrumentPoolConfig` when tracing enabled; Rabbit publish trace context in `shared/events/rabbitmq_publisher.go`.

## Testing

- `go test ./...` including `tests/usecase`.

## Data ownership

Migrations: `services/transaction-service/migrations/` (`accounts`, `categories`, `transactions`, `recurring`, `change_events`, `event_outbox`, …).
