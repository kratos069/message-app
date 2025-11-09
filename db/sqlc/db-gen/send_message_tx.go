package db

import (
	"context"

	"github.com/google/uuid"
)

// ============================================
// TRANSACTION 2:Inserts message + updates conversation timestamp
// Important: Keeps conversation list sorted correctly
// ============================================

type SendMessageTxParams struct {
	ConversationID   uuid.UUID
	SenderID         uuid.UUID
	EncryptedContent string
	ClientMessageID  *string
}

type SendMessageTxResult struct {
	Message      Message
	Conversation Conversation
}

// SendMessageTx creates a message and updates the conversation timestamp atomically
func (store *SQLStore) SendMessageTx(ctx context.Context, arg SendMessageTxParams) (SendMessageTxResult, error) {
	var result SendMessageTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// Create the message
		result.Message, err = q.CreateMessage(ctx, CreateMessageParams{
			ConversationID:   arg.ConversationID,
			SenderID:         arg.SenderID,
			EncryptedContent: arg.EncryptedContent,
			ClientMessageID:  *arg.ClientMessageID,
		})
		if err != nil {
			return err
		}

		// Update conversation timestamp
		err = q.UpdateConversationTimestamp(ctx, arg.ConversationID)
		if err != nil {
			return err
		}

		// Get updated conversation
		result.Conversation, err = q.GetConversationByID(ctx, arg.ConversationID)
		if err != nil {
			return err
		}

		return nil
	})

	return result, err
}
