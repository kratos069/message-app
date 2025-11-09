-- name: CreateConversation :one
INSERT INTO "Conversations" DEFAULT VALUES
RETURNING *;

-- name: GetConversationByID :one
SELECT * FROM "Conversations"
WHERE conversations_id = $1;

-- name: UpdateConversationTimestamp :exec
UPDATE "Conversations"
SET updated_at = now()
WHERE conversations_id = $1;

-- name: GetUserConversations :many
SELECT 
  c.conversations_id,
  c.created_at,
  c.updated_at,
  COUNT(CASE WHEN m.sent_at > cp.last_read_at THEN 1 END) as unread_count
FROM "Conversations" c
INNER JOIN "ConversationParticipants" cp ON c.conversations_id = cp.conversation_id
LEFT JOIN "Messages" m ON c.conversations_id = m.conversation_id
WHERE cp.user_id = $1
GROUP BY c.conversations_id, c.created_at, c.updated_at, cp.last_read_at
ORDER BY c.updated_at DESC
LIMIT $2 OFFSET $3;

-- name: GetConversationWithParticipants :many
SELECT 
  c.conversations_id,
  c.created_at,
  c.updated_at,
  u.id as participant_id,
  u.username as participant_username,
  u.profile_picture_url as participant_avatar,
  u.is_online as participant_online,
  cp.last_read_at,
  cp.joined_at
FROM "Conversations" c
INNER JOIN "ConversationParticipants" cp ON c.conversations_id = cp.conversation_id
INNER JOIN "Users" u ON cp.user_id = u.id
WHERE c.conversations_id = $1;

-- name: FindDirectConversation :one
SELECT cp1.conversation_id
FROM "ConversationParticipants" cp1
INNER JOIN "ConversationParticipants" cp2 
  ON cp1.conversation_id = cp2.conversation_id
WHERE cp1.user_id = $1 
  AND cp2.user_id = $2
  AND cp1.user_id != cp2.user_id
LIMIT 1;