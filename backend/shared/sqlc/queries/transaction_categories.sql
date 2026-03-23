-- name: CategoryCreate :exec
INSERT INTO categories (id, user_id, name, parent_id, version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: CategoryGetByID :one
SELECT id, user_id, name, parent_id, version, created_at, updated_at, deleted_at
FROM categories
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL;

-- name: CategoryCountList :one
SELECT COUNT(*) FROM categories WHERE user_id = $1 AND deleted_at IS NULL;

-- name: CategoryList :many
SELECT id, user_id, name, parent_id, version, created_at, updated_at, deleted_at
FROM categories
WHERE user_id = $1 AND deleted_at IS NULL
ORDER BY name ASC
LIMIT $2 OFFSET $3;

-- name: CategoryUpdate :execrows
UPDATE categories
SET name = $1, parent_id = $2, version = $3, updated_at = $4
WHERE id = $5 AND user_id = $6 AND version = $7 AND deleted_at IS NULL;

-- name: CategorySoftDelete :execrows
UPDATE categories
SET deleted_at = now(), version = version + 1, updated_at = now()
WHERE id = $1 AND user_id = $2 AND version = $3 AND deleted_at IS NULL;
