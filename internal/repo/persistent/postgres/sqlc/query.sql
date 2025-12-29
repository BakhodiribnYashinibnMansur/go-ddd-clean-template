-- PostgreSQL schema for user management with sqlc
-- This file will be used by sqlc to generate type-safe Go code

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByPhone :one
SELECT * FROM users
WHERE phone = $1 AND deleted_at IS NULL
LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (
    username,
    phone,
    password_hash,
    salt,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users
SET 
    username = COALESCE($2, username),
    phone = COALESCE($3, phone),
    password_hash = COALESCE($4, password_hash),
    salt = COALESCE($5, salt),
    updated_at = $6
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = $2, updated_at = $2
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateLastSeen :exec
UPDATE users
SET last_seen = $2
WHERE id = $1;
