-- name: AccountCreate :exec
INSERT INTO accounts (id, user_id, name, type, currency, balance_minor, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: AccountGetByID :one
SELECT id, user_id, name, type, currency, balance_minor, version, created_at, updated_at, deleted_at
FROM accounts
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: AccountCountList :one
SELECT COUNT(*) FROM accounts WHERE user_id = $1 AND deleted_at IS NULL;

-- name: AccountList :many
SELECT id, user_id, name, type, currency, balance_minor, version, created_at, updated_at, deleted_at
FROM accounts
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: AccountUpdate :execrows
UPDATE accounts
SET name = $1, type = $2, version = $3, updated_at = $4
WHERE id = $5 AND user_id = $6 AND version = $7 AND deleted_at IS NULL;

-- name: AccountSoftDelete :execrows
UPDATE accounts
SET deleted_at = now(), version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $3 AND deleted_at IS NULL;
