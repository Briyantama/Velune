# report-service — contract checklist

## API

Source: `services/report-service/internal/delivery/http/router.go`.

| Method | Path                      |
| ------ | ------------------------- | ---------------------------- |
| GET    | `/health`                 |
| GET    | `/metrics`                |
| GET    | `/api/v1/reports/monthly` | Query: year, month, currency |

OpenAPI: TBD.

## Events

| Produced | Consumed |
| -------- | -------- |
| None     | None     |

## Change events

N/A.

## Metrics

Default `/metrics` via shared handler.

## Observability

Correlation; zap; OTEL; transaction client HTTP traced (`internal/infrastructure/transactions/client.go`).

## Testing

`go test ./...` (includes `router_test`).

## Data ownership

No dedicated Postgres migrations in-repo for this service; read-only aggregation via transaction-service HTTP API.
