-- name: SetTypingIndicator :one
INSERT INTO "TypingIndicators" (
  conversation_id,
  user_id
) VALUES (
  $1, $2
)
ON CONFLICT (conversation_id, user_id) 
DO UPDATE SET started_at = now()
RETURNING *;

-- name: RemoveTypingIndicator :exec
DELETE FROM "TypingIndicators"
WHERE conversation_id = $1 AND user_id = $2;

-- name: GetTypingUsers :many
SELECT 
  u.id,
  u.username,
  u.profile_picture_url,
  ti.started_at
FROM "TypingIndicators" ti
INNER JOIN "Users" u ON ti.user_id = u.id
WHERE ti.conversation_id = $1
  AND ti.started_at > now() - interval '10 seconds';

-- name: CleanupStaleTypingIndicators :exec
DELETE FROM "TypingIndicators"
WHERE started_at < now() - interval '10 seconds';
