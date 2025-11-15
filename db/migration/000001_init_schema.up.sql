-- ============================================
-- CHAT APPLICATION DATABASE SCHEMA
-- ============================================

-- ============================================
-- 1. USERS TABLE
-- ============================================
CREATE TABLE "Users" (
  "id" bigserial PRIMARY KEY,
  "username" varchar(50) UNIQUE NOT NULL,
  "email" varchar(255) UNIQUE NOT NULL,
  "password_hash" varchar(255) NOT NULL,
  "profile_picture_url" varchar(500),
  "is_online" boolean DEFAULT false,
  "last_seen_at" timestamp,
  "role" varchar NOT NULL DEFAULT 'customer',
  "is_banned" boolean DEFAULT false,
  "banned_at" timestamp,
  "banned_reason" text,
  "created_at" timestamp DEFAULT (now())
);

-- Users indexes
CREATE INDEX idx_users_username ON "Users" ("username");
CREATE INDEX idx_users_email ON "Users" ("email");
CREATE INDEX idx_users_is_online ON "Users" ("is_online");
CREATE INDEX idx_users_is_banned ON "Users" ("is_banned") WHERE is_banned = true;
CREATE INDEX idx_users_role ON "Users" ("role");

-- ============================================
-- 2. CONVERSATIONS TABLE
-- ============================================
CREATE TABLE "Conversations" (
  "conversations_id" bigserial PRIMARY KEY,
  "created_at" timestamp DEFAULT (now()),
  "updated_at" timestamp DEFAULT (now())
);

-- Conversations indexes
CREATE INDEX idx_conversations_updated_at ON "Conversations" ("updated_at");

-- ============================================
-- 3. CONVERSATION PARTICIPANTS TABLE
-- ============================================
CREATE TABLE "ConversationParticipants" (
  "conversation_participants_id" bigserial PRIMARY KEY,
  "conversation_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "last_read_at" timestamp,
  "joined_at" timestamp DEFAULT (now())
);

-- ConversationParticipants indexes
CREATE UNIQUE INDEX idx_conversation_participants_unique 
  ON "ConversationParticipants" ("conversation_id", "user_id");
CREATE INDEX idx_conversation_participants_user_id 
  ON "ConversationParticipants" ("user_id");
CREATE INDEX idx_conversation_participants_conversation_id 
  ON "ConversationParticipants" ("conversation_id");

-- Comments
COMMENT ON COLUMN "ConversationParticipants"."last_read_at" IS 'For read receipts';

-- ============================================
-- 4. MESSAGES TABLE
-- ============================================
CREATE TABLE "Messages" (
  "messages_id" bigserial PRIMARY KEY,
  "conversation_id" bigint NOT NULL,
  "sender_id" bigint NOT NULL,
  "encrypted_content" text NOT NULL,
  "client_message_id" varchar(100) NOT NULL,
  "sent_at" timestamp DEFAULT (now())
);

-- Messages indexes
CREATE INDEX idx_messages_conversation_sent 
  ON "Messages" ("conversation_id", "sent_at");
CREATE INDEX idx_messages_sender_id 
  ON "Messages" ("sender_id");
CREATE UNIQUE INDEX idx_messages_client_id_unique 
  ON "Messages" ("client_message_id") 
  WHERE "client_message_id" IS NOT NULL;

-- Comments
COMMENT ON COLUMN "Messages"."encrypted_content" IS 'E2E encrypted message';
COMMENT ON COLUMN "Messages"."client_message_id" IS 'For idempotency/deduplication';

-- ============================================
-- 5. TYPING INDICATORS TABLE
-- ============================================
CREATE TABLE "TypingIndicators" (
  "typing_indicators_id" bigserial PRIMARY KEY,
  "conversation_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "started_at" timestamp DEFAULT (now())
);

-- TypingIndicators indexes
CREATE UNIQUE INDEX idx_typing_indicators_unique 
  ON "TypingIndicators" ("conversation_id", "user_id");
CREATE INDEX idx_typing_indicators_conversation_id 
  ON "TypingIndicators" ("conversation_id");

-- ============================================
-- FOREIGN KEY CONSTRAINTS
-- ============================================

-- ConversationParticipants foreign keys
ALTER TABLE "ConversationParticipants" 
  ADD FOREIGN KEY ("conversation_id") 
  REFERENCES "Conversations" ("conversations_id") 
  ON DELETE CASCADE;

ALTER TABLE "ConversationParticipants" 
  ADD FOREIGN KEY ("user_id") 
  REFERENCES "Users" ("id") 
  ON DELETE CASCADE;

-- Messages foreign keys
ALTER TABLE "Messages" 
  ADD FOREIGN KEY ("conversation_id") 
  REFERENCES "Conversations" ("conversations_id") 
  ON DELETE CASCADE;

ALTER TABLE "Messages" 
  ADD FOREIGN KEY ("sender_id") 
  REFERENCES "Users" ("id") 
  ON DELETE CASCADE;

-- TypingIndicators foreign keys
ALTER TABLE "TypingIndicators" 
  ADD FOREIGN KEY ("conversation_id") 
  REFERENCES "Conversations" ("conversations_id") 
  ON DELETE CASCADE;

ALTER TABLE "TypingIndicators" 
  ADD FOREIGN KEY ("user_id") 
  REFERENCES "Users" ("id") 
  ON DELETE CASCADE;