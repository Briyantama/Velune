# legacy-api — contract checklist

## API

Source: `services/legacy-api/internal/delivery/http/router.go`.

| Method | Path | Notes |
|--------|------|------|
| GET | `/health` | |
| GET | `/metrics` | |
| * | `/api/v1/*` | `410 Gone` strangler shell |

OpenAPI: TBD.

## Events

None.

## Change events

N/A.

## Metrics

Shared `/metrics`.

## Observability

Correlation; OTEL HTTP wrapper in `cmd/api/main.go`.

## Testing

`go test ./...` under `services/legacy-api` where present.

## Data ownership

Optional DB via `DATABASE_URL` in env example for legacy shell; strangler routes do not serve full business API.
