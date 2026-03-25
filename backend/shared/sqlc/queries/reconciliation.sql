-- name: ReconcileAccountStoredVsLedger :many
SELECT a.id,
  a.user_id,
  a.balance_minor,
  a.currency,
  COALESCE(SUM(
    CASE le.direction
      WHEN 'credit' THEN le.amount_minor
      WHEN 'debit' THEN -le.amount_minor
      ELSE 0
    END
  ), 0)::bigint AS ledger_sum
FROM accounts a
LEFT JOIN ledger_entries le ON le.account_id = a.id
WHERE a.deleted_at IS NULL
GROUP BY a.id, a.user_id, a.balance_minor, a.currency;
