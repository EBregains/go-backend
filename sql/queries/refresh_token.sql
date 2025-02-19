-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens ( token, created_at, updated_at, user_id, expires_at, revoked_at ) 
VALUES (
  $1,
  NOW(),
  NOW(),
  $2,
  $3,
  NULL
)
RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;

-- name: RevokeToken :one
update refresh_tokens 
Set revoked_at = NOW(), updated_at = NOW()
where token = $1
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1;