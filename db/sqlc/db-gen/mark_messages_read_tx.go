package db

import (
	"context"
)

// ============================================
// TRANSACTION 4: Mark All Messages as Read
// Low risk: Could work without transaction,
// but included for consistency
// ============================================

type MarkMessagesAsReadTxParams struct {
	ConversationID int64
	UserID         int64
}

// MarkMessagesAsReadTx marks all messages in a conversation as read for a user
func (store *SQLStore) MarkMessagesAsReadTx(
	ctx context.Context, arg MarkMessagesAsReadTxParams) error {
	return store.execTx(ctx, func(q *Queries) error {
		// Update last_read_at to current timestamp
		return q.UpdateLastReadAt(ctx, UpdateLastReadAtParams{
			ConversationID: arg.ConversationID,
			UserID:         arg.UserID,
		})
	})
}
