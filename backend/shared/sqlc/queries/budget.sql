-- name: BudgetCreate :exec
INSERT INTO budgets (
  id, user_id, name, period_type, category_id, start_date, end_date,
  limit_amount_minor, currency, version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6::date, $7::date, $8, $9, $10, $11, $12);

-- name: BudgetGetByID :one
SELECT id, user_id, name, period_type, category_id, start_date, end_date, limit_amount_minor, currency, version, created_at, updated_at, deleted_at
FROM budgets
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: BudgetCountList :one
SELECT COUNT(*)
FROM budgets
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (NOT sqlc.arg(has_active_on)::bool OR (start_date <= sqlc.arg(active_on)::date AND end_date >= sqlc.arg(active_on)::date));

-- name: BudgetList :many
SELECT id, user_id, name, period_type, category_id, start_date, end_date, limit_amount_minor, currency, version, created_at, updated_at, deleted_at
FROM budgets
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (NOT sqlc.arg(has_active_on)::bool OR (start_date <= sqlc.arg(active_on)::date AND end_date >= sqlc.arg(active_on)::date))
ORDER BY start_date DESC
LIMIT $2 OFFSET $3;

-- name: BudgetUpdate :execrows
UPDATE budgets
SET name = $1,
    period_type = $2,
    category_id = $3,
    start_date = $4::date,
    end_date = $5::date,
    limit_amount_minor = $6,
    currency = $7,
    version = $8,
    updated_at = $9
WHERE id = $10
  AND user_id = $11
  AND version = $12
  AND deleted_at IS NULL;

-- name: BudgetSoftDelete :execrows
UPDATE budgets
SET deleted_at = now(),
    version = version + 1,
    updated_at = now()
WHERE id = $1
  AND user_id = $2
  AND version = $3
  AND deleted_at IS NULL;
