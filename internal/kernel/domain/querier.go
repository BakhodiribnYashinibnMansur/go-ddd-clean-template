package domain

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier is the minimal database access interface satisfied by both
// *pgxpool.Pool and pgx.Tx. Write-side repository methods accept this
// so callers control transaction boundaries explicitly.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// DB combines transaction-starting ability with query execution.
// *pgxpool.Pool satisfies this interface.
type DB interface {
	TxBeginner
	Querier
}

// TxBeginner is the interface for starting a transaction.
type TxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Compile-time check that *pgxpool.Pool satisfies DB.
var _ DB = (*pgxpool.Pool)(nil)
