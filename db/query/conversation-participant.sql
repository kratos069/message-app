-- name: AddParticipantToConversation :one
INSERT INTO "ConversationParticipants" (
  conversation_id,
  user_id
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetConversationParticipants :many
SELECT 
  u.id,
  u.username,
  u.email,
  u.profile_picture_url,
  u.is_online,
  cp.last_read_at,
  cp.joined_at
FROM "ConversationParticipants" cp
INNER JOIN "Users" u ON cp.user_id = u.id
WHERE cp.conversation_id = $1;

-- name: UpdateLastReadAt :exec
UPDATE "ConversationParticipants"
SET last_read_at = now()
WHERE conversation_id = $1 AND user_id = $2;

-- name: GetUnreadCount :one
SELECT COUNT(*) as unread_count
FROM "Messages" m
INNER JOIN "ConversationParticipants" cp 
  ON m.conversation_id = cp.conversation_id
WHERE cp.conversation_id = $1 
  AND cp.user_id = $2
  AND m.sent_at > COALESCE(cp.last_read_at, '1970-01-01'::timestamp);

-- name: IsUserInConversation :one
SELECT EXISTS(
  SELECT 1 FROM "ConversationParticipants"
  WHERE conversation_id = $1 AND user_id = $2
) as is_participant;

-- name: RemoveParticipantFromConversation :exec
DELETE FROM "ConversationParticipants"
WHERE conversation_id = $1 AND user_id = $2;
