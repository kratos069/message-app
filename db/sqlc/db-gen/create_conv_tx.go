package db

import (
	"context"
)

// ============================================
// TRANSACTION 1: Create Conv with Participants
// ============================================

type CreateConversationTxParams struct {
	User1ID int64
	User2ID int64
}

type CreateConversationTxResult struct {
	Conversation Conversation
	Participant1 ConversationParticipant
	Participant2 ConversationParticipant
}

// CreateConversationTx creates a new conversation and adds both participants atomically
func (store *SQLStore) CreateConversationTx(
	ctx context.Context,
	arg CreateConversationTxParams) (
	CreateConversationTxResult, error) {
	var result CreateConversationTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// Create conversation
		result.Conversation, err = q.CreateConversation(ctx)
		if err != nil {
			return err
		}

		// Add first participant
		result.Participant1, err = q.AddParticipantToConversation(ctx, AddParticipantToConversationParams{
			ConversationID: result.Conversation.ConversationsID,
			UserID:         arg.User1ID,
		})
		if err != nil {
			return err
		}

		// Add second participant
		result.Participant2, err = q.AddParticipantToConversation(ctx, AddParticipantToConversationParams{
			ConversationID: result.Conversation.ConversationsID,
			UserID:         arg.User2ID,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
