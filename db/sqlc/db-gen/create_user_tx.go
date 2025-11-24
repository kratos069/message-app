package db

import "context"

type CreateUserTxParams struct {
	CreateUserParams
	AfterCreate func(user User) error
}

type CreateUserTxResults struct {
	User User
}

func (store *SQLStore) CreateUserTx(ctx context.Context,
	arg CreateUserTxParams) (CreateUserTxResults, error) {
	var result CreateUserTxResults

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.User, err = q.CreateUser(ctx, arg.CreateUserParams)
		if err != nil {
			return err
		}

		err = arg.AfterCreate(result.User)
		return err
	})

	return result, err
}