-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
gen_random_uuid(),
NOW(),
NOW(),
$1,
$2
)
RETURNING *;

-- name: GetUsers :many
SELECT * FROM users
ORDER BY created_at ASC;


-- name: GetUserByID :many
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: UpdateUserPassAndEmail :one
Update users
Set email = $1, hashed_password = $2
Where id = $3
RETURNING *;

-- name: UpdateUserToRed :one
Update users
set is_chirpy_red = true
where id = $1
RETURNING *;