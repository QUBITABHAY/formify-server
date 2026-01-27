-- name: CreateUser :one
INSERT INTO users (
    name, email, password
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: CreateOAuthUser :one
INSERT INTO users (
    name, email, password, oauth_provider, oauth_id, is_oauth
) VALUES (
    $1, $2, '', $3, $4, true
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUserByOAuthID :one
SELECT * FROM users
WHERE oauth_provider = $1 AND oauth_id = $2;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY created_at DESC;

-- name: UpdateUser :one
UPDATE users
SET 
    name = COALESCE($2, name),
    email = COALESCE($3, email),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
