-- name: AuditLogInsert :exec
INSERT INTO audit_logs (type, status, details, created_at)
VALUES ($1, $2, $3, now());

-- name: AuditLogsList :many
SELECT id::text, type, status, details, created_at
FROM audit_logs
WHERE ($1::text = '' OR type = $1)
ORDER BY created_at DESC
LIMIT $2;
