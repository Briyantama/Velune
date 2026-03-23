# report-service

Read-only analytics service for `/api/v1/reports/*`.

## Ownership

- Serves report endpoints (monthly report parity first).
- Consumes transaction analytics from `transaction-service` via HTTP contracts.
- Does not own or write to transaction tables.

## Configuration

- `HTTP_PORT` (default `8085`)
- `HTTP_HOST` (default `0.0.0.0`)
- `JWT_SECRET` (required)
- `TRANSACTION_SERVICE_URL` (required, e.g. `http://transaction-service:8082`)

## Run

```bash
cd services/report-service
export HTTP_PORT=8085
export JWT_SECRET=dev-secret
export TRANSACTION_SERVICE_URL=http://localhost:8082
go run ./cmd/api
```

## Endpoints

- `GET /health`
- `GET /api/v1/reports/monthly?year=2026&month=3&currency=USD`

## Gateway cutover mode

`api-gateway` forwards `/api/v1/reports/*` to `report-service` first and falls back to `LEGACY_API_URL` when `report-service` returns 404 or 5xx during parity rollout.
