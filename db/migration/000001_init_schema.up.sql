-- ============================================
-- 1. USERS TABLE
-- ============================================
CREATE TABLE "Users" (
  "id" bigserial PRIMARY KEY,
  "username" varchar(50) UNIQUE NOT NULL,
  "email" varchar(255) UNIQUE NOT NULL,
  "is_email_verified" boolean NOT NULL DEFAULT false,
  "password_hash" varchar(255) NOT NULL,
  "profile_picture_url" varchar(500),
  "is_online" boolean NOT NULL DEFAULT false,
  "last_seen_at" timestamptz,
  "role" varchar(20) NOT NULL DEFAULT 'customer',
  "is_banned" boolean NOT NULL DEFAULT false,
  "banned_at" timestamptz,
  "banned_reason" text,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

-- Users indexes
CREATE INDEX idx_users_username ON "Users" ("username");
CREATE INDEX idx_users_email ON "Users" ("email");
CREATE INDEX idx_users_is_online ON "Users" ("is_online");
CREATE INDEX idx_users_is_banned ON "Users" ("is_banned") WHERE is_banned = true;
CREATE INDEX idx_users_role ON "Users" ("role");

-- ============================================
-- 2. SESSIONS TABLE
-- ============================================
CREATE TABLE "Sessions" (
  "id" uuid PRIMARY KEY,
  "username" varchar(50) NOT NULL,
  "refresh_token" text NOT NULL,
  "user_agent" text NOT NULL,
  "client_ip" varchar(50) NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "expired_at" timestamptz NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now())
);

-- Sessions indexes
CREATE INDEX idx_sessions_username ON "Sessions" ("username");
CREATE INDEX idx_sessions_expired_at ON "Sessions" ("expired_at");

-- Sessions foreign key
ALTER TABLE "Sessions" ADD FOREIGN KEY ("username") REFERENCES "Users" ("username") ON DELETE CASCADE;

-- ============================================
-- 3. VERIFY EMAILS TABLE
-- ============================================
CREATE TABLE "Verify_Emails" (
  "email_id" bigserial PRIMARY KEY,
  "username" varchar(50) NOT NULL,
  "email" varchar(255) NOT NULL,
  "secret_code" varchar(100) NOT NULL,
  "is_used" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "expired_at" timestamptz NOT NULL DEFAULT (now() + interval '15 minutes')
);

-- Verify_Emails indexes
CREATE INDEX idx_verify_emails_username ON "Verify_Emails" ("username");
CREATE INDEX idx_verify_emails_email ON "Verify_Emails" ("email");
CREATE INDEX idx_verify_emails_secret_code ON "Verify_Emails" ("secret_code");

-- Verify_Emails foreign key
ALTER TABLE "Verify_Emails" ADD FOREIGN KEY ("username") REFERENCES "Users" ("username") ON DELETE CASCADE;

-- ============================================
-- 4. CONVERSATIONS TABLE
-- ============================================
CREATE TABLE "Conversations" (
  "conversations_id" bigserial PRIMARY KEY,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT (now())
);

-- Conversations indexes
CREATE INDEX idx_conversations_updated_at ON "Conversations" ("updated_at");

-- ============================================
-- 5. CONVERSATION PARTICIPANTS TABLE
-- ============================================
CREATE TABLE "ConversationParticipants" (
  "conversation_participants_id" bigserial PRIMARY KEY,
  "conversation_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "last_read_at" timestamptz,
  "joined_at" timestamptz NOT NULL DEFAULT (now())
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

-- ConversationParticipants foreign keys
ALTER TABLE "ConversationParticipants" 
  ADD FOREIGN KEY ("conversation_id") 
  REFERENCES "Conversations" ("conversations_id") 
  ON DELETE CASCADE;

ALTER TABLE "ConversationParticipants" 
  ADD FOREIGN KEY ("user_id") 
  REFERENCES "Users" ("id") 
  ON DELETE CASCADE;

-- ============================================
-- 6. MESSAGES TABLE
-- ============================================
CREATE TABLE "Messages" (
  "messages_id" bigserial PRIMARY KEY,
  "conversation_id" bigint NOT NULL,
  "sender_id" bigint NOT NULL,
  "encrypted_content" text NOT NULL,
  "client_message_id" varchar(100) NOT NULL,
  "sent_at" timestamptz NOT NULL DEFAULT (now())
);

-- Messages indexes
CREATE INDEX idx_messages_conversation_sent 
  ON "Messages" ("conversation_id", "sent_at" DESC);
CREATE INDEX idx_messages_sender_id 
  ON "Messages" ("sender_id");
CREATE UNIQUE INDEX idx_messages_client_id_unique 
  ON "Messages" ("client_message_id");

-- Comments
COMMENT ON COLUMN "Messages"."encrypted_content" IS 'E2E encrypted message';
COMMENT ON COLUMN "Messages"."client_message_id" IS 'For idempotency/deduplication';

-- Messages foreign keys
ALTER TABLE "Messages" 
  ADD FOREIGN KEY ("conversation_id") 
  REFERENCES "Conversations" ("conversations_id") 
  ON DELETE CASCADE;

ALTER TABLE "Messages" 
  ADD FOREIGN KEY ("sender_id") 
  REFERENCES "Users" ("id") 
  ON DELETE CASCADE;

-- ============================================
-- 7. TYPING INDICATORS TABLE
-- ============================================
CREATE TABLE "TypingIndicators" (
  "typing_indicators_id" bigserial PRIMARY KEY,
  "conversation_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "started_at" timestamptz NOT NULL DEFAULT (now())
);

-- TypingIndicators indexes
CREATE UNIQUE INDEX idx_typing_indicators_unique 
  ON "TypingIndicators" ("conversation_id", "user_id");
CREATE INDEX idx_typing_indicators_conversation_id 
  ON "TypingIndicators" ("conversation_id");

-- TypingIndicators foreign keys
ALTER TABLE "TypingIndicators" 
  ADD FOREIGN KEY ("conversation_id") 
  REFERENCES "Conversations" ("conversations_id") 
  ON DELETE CASCADE;

ALTER TABLE "TypingIndicators" 
  ADD FOREIGN KEY ("user_id") 
  REFERENCES "Users" ("id") 
  ON DELETE CASCADE;
