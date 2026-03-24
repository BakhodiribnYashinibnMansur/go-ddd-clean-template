package pgxutil

import (
	"context"

	apperrors "gct/internal/shared/infrastructure/errors"

	"github.com/jackc/pgx/v5"
)

// TxBeginner is the interface for starting a transaction.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// WithTx runs fn inside a database transaction.
// It begins a transaction, calls fn, commits on success, and rolls back on error.
func WithTx(ctx context.Context, pool TxBeginner, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to begin transaction")
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return apperrors.NewRepoError(apperrors.ErrRepoDatabase, "failed to commit transaction")
	}

	return nil
}
