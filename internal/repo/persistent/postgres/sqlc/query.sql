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
    role_id,
    username,
    email,
    phone,
    password_hash,
    salt,
    active,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: UpdateUser :exec
UPDATE users
SET 
    role_id = COALESCE($2, role_id),
    username = COALESCE($3, username),
    email = COALESCE($4, email),
    phone = COALESCE($5, phone),
    password_hash = COALESCE($6, password_hash),
    salt = COALESCE($7, salt),
    active = COALESCE($8, active),
    updated_at = $9
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = $2, updated_at = $2
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateLastSeen :exec
UPDATE users
SET last_seen = $2
WHERE id = $1;

-- name: GetUserAttributes :many
SELECT key, value FROM user_attributes
WHERE user_id = $1;

-- name: SetUserAttribute :exec
INSERT INTO user_attributes (user_id, key, value)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, key) DO UPDATE
SET value = $3, updated_at = NOW();

-- name: DeleteUserAttribute :exec
DELETE FROM user_attributes
WHERE user_id = $1 AND key = $2;
