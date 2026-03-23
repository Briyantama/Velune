-- name: StoreRefreshToken :exec
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, 1, now(), now());

-- name: GetRefreshTokenByTokenHash :one
SELECT id, user_id, token_hash, expires_at, version, created_at, updated_at, deleted_at
FROM refresh_tokens
WHERE token_hash = $1 AND deleted_at IS NULL AND expires_at > now();

-- name: RotateRefreshToken :execrows
UPDATE refresh_tokens
SET token_hash = $1,
    expires_at = $2,
    version = version + 1,
    updated_at = now()
WHERE id = $3 AND deleted_at IS NULL;

-- name: SoftDeleteRefreshToken :execrows
UPDATE refresh_tokens
SET deleted_at = now(),
    version = version + 1,
    updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;
