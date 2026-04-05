package logger

import (
	"context"
	"testing"
)

func TestSafeContext_Nil(t *testing.T) {
	ctx := safeContext(nil)
	if ctx == nil {
		t.Fatal("expected non-nil context for nil input")
	}
}

func TestSafeContext_ActiveContext(t *testing.T) {
	original := context.Background()
	ctx := safeContext(original)
	if ctx != original {
		t.Error("expected same context back for active context")
	}
}

func TestSafeContext_CancelledContext(t *testing.T) {
	original, cancel := context.WithCancel(context.Background())
	cancel()

	ctx := safeContext(original)
	if ctx == nil {
		t.Fatal("expected non-nil context for cancelled input")
	}

	// The returned context should not be done
	select {
	case <-ctx.Done():
		t.Error("expected returned context to not be cancelled")
	default:
		// ok
	}
}

func TestSafeContext_WithValues(t *testing.T) {
	type ctxKey string
	key := ctxKey("testkey")

	original := context.WithValue(context.Background(), key, "testval")
	cancelCtx, cancel := context.WithCancel(original)
	cancel()

	ctx := safeContext(cancelCtx)
	// WithoutCancel should preserve values
	val := ctx.Value(key)
	if val != "testval" {
		t.Errorf("expected preserved value 'testval', got %v", val)
	}
}
