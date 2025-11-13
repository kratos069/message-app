package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// ============================================
// TRANSACTION 3: Get or Create Direct Conversation
// Uses SERIALIZABLE isolation level -> Prevents
// duplicate conversations between same users
// Race-condition safe for concurrent requests
// ============================================

type GetOrCreateDirectConversationTxParams struct {
	User1ID int64
	User2ID int64
}

type GetOrCreateDirectConversationTxResult struct {
	Conversation Conversation
	IsNew        bool
}

// GetOrCreateDirectConversationTx atomically
// gets or creates a direct conversation
// Uses SERIALIZABLE isolation to prevent duplicate conversations
func (store *SQLStore) GetOrCreateDirectConversationTx(
	ctx context.Context,
	arg GetOrCreateDirectConversationTxParams) (
	GetOrCreateDirectConversationTxResult, error) {
	var result GetOrCreateDirectConversationTxResult

	// Use serializable isolation level to prevent race conditions
	tx, err := store.connPool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return result, err
	}
	defer tx.Rollback(ctx)

	q := New(tx)

	// Try to find existing conversation
	conversationID, err := q.FindDirectConversation(ctx, FindDirectConversationParams{
		UserID:   arg.User1ID,
		UserID_2: arg.User2ID,
	})

	if err == pgx.ErrNoRows {
		// Conversation doesn't exist, create it
		var conversation Conversation
		conversation, err = q.CreateConversation(ctx)
		if err != nil {
			return result, err
		}

		// Add both participants
		_, err = q.AddParticipantToConversation(ctx, AddParticipantToConversationParams{
			ConversationID: conversation.ConversationsID,
			UserID:         arg.User1ID,
		})
		if err != nil {
			return result, err
		}

		_, err = q.AddParticipantToConversation(ctx, AddParticipantToConversationParams{
			ConversationID: conversation.ConversationsID,
			UserID:         arg.User2ID,
		})
		if err != nil {
			return result, err
		}

		result.Conversation = conversation
		result.IsNew = true
	} else if err != nil {
		return result, err
	} else {
		// Conversation exists
		result.Conversation, err = q.GetConversationByID(ctx, conversationID)
		if err != nil {
			return result, err
		}
		result.IsNew = false
	}

	if err := tx.Commit(ctx); err != nil {
		return result, err
	}

	return result, nil
}
