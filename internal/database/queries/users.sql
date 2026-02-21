-- name: CreateUser :one
INSERT INTO users (
    name, email, password
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: CreateOAuthUser :one
INSERT INTO users (
    name, email, password, oauth_provider, oauth_id, is_oauth, google_access_token, google_refresh_token, google_token_expiry
) VALUES (
    $1, $2, '', $3, $4, true, $5, $6, $7
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

-- name: UpdateUserOAuthTokens :one
UPDATE users
SET 
    google_access_token = $2,
    google_refresh_token = COALESCE($3, google_refresh_token),
    google_token_expiry = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
