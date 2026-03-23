-- name: TransactionGetByID :one
SELECT id, user_id, account_id, category_id, counterparty_account_id, amount_minor, currency, type,
       description, occurred_at, version, created_at, updated_at, deleted_at
FROM transactions
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: TransactionCountList :one
SELECT COUNT(*)
FROM transactions
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (NOT sqlc.arg(has_account_id)::bool OR account_id = sqlc.arg(account_id))
  AND (NOT sqlc.arg(has_category_id)::bool OR category_id = sqlc.arg(category_id))
  AND (NOT sqlc.arg(has_type)::bool OR type = sqlc.arg(type))
  AND (NOT sqlc.arg(has_from)::bool OR occurred_at >= sqlc.arg(from_at))
  AND (NOT sqlc.arg(has_to)::bool OR occurred_at < sqlc.arg(to_at))
  AND (sqlc.arg(currency) = '' OR currency = sqlc.arg(currency));

-- name: TransactionList :many
SELECT id, user_id, account_id, category_id, counterparty_account_id, amount_minor, currency, type,
       description, occurred_at, version, created_at, updated_at, deleted_at
FROM transactions
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (NOT sqlc.arg(has_account_id)::bool OR account_id = sqlc.arg(account_id))
  AND (NOT sqlc.arg(has_category_id)::bool OR category_id = sqlc.arg(category_id))
  AND (NOT sqlc.arg(has_type)::bool OR type = sqlc.arg(type))
  AND (NOT sqlc.arg(has_from)::bool OR occurred_at >= sqlc.arg(from_at))
  AND (NOT sqlc.arg(has_to)::bool OR occurred_at < sqlc.arg(to_at))
  AND (sqlc.arg(currency) = '' OR currency = sqlc.arg(currency))
ORDER BY occurred_at DESC, id DESC
LIMIT $2 OFFSET $3;

-- name: TransactionSumIncomeExpenseInRange :one
SELECT
  COALESCE(SUM(CASE WHEN type = 'income' THEN amount_minor ELSE 0 END), 0) AS income_minor,
  COALESCE(SUM(CASE WHEN type = 'expense' THEN amount_minor ELSE 0 END), 0) AS expense_minor
FROM transactions
WHERE user_id = $1
  AND deleted_at IS NULL
  AND currency = $2
  AND occurred_at >= $3
  AND occurred_at < $4;

-- name: TransactionSumExpensesByCategoryInRange :many
SELECT category_id, COALESCE(SUM(amount_minor), 0) AS total_minor
FROM transactions
WHERE user_id = $1
  AND deleted_at IS NULL
  AND type = 'expense'
  AND occurred_at >= $2
  AND occurred_at < $3
  AND currency = $4
  AND category_id IS NOT NULL
GROUP BY category_id;
