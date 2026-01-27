-- name: CreateForm :one
INSERT INTO forms (
    name, description, user_id, status, schema, settings, share_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetFormByID :one
SELECT * FROM forms
WHERE id = $1;

-- name: GetFormByShareURL :one
SELECT * FROM forms
WHERE share_url = $1;

-- name: ListFormsByUserID :many
SELECT * FROM forms
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListPublishedFormsByUserID :many
SELECT * FROM forms
WHERE user_id = $1 AND status = 'published'
ORDER BY created_at DESC;

-- name: UpdateForm :one
UPDATE forms
SET 
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    schema = COALESCE($4, schema),
    settings = COALESCE($5, settings),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateFormStatus :one
UPDATE forms
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateFormShareURL :one
UPDATE forms
SET share_url = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteForm :exec
DELETE FROM forms
WHERE id = $1;

-- name: CountFormsByUserID :one
SELECT COUNT(*) FROM forms
WHERE user_id = $1;
