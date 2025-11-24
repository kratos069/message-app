package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store can do both queries and transcations
type Store interface {
	Querier
	// Transactions
	CreateConversationTx(ctx context.Context,
		arg CreateConversationTxParams) (
		CreateConversationTxResult, error)
	SendMessageTx(ctx context.Context, arg SendMessageTxParams) (
		SendMessageTxResult, error)
	GetOrCreateDirectConversationTx(
		ctx context.Context,
		arg GetOrCreateDirectConversationTxParams) (
		GetOrCreateDirectConversationTxResult, error)
	MarkMessagesAsReadTx(
		ctx context.Context, arg MarkMessagesAsReadTxParams) error
	CreateUserTx(ctx context.Context,
		arg CreateUserTxParams) (CreateUserTxResults, error)
	VerifyEmailTx(ctx context.Context,
		arg VerifyEmailTxParams) (VerifyEmailTxResults, error)
}

// SQLStore provides all funcs for SQL queries and transactions
type SQLStore struct {
	// queries only supports queries not transactions,
	// so we use it in store struct and add more functionality.
	connPool *pgxpool.Pool
	*Queries
}

func NewStore(connPool *pgxpool.Pool) Store {
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}
}
