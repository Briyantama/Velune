# notification-service

Event-driven delivery service for overspend alerts.

## Ownership

- Consumes budget and transaction contract events through RabbitMQ.
- Applies overspend threshold policy:
  - `<100%` usage: in-app/log only
  - `>=100%` usage: in-app/log + email
- Produces `notification.dispatched` events.
- No direct coupling to other service databases.

## Endpoints

- `GET /health`
- `GET /api/v1/notifications/ping` (JWT/internal `X-User-ID` protected)

## Environment

- `HTTP_PORT` (default `8086`)
- `JWT_SECRET` (required)
- `BROKER_URL` (default `amqp://guest:guest@localhost:5672/`)
- `BROKER_EXCHANGE` (default `velune.events`)
- `BROKER_QUEUE` (default `notification.overspend`)
- `BROKER_ROUTING_KEY` (default `notification.overspend.requested`)
- `EMAIL_FROM` (default `noreply@velune.local`)

## Local run

```bash
cd services/notification-service
export HTTP_PORT=8086
export JWT_SECRET=dev-secret
export BROKER_URL=amqp://guest:guest@localhost:5672/
export BROKER_EXCHANGE=velune.events
export BROKER_QUEUE=notification.overspend
export BROKER_ROUTING_KEY=notification.overspend.requested
go run ./cmd/api
```
