# api-gateway — contract checklist

## API

Source: `services/api-gateway/cmd/server/main.go` (`net/http.ServeMux`).

| Method | Path                                   | Notes                                                        |
| ------ | -------------------------------------- | ------------------------------------------------------------ |
| GET    | `/health`                              |                                                              |
| GET    | `/metrics`                             |                                                              |
| GET    | `/api/v1/gateway/routes`               | Route listing                                                |
| GET    | `/api/v1/reports`, `/api/v1/reports/*` | Proxied to report-service                                    |
| \*     | `/api/v1/*`                            | Prefix routing to microservices or `LEGACY_API_URL` fallback |

OpenAPI: TBD.

## Events

None (stateless forwarder).

## Change events

N/A.

## Metrics

`GatewayRequestsTotal`, `GatewayFallbackHitsTotal` (labels per route group).

## Observability

`middlewares.CorrelationIDHeader`; gateway request logging; `otelx.HTTPHandler` on outer stack; `shared/httpx` reverse proxies inject W3C headers on upstream requests via `otelx.InjectHTTP`.

## Testing

`go test` in `services/api-gateway` (e.g. `main_test.go` if present).

## Data ownership

None (no database).
