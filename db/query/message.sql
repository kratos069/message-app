-- name: CreateMessage :one
INSERT INTO "Messages" (
  conversation_id,
  sender_id,
  encrypted_content,
  client_message_id
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetMessageByID :one
SELECT 
  m.*,
  u.username as sender_username,
  u.profile_picture_url as sender_avatar
FROM "Messages" m
INNER JOIN "Users" u ON m.sender_id = u.id
WHERE m.messages_id = $1;

-- name: GetMessageByClientID :one
SELECT * FROM "Messages"
WHERE client_message_id = $1;

-- name: GetConversationMessages :many
SELECT 
  m.messages_id,
  m.conversation_id,
  m.sender_id,
  m.encrypted_content,
  m.sent_at,
  u.username as sender_username,
  u.profile_picture_url as sender_avatar
FROM "Messages" m
INNER JOIN "Users" u ON m.sender_id = u.id
WHERE m.conversation_id = $1
ORDER BY m.sent_at DESC
LIMIT $2 OFFSET $3;

-- name: GetMessagesSince :many
SELECT 
  m.messages_id,
  m.conversation_id,
  m.sender_id,
  m.encrypted_content,
  m.sent_at,
  u.username as sender_username,
  u.profile_picture_url as sender_avatar
FROM "Messages" m
INNER JOIN "Users" u ON m.sender_id = u.id
WHERE m.conversation_id = $1 
  AND m.sent_at > $2
ORDER BY m.sent_at ASC;

-- name: GetMessagesBefore :many
SELECT 
  m.messages_id,
  m.conversation_id,
  m.sender_id,
  m.encrypted_content,
  m.sent_at,
  u.username as sender_username,
  u.profile_picture_url as sender_avatar
FROM "Messages" m
INNER JOIN "Users" u ON m.sender_id = u.id
WHERE m.conversation_id = $1 
  AND m.sent_at < $2
ORDER BY m.sent_at DESC
LIMIT $3;

-- name: GetLatestMessage :one
SELECT 
  m.messages_id,
  m.encrypted_content,
  m.sent_at,
  u.username as sender_username
FROM "Messages" m
INNER JOIN "Users" u ON m.sender_id = u.id
WHERE m.conversation_id = $1
ORDER BY m.sent_at DESC
LIMIT 1;

-- name: DeleteMessage :exec
DELETE FROM "Messages"
WHERE messages_id = $1;

-- name: GetMessageCount :one
SELECT COUNT(*) as total_messages
FROM "Messages"
WHERE conversation_id = $1;

-- name: SearchMessages :many
SELECT 
  m.messages_id,
  m.conversation_id,
  m.sender_id,
  m.encrypted_content,
  m.sent_at,
  u.username as sender_username
FROM "Messages" m
INNER JOIN "Users" u ON m.sender_id = u.id
WHERE m.conversation_id = $1
  AND m.encrypted_content ILIKE $2
ORDER BY m.sent_at DESC
LIMIT $3;