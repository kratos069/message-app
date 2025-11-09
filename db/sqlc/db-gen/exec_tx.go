package db

import (
	"context"
	"fmt"
)

// executes a function within a DB Transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.connPool.Begin(ctx)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("txErr: %v, rbErr: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
