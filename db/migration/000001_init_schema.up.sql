CREATE TABLE "Users" (
  "id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "username" varchar(50) UNIQUE NOT NULL,
  "email" varchar(255) UNIQUE NOT NULL,
  "password_hash" varchar(255) NOT NULL,
  "public_key" text NOT NULL,
  "profile_picture_url" varchar(500),
  "is_online" boolean DEFAULT false,
  "last_seen_at" timestamp,
  "created_at" timestamp DEFAULT (now())
);

CREATE TABLE "Conversations" (
  "conversations_id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "created_at" timestamp DEFAULT (now()),
  "updated_at" timestamp DEFAULT (now())
);

CREATE TABLE "ConversationParticipants" (
  "conversation_participants_id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "conversation_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "last_read_at" timestamp,
  "joined_at" timestamp DEFAULT (now())
);

CREATE TABLE "Messages" (
  "messages_id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "conversation_id" uuid NOT NULL,
  "sender_id" uuid NOT NULL,
  "encrypted_content" text NOT NULL,
  "client_message_id" varchar(100) NOT NULL,
  "sent_at" timestamp DEFAULT (now())
);

CREATE TABLE "TypingIndicators" (
  "typing_indicators_id" uuid PRIMARY KEY DEFAULT (gen_random_uuid()),
  "conversation_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "started_at" timestamp DEFAULT (now())
);

CREATE INDEX ON "Users" ("username");

CREATE INDEX ON "Users" ("email");

CREATE INDEX ON "Users" ("is_online");

CREATE INDEX ON "Conversations" ("updated_at");

CREATE UNIQUE INDEX ON "ConversationParticipants" ("conversation_id", "user_id");

CREATE INDEX ON "ConversationParticipants" ("user_id");

CREATE INDEX ON "ConversationParticipants" ("conversation_id");

CREATE INDEX ON "Messages" ("conversation_id", "sent_at");

CREATE INDEX ON "Messages" ("sender_id");

CREATE UNIQUE INDEX idx_messages_client_id_unique 
ON "Messages" ("client_message_id") 
WHERE "client_message_id" IS NOT NULL;

CREATE UNIQUE INDEX ON "TypingIndicators" ("conversation_id", "user_id");

CREATE INDEX ON "TypingIndicators" ("conversation_id");

COMMENT ON COLUMN "Users"."public_key" IS 'For E2E encryption';

COMMENT ON COLUMN "ConversationParticipants"."last_read_at" IS 'For read receipts';

COMMENT ON COLUMN "Messages"."encrypted_content" IS 'E2E encrypted message';

COMMENT ON COLUMN "Messages"."client_message_id" IS 'For idempotency/deduplication';

ALTER TABLE "ConversationParticipants" ADD FOREIGN KEY ("conversation_id") REFERENCES "Conversations" ("conversations_id");

ALTER TABLE "ConversationParticipants" ADD FOREIGN KEY ("user_id") REFERENCES "Users" ("id");

ALTER TABLE "Messages" ADD FOREIGN KEY ("conversation_id") REFERENCES "Conversations" ("conversations_id");

ALTER TABLE "Messages" ADD FOREIGN KEY ("sender_id") REFERENCES "Users" ("id");

ALTER TABLE "TypingIndicators" ADD FOREIGN KEY ("conversation_id") REFERENCES "Conversations" ("conversations_id");

ALTER TABLE "TypingIndicators" ADD FOREIGN KEY ("user_id") REFERENCES "Users" ("id");
