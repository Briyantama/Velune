-- name: LedgerUpdateTransaction :execrows
UPDATE transactions
SET account_id = $1,
    category_id = $2,
    counterparty_account_id = $3,
    amount_minor = $4,
    currency = $5,
    type = $6,
    description = $7,
    occurred_at = $8,
    version = $9,
    updated_at = $10
WHERE id = $11 AND user_id = $12 AND version = $13 AND deleted_at IS NULL;

-- name: LedgerSoftDeleteTransaction :execrows
UPDATE transactions
SET deleted_at = now(), version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $3 AND deleted_at IS NULL;

-- name: LedgerLoadTransactionForUpdate :one
SELECT id, user_id, account_id, category_id, counterparty_account_id, amount_minor, currency, type,
       description, occurred_at, version, created_at, updated_at, deleted_at
FROM transactions
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
FOR UPDATE;

-- name: LedgerSelectAccountForUpdate :one
SELECT id, balance_minor, currency, version
FROM accounts
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
FOR UPDATE;

-- name: LedgerUpdateAccountBalanceVersion :execrows
UPDATE accounts
SET balance_minor = $1, version = version + 1, updated_at = now()
WHERE id = $2 AND user_id = $3 AND version = $4 AND deleted_at IS NULL;

-- name: LedgerInsertTransaction :exec
INSERT INTO transactions (
  id, user_id, account_id, category_id, counterparty_account_id, amount_minor, currency, type, description, occurred_at, version, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);

-- name: LedgerInsertChangeEvent :exec
INSERT INTO change_events (user_id, entity_type, entity_id, operation, version, payload)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: LedgerInsertLedgerEntry :exec
INSERT INTO ledger_entries (transaction_id, user_id, account_id, direction, amount_minor, currency, reason)
VALUES ($1, $2, $3, $4, $5, $6, $7);
