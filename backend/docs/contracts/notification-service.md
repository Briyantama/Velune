# notification-service — contract checklist

## API

Source: `services/notification-service/internal/delivery/http/router.go`.

| Method | Path                         |
| ------ | ---------------------------- | --- |
| GET    | `/health`                    |
| GET    | `/metrics`                   |
| GET    | `/api/v1/notifications/ping` | JWT |

OpenAPI: TBD.

## Events

| Consumed                                                                    | Notes                                                                                                   |
| --------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| `notification.overspend.requested` (and envelope types on same routing key) | Rabbit consumer `internal/infrastructure/broker/rabbitmq.go`; trace context extracted from AMQP headers |

| Produced         | Notes                                                     |
| ---------------- | --------------------------------------------------------- |
| Internal publish | `RabbitMQ.Publish` for downstream fan-out when configured |
| Outbox           | N/A as primary writer here                                |

## Change events

N/A.

## Metrics

`NotificationSentTotal`, `NotificationFailedTotal`, `NotificationRetryTotal`, etc.

## Observability

Correlation; zap; OTEL; consumer uses `otelx.ExtractAMQP`; HTTP server wrapped with `otelx.HTTPHandler`; pgx instrumented when tracing on.

## Testing

`go test ./...`; dedupe tests may use build tags (`dedupe_integration_test.go`).

## Data ownership

Migrations: `services/notification-service/migrations/` (jobs, dedupe, …).
