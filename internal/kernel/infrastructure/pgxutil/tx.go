package pgxutil

import (
	"context"

	shareddomain "gct/internal/kernel/domain"
	apperrors "gct/internal/kernel/infrastructure/errorx"
)

// WithTx runs fn inside a database transaction. A new transaction is begun,
// committed on success, and rolled back on error. The callback receives the
// transaction as a Querier so callers never need to import pgx.Tx directly.
func WithTx(ctx context.Context, pool shareddomain.TxBeginner, fn func(q shareddomain.Querier) error) error {
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
