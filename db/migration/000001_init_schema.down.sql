-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS "TypingIndicators" CASCADE;
DROP TABLE IF EXISTS "Messages" CASCADE;
DROP TABLE IF EXISTS "ConversationParticipants" CASCADE;
DROP TABLE IF EXISTS "Conversations" CASCADE;
DROP TABLE IF EXISTS "Verify_Emails" CASCADE;
DROP TABLE IF EXISTS "Sessions" CASCADE;
DROP TABLE IF EXISTS "Users" CASCADE;
