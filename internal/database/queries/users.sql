-- name: CreateUser :one
INSERT INTO users (
    email, password, name
) VALUES (
    $1, $2, $3
)
RETURNING id, email, password, name, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, email, password, name, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, password, name, created_at, updated_at
FROM users
WHERE email = $1;

-- name: ListUsers :many
SELECT id, email, password, name, created_at, updated_at
FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE users
SET name = $1, email = $2, updated_at = NOW()
WHERE id = $3
RETURNING id, email, password, name, created_at, updated_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
