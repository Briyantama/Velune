# Gap report — backend vs architecture rules

This document records alignment and intentional gaps after aligning docs, OpenTelemetry, and per-service contract notes.

## Aligned

- **Microservices** with Postgres per service and **no cross-service DB access** in application code.
- **sqlc + pgx** for typed SQL; migrations per service.
- **RabbitMQ** integration with **transactional outbox** in transaction-service and budget-service; dispatch loops in `cmd/api`.
- **Prometheus** `/metrics` on deployable services using `shared/metrics`.
- **Correlation ID** middleware (`shared/middlewares`) on HTTP entrypoints.
- **Clean Architecture-style layout**: `domain`, `usecase`, `repository`, `delivery/http`, `infrastructure/*`.
- **OpenTelemetry (minimal)**:
  - `shared/otelx`: TracerProvider (stdout or OTLP/HTTP), W3C propagation, HTTP server (`otelhttp`), HTTP client transport, pgx (`otelpgx`) when tracing enabled, AMQP header inject/extract for publisher and notification consumer.
  - Env gates documented in `env.example` (`OTEL_TRACES_ENABLED`, OTLP endpoints, `OTEL_SERVICE_NAME`).

## Partial

- **`change_events`**: implemented for **transaction** mutations via ledger only. Accounts, categories, recurring, budgets, etc. do **not** emit `change_events` today — narrower than the historical “all financial mutations” rule. Extend if audit/product requires.
- **OpenAPI**: not generated or centralized per service; contract truth is `router.go` + `docs/contracts/*.md`.
- **Redis Streams**: documented as **not used**; RabbitMQ is the broker in this repo.
- **OTEL**: metrics bridge and optional log-trace correlation are **not** implemented; focus is traces + propagation.

## Missing / follow-ups

- **Broader change_events** coverage for non-transaction aggregates (if desired).
- **testcontainers-go**-based integration tests across services (not mandatory for current CI).
- **Per-service OpenAPI** (or aggregated spec) kept in sync with routers.
- **End-to-end trace verification** in CI (optional smoke with OTLP collector).
- **Full worker** boundaries remain distributed (outbox + notification consumer) vs. a single `worker-service` module — by design per current architecture.
