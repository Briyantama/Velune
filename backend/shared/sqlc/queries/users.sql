-- name: CreateUser :exec
INSERT INTO users (id, email, password_hash, base_currency, status, email_verified_at, display_name, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetUserByID :one
SELECT id, email, password_hash, base_currency, status, email_verified_at, display_name, version, created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, base_currency, status, email_verified_at, display_name, version, created_at, updated_at, deleted_at
FROM users
WHERE lower(email) = lower($1) AND deleted_at IS NULL;
