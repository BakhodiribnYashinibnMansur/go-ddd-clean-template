package postgres

import (
	"testing"
	"time"

	"gct/internal/shared/infrastructure/logger"
)

func TestExtractOperation(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want string
	}{
		{"SELECT", "SELECT * FROM users", "SELECT"},
		{"INSERT", "INSERT INTO users", "INSERT"},
		{"UPDATE", "UPDATE users SET", "UPDATE"},
		{"DELETE", "DELETE FROM users", "DELETE"},
		{"WITH", "WITH cte AS (...)", "WITH"},
		{"BEGIN", "BEGIN", "TX"},
		{"COMMIT", "COMMIT", "TX"},
		{"ROLLBACK", "ROLLBACK", "TX"},
		{"OTHER", "EXPLAIN ANALYZE", "OTHER"},
		{"empty", "", "unknown"},
		{"leading whitespace", "  SELECT * FROM users", "SELECT"},
		{"lowercase", "select * from users", "SELECT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractOperation(tt.sql)
			if got != tt.want {
				t.Errorf("extractOperation(%q) = %q, want %q", tt.sql, got, tt.want)
			}
		})
	}
}

func TestNewMetricsTracer(t *testing.T) {
	t.Run("non-nil result", func(t *testing.T) {
		tracer := NewMetricsTracer(nil, logger.Noop(), 100*time.Millisecond)
		if tracer == nil {
			t.Fatal("expected non-nil MetricsTracer")
		}
	})

	t.Run("nil inner tracer accepted", func(t *testing.T) {
		tracer := NewMetricsTracer(nil, logger.Noop(), 100*time.Millisecond)
		if tracer == nil {
			t.Fatal("expected non-nil MetricsTracer with nil inner")
		}
		if tracer.inner != nil {
			t.Error("expected inner to be nil")
		}
	})
}
