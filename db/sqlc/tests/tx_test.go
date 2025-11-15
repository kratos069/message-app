package db

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kratos069/message-app/db/sqlc/db-gen"
	"github.com/kratos069/message-app/util"
	"github.com/stretchr/testify/require"
)

// ============================================
// TEST: CreateConversationTx
// ============================================

func TestCreateConversationTx(t *testing.T) {
	ctx := context.Background()

	// Create two test users
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)

	// Test successful conversation creation
	result, err := testStore.CreateConversationTx(ctx,
		db.CreateConversationTxParams{
			User1ID: user1.ID,
			User2ID: user2.ID,
		})

	require.NoError(t, err)
	require.NotEmpty(t, result.Conversation.ConversationsID)
	require.NotEmpty(t, result.Conversation.CreatedAt)
	require.NotEmpty(t, result.Conversation.UpdatedAt)

	// Verify participant 1
	require.NotEmpty(t, result.Participant1.ConversationParticipantsID)
	require.Equal(t, result.Conversation.ConversationsID, result.Participant1.ConversationID)
	require.Equal(t, user1.ID, result.Participant1.UserID)
	require.NotEmpty(t, result.Participant1.JoinedAt)

	// Verify participant 2
	require.NotEmpty(t, result.Participant2.ConversationParticipantsID)
	require.Equal(t, result.Conversation.ConversationsID, result.Participant2.ConversationID)
	require.Equal(t, user2.ID, result.Participant2.UserID)
	require.NotEmpty(t, result.Participant2.JoinedAt)

	// Verify both participants are in database
	participants, err := testStore.GetConversationParticipants(
		ctx, result.Conversation.ConversationsID)
	require.NoError(t, err)
	require.Len(t, participants, 2)
}

// ============================================
// TEST: SendMessageTx
// ============================================

func TestSendMessageTx(t *testing.T) {
	ctx := context.Background()

	// Setup: Create conversation with participants
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)

	convResult, err := testStore.CreateConversationTx(
		ctx, db.CreateConversationTxParams{
			User1ID: user1.ID,
			User2ID: user2.ID,
		})
	require.NoError(t, err)

	// Test sending message
	clientMsgID := util.RandomClientMessageID()
	encryptedContent := util.RandomEncryptedContent()

	result, err := testStore.SendMessageTx(ctx,
		db.SendMessageTxParams{
			ConversationID:   convResult.Conversation.ConversationsID,
			SenderID:         user1.ID,
			EncryptedContent: encryptedContent,
			ClientMessageID:  &clientMsgID,
		})

	require.NoError(t, err)
	require.NotEmpty(t, result.Message.MessagesID)
	require.Equal(t, convResult.Conversation.ConversationsID, result.Message.ConversationID)
	require.Equal(t, user1.ID, result.Message.SenderID)
	require.Equal(t, encryptedContent, result.Message.EncryptedContent)
	require.NotEmpty(t, result.Message.SentAt)

	// Verify conversation timestamp was updated
	require.NotEmpty(t, result.Conversation.UpdatedAt)
	require.True(t, result.Conversation.UpdatedAt.Time.After(
		convResult.Conversation.CreatedAt.Time) ||
		result.Conversation.UpdatedAt.Time.Equal(
			convResult.Conversation.CreatedAt.Time))
}

// Idempotency → means: If you send the same message twice
// with the same ID, only one should be saved
func TestSendMessageTxIdempotency(t *testing.T) {
	ctx := context.Background()

	// Setup
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)

	convResult, err := testStore.CreateConversationTx(ctx,
		db.CreateConversationTxParams{
			User1ID: user1.ID,
			User2ID: user2.ID,
		})
	require.NoError(t, err)

	// Send message with client message ID
	clientMsgID := util.RandomClientMessageID()
	encryptedContent := util.RandomEncryptedContent()

	result1, err := testStore.SendMessageTx(ctx,
		db.SendMessageTxParams{
			ConversationID:   convResult.Conversation.ConversationsID,
			SenderID:         user1.ID,
			EncryptedContent: encryptedContent,
			ClientMessageID:  &clientMsgID,
		})
	require.NoError(t, err)

	// Check if message with this client_message_id already exists
	existingMsg, err := testStore.GetMessageByClientID(
		ctx, clientMsgID)
	require.NoError(t, err)
	require.Equal(t, result1.Message.MessagesID, existingMsg.MessagesID)

	// UNIQUE constraint on client_message_id
	result2, err := testStore.SendMessageTx(ctx,
		db.SendMessageTxParams{
			ConversationID:   convResult.Conversation.ConversationsID,
			SenderID:         user1.ID,
			EncryptedContent: "different content",
			ClientMessageID:  &clientMsgID,
		})

	// Should fail due to unique constraint on
	// client_message_id, even when content is different
	require.Error(t, err)
	require.Empty(t, result2.Message.MessagesID)

	// Verify only ONE message exists with this client_message_id
	// It fetches all messages in that conversation.
	// Confirms there’s only one message total —
	// meaning the duplicate wasn’t added.
	messages, err := testStore.GetConversationMessages(ctx,
		db.GetConversationMessagesParams{
			ConversationID: convResult.Conversation.ConversationsID,
			Limit:          10,
			Offset:         0,
		})
	require.NoError(t, err)
	require.Len(t, messages, 1)
}

func TestSendMessageTxInvalidConversation(t *testing.T) {
	ctx := context.Background()

	user1 := createRandomUser(t)
	invalidConvID := util.RandomInt(1, 1000)
	clientMsgID := util.RandomClientMessageID()

	// Should fail with invalid conversation ID
	result, err := testStore.SendMessageTx(ctx,
		db.SendMessageTxParams{
			ConversationID:   invalidConvID,
			SenderID:         user1.ID,
			EncryptedContent: util.RandomEncryptedContent(),
			ClientMessageID:  &clientMsgID,
		})

	require.Error(t, err)
	require.Empty(t, result.Message.MessagesID)
}

// ============================================
// TEST: GetOrCreateDirectConversationTx
// ============================================

func TestGetOrCreateDirectConversationTxCreate(t *testing.T) {
	ctx := context.Background()

	user1 := createRandomUser(t)
	user2 := createRandomUser(t)

	// First call should create new conversation
	result, err := testStore.GetOrCreateDirectConversationTx(
		ctx, db.GetOrCreateDirectConversationTxParams{
			User1ID: user1.ID,
			User2ID: user2.ID,
		})

	require.NoError(t, err)
	require.NotEmpty(t, result.Conversation.ConversationsID)
	require.True(t, result.IsNew)

	// Verify participants were added
	participants, err := testStore.GetConversationParticipants(
		ctx, result.Conversation.ConversationsID)
	require.NoError(t, err)
	require.Len(t, participants, 2)
}

func TestGetOrCreateDirectConversationTxGet(t *testing.T) {
	ctx := context.Background()

	user1 := createRandomUser(t)
	user2 := createRandomUser(t)

	// Create conversation first
	result1, err := testStore.GetOrCreateDirectConversationTx(
		ctx, db.GetOrCreateDirectConversationTxParams{
			User1ID: user1.ID,
			User2ID: user2.ID,
		})
	require.NoError(t, err)
	require.True(t, result1.IsNew)

	// Second call should return existing conversation
	result2, err := testStore.GetOrCreateDirectConversationTx(
		ctx, db.GetOrCreateDirectConversationTxParams{
			User1ID: user1.ID,
			User2ID: user2.ID,
		})
	require.NoError(t, err)
	require.False(t, result2.IsNew)
	require.Equal(t, result1.Conversation.ConversationsID, result2.Conversation.ConversationsID)

	// Order shouldn't matter
	result3, err := testStore.GetOrCreateDirectConversationTx(
		ctx, db.GetOrCreateDirectConversationTxParams{
			User1ID: user2.ID,
			User2ID: user1.ID,
		})
	require.NoError(t, err)
	require.False(t, result3.IsNew)
	require.Equal(t, result1.Conversation.ConversationsID, result3.Conversation.ConversationsID)
}

// 10 simultaneous requests chat between the same two users
// ➜ Only one conversation should be created
// ➜ All other calls should return that same conversation
// ➜ No duplicates, no extra participants
func TestGetOrCreateDirectConversationTxConcurrency(t *testing.T) {
	ctx := context.Background()

	user1 := createRandomUser(t)
	user2 := createRandomUser(t)

	// Simulate concurrent requests
	const numGoroutines = 10
	resultChan := make(
		chan db.GetOrCreateDirectConversationTxResult,
		numGoroutines)
	errChan := make(chan error, numGoroutines)

	for range numGoroutines {
		go func() {
			result, err := testStore.GetOrCreateDirectConversationTx(
				ctx, db.GetOrCreateDirectConversationTxParams{
					User1ID: user1.ID,
					User2ID: user2.ID,
				})
			if err != nil {
				errChan <- err
				return
			}
			resultChan <- result
		}()
	}

	// Collect results
	var results []db.GetOrCreateDirectConversationTxResult
	for range numGoroutines {
		select {
		case result := <-resultChan:
			results = append(results, result)
		case err := <-errChan:
			// Some errors expected due to serialization conflicts
			// But at least one should succeed
			t.Logf("Concurrent request error (expected): %v", err)
		}
	}

	// Verify all successful results point to the same conversation
	require.NotEmpty(
		t, results, "at least one request should succeed")
	firstConvID := results[0].Conversation.ConversationsID

	for _, result := range results {
		require.Equal(t, firstConvID, result.Conversation.ConversationsID)
	}

	// Verify only one conversation was created
	participants, err := testStore.GetConversationParticipants(
		ctx, firstConvID)
	require.NoError(t, err)
	require.Len(t, participants, 2)
}

// ============================================
// TEST: MarkMessagesAsReadTx
// ============================================

func TestMarkMessagesAsReadTx(t *testing.T) {
	ctx := context.Background()

	// Setup: Create conversation and send messages
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)

	convResult, err := testStore.CreateConversationTx(
		ctx, db.CreateConversationTxParams{
			User1ID: user1.ID,
			User2ID: user2.ID,
		})
	require.NoError(t, err)

	// Send some messages from user1
	for range 3 {
		clientMsgID := util.RandomClientMessageID()
		_, err := testStore.SendMessageTx(ctx,
			db.SendMessageTxParams{
				ConversationID:   convResult.Conversation.ConversationsID,
				SenderID:         user1.ID,
				EncryptedContent: util.RandomEncryptedContent(),
				ClientMessageID:  &clientMsgID,
			})
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get initial unread count for user2
	unreadBefore, err := testStore.GetUnreadCount(
		ctx, db.GetUnreadCountParams{
			ConversationID: convResult.Conversation.ConversationsID,
			UserID:         user2.ID,
		})
	require.NoError(t, err)
	require.Equal(t, int64(3), unreadBefore)

	// Mark messages as read
	err = testStore.MarkMessagesAsReadTx(ctx,
		db.MarkMessagesAsReadTxParams{
			ConversationID: convResult.Conversation.ConversationsID,
			UserID:         user2.ID,
		})
	require.NoError(t, err)

	// Verify unread count is now 0
	unreadAfter, err := testStore.GetUnreadCount(
		ctx, db.GetUnreadCountParams{
			ConversationID: convResult.Conversation.ConversationsID,
			UserID:         user2.ID,
		})
	require.NoError(t, err)
	require.Equal(t, int64(0), unreadAfter)

	// Verify last_read_at was updated
	participants, err := testStore.GetConversationParticipants(
		ctx, convResult.Conversation.ConversationsID)
	require.NoError(t, err)

	// Find user2's participant record
	var user2LastReadAt pgtype.Timestamp
	found := false
	for _, p := range participants {
		if p.ID == user2.ID { // p.ID is the user's ID from the JOIN
			user2LastReadAt = p.LastReadAt
			found = true
			break
		}
	}

	require.True(t, found, "user2 should be a participant")
	require.True(t, user2LastReadAt.Valid, "last_read_at should be set")
	require.False(t, user2LastReadAt.Time.IsZero(), "last_read_at should not be zero")
}

func TestMarkMessagesAsReadTxInvalidConversation(t *testing.T) {
	ctx := context.Background()

	user := createRandomUser(t)
	invalidConvID := util.RandomInt(1, 100)

	// Should not error, but won't update anything
	err := testStore.MarkMessagesAsReadTx(ctx,
		db.MarkMessagesAsReadTxParams{
			ConversationID: invalidConvID,
			UserID:         user.ID,
		})
	require.NoError(t, err)
}

func TestMarkMessagesAsReadTxNewMessages(t *testing.T) {
	ctx := context.Background()

	// Setup
	user1 := createRandomUser(t)
	user2 := createRandomUser(t)

	convResult, err := testStore.CreateConversationTx(
		ctx, db.CreateConversationTxParams{
			User1ID: user1.ID,
			User2ID: user2.ID,
		})
	require.NoError(t, err)

	// Send 2 messages
	for range 2 {
		clientMsgID := util.RandomClientMessageID()
		_, err := testStore.SendMessageTx(ctx,
			db.SendMessageTxParams{
				ConversationID:   convResult.Conversation.ConversationsID,
				SenderID:         user1.ID,
				EncryptedContent: util.RandomEncryptedContent(),
				ClientMessageID:  &clientMsgID,
			})
		require.NoError(t, err)
	}

	// Mark as read
	err = testStore.MarkMessagesAsReadTx(ctx,
		db.MarkMessagesAsReadTxParams{
			ConversationID: convResult.Conversation.ConversationsID,
			UserID:         user2.ID,
		})
	require.NoError(t, err)

	// Send new message
	time.Sleep(100 * time.Millisecond)
	clientMsgID := util.RandomClientMessageID()
	_, err = testStore.SendMessageTx(ctx,
		db.SendMessageTxParams{
			ConversationID:   convResult.Conversation.ConversationsID,
			SenderID:         user1.ID,
			EncryptedContent: util.RandomEncryptedContent(),
			ClientMessageID:  &clientMsgID,
		})
	require.NoError(t, err)

	// Verify unread count is now 1
	unread, err := testStore.GetUnreadCount(ctx,
		db.GetUnreadCountParams{
			ConversationID: convResult.Conversation.ConversationsID,
			UserID:         user2.ID,
		})
	require.NoError(t, err)
	require.Equal(t, int64(1), unread)
}
