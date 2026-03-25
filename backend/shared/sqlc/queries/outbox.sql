-- name: OutboxInsert :exec
INSERT INTO event_outbox (id, event_type, payload, status, retry_count, next_retry_at, created_at, updated_at)
VALUES ($1, $2, $3, 'pending', 0, now(), now(), now());

-- name: OutboxMarkSent :exec
UPDATE event_outbox
SET status = 'sent', updated_at = now()
WHERE id = $1;

-- name: OutboxListDueForDispatch :many
SELECT id, payload, retry_count
FROM event_outbox
WHERE status IN ('pending', 'failed')
  AND retry_count < $1
  AND next_retry_at <= now()
ORDER BY created_at ASC
LIMIT $2;

-- name: OutboxMarkFailedBumpRetry :exec
UPDATE event_outbox
SET status = 'failed', retry_count = retry_count + 1, updated_at = now()
WHERE id = $1;

-- name: OutboxMarkFailedScheduleRetry :exec
UPDATE event_outbox
SET status = 'failed', retry_count = retry_count + 1, next_retry_at = $2, updated_at = now()
WHERE id = $1;

-- name: OutboxCountRetryEligible :one
SELECT COUNT(*)::bigint
FROM event_outbox
WHERE status IN ('pending', 'failed')
  AND retry_count < $1;

-- name: OutboxCountPendingOrFailed :one
SELECT COUNT(*)::int
FROM event_outbox
WHERE status IN ('pending', 'failed');

-- name: OutboxListAdmin :many
SELECT id::text,
  event_type,
  status,
  retry_count,
  next_retry_at,
  created_at,
  CASE
    WHEN length(payload::text) > 512 THEN left(payload::text, 512) || '...'
    ELSE payload::text
  END AS payload_preview
FROM event_outbox
WHERE ($1::text = '' OR status = $1)
ORDER BY created_at DESC
LIMIT $2;

-- name: OutboxRetryReset :execrows
UPDATE event_outbox
SET status = 'pending', next_retry_at = now(), updated_at = now()
WHERE id = $1;

-- name: OutboxPayloadsForReplay :many
SELECT payload::text
FROM event_outbox
WHERE created_at >= $1
  AND created_at < $2
  AND status IN ('pending', 'failed')
  AND ($3::text = '' OR event_type = $3)
ORDER BY created_at ASC
LIMIT 100;
