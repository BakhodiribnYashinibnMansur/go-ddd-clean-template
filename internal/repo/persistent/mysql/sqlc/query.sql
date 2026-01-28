-- MySQL schema for user management with sqlc

-- name: GetUser :one
SELECT * FROM users
WHERE id = ? AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByPhone :one
SELECT * FROM users
WHERE phone = ? AND deleted_at IS NULL
LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE deleted_at IS NULL;

-- name: CreateUser :execresult
INSERT INTO users (
    username,
    phone,
    password_hash,
    salt,
    created_at,
    updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?
);

-- name: UpdateUser :exec
UPDATE users
SET 
    username = COALESCE(?, username),
    phone = COALESCE(?, phone),
    password_hash = COALESCE(?, password_hash),
    salt = COALESCE(?, salt),
    updated_at = ?
WHERE id = ? AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = ?, updated_at = ?
WHERE id = ? AND deleted_at IS NULL;

-- name: UpdateLastSeen :exec
UPDATE users
SET last_seen = ?
WHERE id = ?;
