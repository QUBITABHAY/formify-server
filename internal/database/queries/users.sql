INSERT INTO users (
    email, password, name
) VALUES (
    $1, $2, $3
)
RETURNING id, email, password, name, created_at, updated_at;

SELECT id, email, password, name, created_at, updated_at
FROM users
WHERE id = $1;

SELECT id, email, password, name, created_at, updated_at
FROM users
WHERE email = $1;

SELECT id, email, password, name, created_at, updated_at
FROM users
ORDER BY created_at DESC;

UPDATE users
SET name = $1, email = $2, updated_at = NOW()
WHERE id = $3
RETURNING id, email, password, name, created_at, updated_at;


DELETE FROM users
WHERE id = $1;
