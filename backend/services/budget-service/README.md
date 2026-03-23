# budget-service

Owns budget definitions and budget progress computation.

It stores budgets in `velune_budget` and consumes transaction summary data from `transaction-service` via HTTP contracts.

## Endpoints

Base path: `/api/v1` (JWT required, except `/health`).

- `POST /budgets`
- `GET /budgets`
- `PUT /budgets/{id}?version={n}`
- `DELETE /budgets/{id}?version={n}`
- `GET /budgets/{id}/usage`

`/budgets/{id}/usage` calculates:

- `spentMinor`
- `remainingMinor`
- `overspentMinor`
- `isOverspent`

using `transaction-service` summary endpoints.

## Environment variables

- `DATABASE_URL` (required): `velune_budget` PostgreSQL DSN.
- `JWT_SECRET` (required): JWT verification secret.
- `TRANSACTION_SERVICE_URL` (required): upstream transaction-service base URL.
- `HTTP_PORT` (default `8084` in split profile).
- `MIGRATIONS_PATH` (e.g. `file:///app/migrations`).

## Contract usage

Budget-service consumes shared transaction contracts from `shared/contracts`, including summary payload shapes used for utilization/overspend calculations.

