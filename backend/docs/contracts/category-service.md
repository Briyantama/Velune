# category-service — contract checklist

## API

Source: `services/category-service/cmd/server/main.go` (inline chi).

| Method | Path |
|--------|------|
| GET | `/health` |
| GET | `/metrics` |
| GET | `/api/v1/meta` | Stub metadata |

OpenAPI: TBD.

## Events

None.

## Change events

N/A.

## Metrics

Shared `/metrics` handler.

## Observability

Correlation + OTEL HTTP wrapper.

## Testing

Limited `go test` surface; smoke via `/health`.

## Data ownership

None in-repo (no migrations); categories live in transaction-service DB for production data.
