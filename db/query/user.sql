-- name: CreateUser :one
INSERT INTO "Users" (
  username,
  email,
  password_hash,
  profile_picture_url
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM "Users"
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM "Users"
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT * FROM "Users"
WHERE email = $1;

-- name: UpdateUserOnlineStatus :exec
UPDATE "Users"
SET is_online = $2,
    last_seen_at = CASE WHEN $2 = false THEN now() ELSE last_seen_at END
WHERE id = $1;

-- name: UpdateUserProfile :exec
UPDATE "Users"
SET profile_picture_url = COALESCE($2, profile_picture_url),
    username = COALESCE($3, username)
WHERE id = $1;

-- name: GetOnlineUsers :many
SELECT id, username, profile_picture_url, last_seen_at
FROM "Users"
WHERE is_online = true
ORDER BY last_seen_at DESC;

-- name: SearchUsersByUsername :many
SELECT id, username, email, profile_picture_url, is_online
FROM "Users"
WHERE username ILIKE $1
LIMIT $2;

-- name: GetAllUsers :many
SELECT * FROM "Users" 
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: BanUser :exec
UPDATE "Users"
SET is_banned = true,
    banned_at = now(),
    banned_reason = $2
WHERE id = $1;

-- name: UnbanUser :exec
UPDATE "Users"
SET is_banned = false,
    banned_at = NULL,
    banned_reason = NULL
WHERE id = $1;