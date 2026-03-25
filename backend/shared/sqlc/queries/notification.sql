-- name: NotificationJobEnqueue :exec
INSERT INTO notification_jobs (id, user_id, channel, payload, status, retry_count, next_retry_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, 'pending', 0, now(), now(), now());

-- name: NotificationJobsFetchDue :many
SELECT id, user_id, channel, payload, status, retry_count, next_retry_at
FROM notification_jobs
WHERE status IN ('pending', 'processing')
  AND next_retry_at <= now()
ORDER BY created_at ASC
LIMIT $1;

-- name: NotificationJobMarkSent :exec
UPDATE notification_jobs
SET status = 'sent', updated_at = now()
WHERE id = $1;

-- name: NotificationJobMarkRetry :exec
UPDATE notification_jobs
SET status = 'processing', retry_count = $2, next_retry_at = $3, updated_at = now()
WHERE id = $1;

-- name: NotificationJobMarkFailed :exec
UPDATE notification_jobs
SET status = 'failed', updated_at = now()
WHERE id = $1;

-- name: NotificationJobsCountFailedOrPending :one
SELECT COUNT(*)::int
FROM notification_jobs
WHERE status IN ('failed', 'pending');

-- name: NotificationJobsListAdmin :many
SELECT id::text,
  user_id::text,
  channel,
  status,
  retry_count,
  next_retry_at,
  created_at,
  CASE
    WHEN length(payload::text) > 512 THEN left(payload::text, 512) || '...'
    ELSE payload::text
  END AS payload_preview
FROM notification_jobs
WHERE ($1::text = '' OR status = $1)
ORDER BY created_at DESC
LIMIT $2;

-- name: NotificationJobRetryReset :execrows
UPDATE notification_jobs
SET status = 'pending', next_retry_at = now(), updated_at = now()
WHERE id = $1
  AND status IN ('failed', 'pending');

-- name: EventDedupeInsert :exec
INSERT INTO event_dedupe (idempotency_key, event_id, processed_at)
VALUES ($1, $2, now());

-- name: EventDedupeTruncate :exec
TRUNCATE event_dedupe;
