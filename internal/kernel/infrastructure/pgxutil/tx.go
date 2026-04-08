package pgxutil

import (
	"context"

	apperrors "gct/internal/kernel/infrastructure/errorx"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// -------------------------------------------------------------------------
// Context-carried transaction
// -------------------------------------------------------------------------

// txKey is the context key for an injected pgx.Tx.
type txKey struct{}

// InjectTx stores a transaction in the context so that downstream code
// (repositories, outbox writer) can participate in the same transaction
// without changing method signatures.
func InjectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// ExtractTx retrieves a transaction previously stored by InjectTx.
func ExtractTx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}

// -------------------------------------------------------------------------
// Querier — common subset of pgxpool.Pool and pgx.Tx
// -------------------------------------------------------------------------

// Querier is satisfied by both *pgxpool.Pool and pgx.Tx, covering all
// single-statement operations that repositories need.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// QuerierFromContext returns the transaction from the context when present,
// falling back to the pool otherwise. Use this in repository write methods
// so they automatically participate in an outer transaction managed by
// EventCommitter.
func QuerierFromContext(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx, ok := ExtractTx(ctx); ok {
		return tx
	}
	return pool
}

// -------------------------------------------------------------------------
// Transaction helpers
// -------------------------------------------------------------------------

// TxBeginner is the interface for starting a transaction.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// WithTx runs fn inside a database transaction.
// If the context already carries a transaction (via InjectTx), fn is called
// with that transaction and no new transaction is started — the outer caller
// owns commit/rollback. Otherwise a new transaction is begun, committed on
// success, and rolled back on error.
func WithTx(ctx context.Context, pool TxBeginner, fn func(tx pgx.Tx) error) error {
	if tx, ok := ExtractTx(ctx); ok {
		return fn(tx)
	}

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
