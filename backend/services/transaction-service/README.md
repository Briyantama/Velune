# transaction-service

Ledger source of truth for transaction flows:

- transactions (create/list/get/update/delete)
- account balance mutations
- ledger entries
- category linkage
- recurring rules
- transaction summaries for budget consumption

## Endpoints

All endpoints are under `/api/v1` and require JWT (`Authorization: Bearer <access_token>`), except `/health`.

- `POST /transactions`
- `GET /transactions`
- `GET /transactions/{id}`
- `PATCH /transactions/{id}?version={n}`
- `DELETE /transactions/{id}?version={n}`
- `GET /transactions/summary?from=&to=&currency=`
- `GET /transactions/summary/categories?from=&to=&currency=`

Additional migrated routes:

- `/accounts` CRUD
- `/categories` CRUD
- `/recurring` CRUD

## Environment variables

- `DATABASE_URL` (required): `velune_transaction` PostgreSQL DSN.
- `JWT_SECRET` (required): JWT verification secret.
- `HTTP_PORT` (default `8082` in compose split profile).
- `MIGRATIONS_PATH` (for startup migrations, e.g. `file:///app/migrations`).
- `JWT_EXPIRY` (shared setting, optional).

## Notes

- `user_id` is an opaque cross-service identifier (no FK to auth DB tables).
- money values use `amount_minor` integer style, never floating-point.
- ledger writes are transactional and update account balances atomically.

