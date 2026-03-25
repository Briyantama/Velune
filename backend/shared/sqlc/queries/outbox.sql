-- name: OutboxInsert :exec
INSERT INTO event_outbox (id, event_type, payload, status, retry_count, next_retry_at, created_at, updated_at)
VALUES ($1, $2, $3, 'pending', 0, now(), now(), now());

-- name: OutboxMarkSent :exec
UPDATE event_outbox
SET status = 'sent', updated_at = now()
WHERE id = $1;
