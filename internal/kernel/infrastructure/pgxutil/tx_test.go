package pgxutil

import (
	"context"
	"errors"
	"testing"

	shareddomain "gct/internal/kernel/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// mockTx implements pgx.Tx for testing
type mockTx struct {
	commitErr   error
	rollbackErr error
	committed   bool
	rolledBack  bool
}

func (m *mockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	return &mockTx{}, nil
}
func (m *mockTx) Commit(ctx context.Context) error {
	m.committed = true
	return m.commitErr
}
func (m *mockTx) Rollback(ctx context.Context) error {
	m.rolledBack = true
	return m.rollbackErr
}
func (m *mockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}
func (m *mockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return nil
}
func (m *mockTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}
func (m *mockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}
func (m *mockTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}
func (m *mockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}
func (m *mockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return nil
}
func (m *mockTx) Conn() *pgx.Conn {
	return nil
}

// mockPool implements TxBeginner for testing
type mockPool struct {
	tx       *mockTx
	beginErr error
}

func (m *mockPool) Begin(ctx context.Context) (pgx.Tx, error) {
	if m.beginErr != nil {
		return nil, m.beginErr
	}
	return m.tx, nil
}

func TestWithTx_Success(t *testing.T) {
	tx := &mockTx{}
	pool := &mockPool{tx: tx}

	called := false
	err := WithTx(context.Background(), pool, func(q shareddomain.Querier) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("WithTx returned error: %v", err)
	}
	if !called {
		t.Error("expected fn to be called")
	}
	if !tx.committed {
		t.Error("expected transaction to be committed")
	}
}

func TestWithTx_FnError(t *testing.T) {
	tx := &mockTx{}
	pool := &mockPool{tx: tx}
	fnErr := errors.New("fn error")

	err := WithTx(context.Background(), pool, func(q shareddomain.Querier) error {
		return fnErr
	})
	if err == nil {
		t.Fatal("expected error from WithTx")
	}
	if !errors.Is(err, fnErr) {
		t.Errorf("expected fn error, got %v", err)
	}
	if tx.committed {
		t.Error("expected transaction NOT to be committed on fn error")
	}
}

func TestWithTx_BeginError(t *testing.T) {
	pool := &mockPool{beginErr: errors.New("begin failed")}

	err := WithTx(context.Background(), pool, func(q shareddomain.Querier) error {
		t.Fatal("fn should not be called when Begin fails")
		return nil
	})
	if err == nil {
		t.Fatal("expected error from WithTx when Begin fails")
	}
}

func TestWithTx_CommitError(t *testing.T) {
	tx := &mockTx{commitErr: errors.New("commit failed")}
	pool := &mockPool{tx: tx}

	err := WithTx(context.Background(), pool, func(q shareddomain.Querier) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error from WithTx when Commit fails")
	}
}
