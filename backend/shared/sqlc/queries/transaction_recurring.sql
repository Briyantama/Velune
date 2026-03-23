-- name: RecurringCreate :exec
INSERT INTO recurring_rules (
  id, user_id, account_id, category_id, amount_minor, currency, type, frequency,
  next_run_at, last_run_at, is_active, description, version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);

-- name: RecurringGetByID :one
SELECT id, user_id, account_id, category_id, amount_minor, currency, type, frequency, next_run_at, last_run_at, is_active, description, version, created_at, updated_at, deleted_at
FROM recurring_rules
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: RecurringCountList :one
SELECT COUNT(*)
FROM recurring_rules
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (NOT sqlc.arg(active_only)::bool OR is_active = true);

-- name: RecurringList :many
SELECT id, user_id, account_id, category_id, amount_minor, currency, type, frequency, next_run_at, last_run_at, is_active, description, version, created_at, updated_at, deleted_at
FROM recurring_rules
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (NOT sqlc.arg(active_only)::bool OR is_active = true)
ORDER BY next_run_at ASC
LIMIT $2 OFFSET $3;

-- name: RecurringListDue :many
SELECT id, user_id, account_id, category_id, amount_minor, currency, type, frequency, next_run_at, last_run_at, is_active, description, version, created_at, updated_at, deleted_at
FROM recurring_rules
WHERE deleted_at IS NULL AND is_active = true AND next_run_at <= $1
ORDER BY next_run_at ASC
LIMIT $2;

-- name: RecurringUpdate :execrows
UPDATE recurring_rules
SET account_id = $1,
    category_id = $2,
    amount_minor = $3,
    currency = $4,
    type = $5,
    frequency = $6,
    next_run_at = $7,
    last_run_at = $8,
    is_active = $9,
    description = $10,
    version = $11,
    updated_at = $12
WHERE id = $13 AND user_id = $14 AND version = $15 AND deleted_at IS NULL;

-- name: RecurringSoftDelete :execrows
UPDATE recurring_rules
SET deleted_at = now(), version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $3 AND deleted_at IS NULL;
