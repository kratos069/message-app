-- name: GetTotalUsers :one
SELECT COUNT(*) as total
FROM "Users";

-- name: GetTotalConversations :one
SELECT COUNT(*) as total
FROM "Conversations";

-- name: GetTotalMessages :one
SELECT COUNT(*) as total
FROM "Messages";

-- name: GetOnlineUsersCount :one
SELECT COUNT(*) as total
FROM "Users"
WHERE is_online = true;