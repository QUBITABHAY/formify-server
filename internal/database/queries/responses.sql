-- name: CreateResponse :one
INSERT INTO responses (
    form_id, data, meta
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetResponseByID :one
SELECT * FROM responses
WHERE id = $1;

-- name: ListResponsesByFormID :many
SELECT * FROM responses
WHERE form_id = $1
ORDER BY created_at DESC;

-- name: ListResponsesByFormIDPaginated :many
SELECT * FROM responses
WHERE form_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: DeleteResponse :exec
DELETE FROM responses
WHERE id = $1;

-- name: DeleteResponsesByFormID :exec
DELETE FROM responses
WHERE form_id = $1;

-- name: CountResponsesByFormID :one
SELECT COUNT(*) FROM responses
WHERE form_id = $1;
