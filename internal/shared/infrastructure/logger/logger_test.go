package logger

import (
	"testing"

	"go.uber.org/zap/zapcore"
)

// noopCore is a minimal zapcore.Core implementation for testing.
type noopCore struct{}

func (n *noopCore) Enabled(zapcore.Level) bool                                  { return true }
func (n *noopCore) With([]zapcore.Field) zapcore.Core                           { return n }
func (n *noopCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry { return ce }
func (n *noopCore) Write(zapcore.Entry, []zapcore.Field) error                  { return nil }
func (n *noopCore) Sync() error                                                 { return nil }

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  zapcore.Level
	}{
		{"debug", "debug", zapcore.DebugLevel},
		{"info", "info", zapcore.InfoLevel},
		{"warn", "warn", zapcore.WarnLevel},
		{"error", "error", zapcore.ErrorLevel},
		{"unknown defaults to info", "unknown", zapcore.InfoLevel},
		{"empty defaults to info", "", zapcore.InfoLevel},
		{"uppercase matches via ToLower", "DEBUG", zapcore.DebugLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLevel(tt.input)
			if got != tt.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	l := New("info")
	if l == nil {
		t.Fatal("New returned nil")
	}

	// Verify it satisfies the Log interface.
	var _ Log = l
}

func TestNewWithFormat_Console(t *testing.T) {
	l := NewWithFormat("info", FormatConsole)
	if l == nil {
		t.Fatal("NewWithFormat with console format returned nil")
	}
}

func TestNewWithFormat_JSON(t *testing.T) {
	l := NewWithFormat("info", FormatJSON)
	if l == nil {
		t.Fatal("NewWithFormat with JSON format returned nil")
	}
}

func TestNewWithFormat_UnknownFormat(t *testing.T) {
	// Should default to console and not panic.
	l := NewWithFormat("info", "xml")
	if l == nil {
		t.Fatal("NewWithFormat with unknown format returned nil")
	}
}

func TestWithPersistCore(t *testing.T) {
	base := New("info")
	result := WithPersistCore(base, &noopCore{})

	if result == nil {
		t.Fatal("WithPersistCore returned nil")
	}

	if _, ok := result.(*logger); !ok {
		t.Fatalf("WithPersistCore returned %T, want *logger", result)
	}
}

func TestWithPersistCore_NonLoggerType(t *testing.T) {
	base := Noop()
	result := WithPersistCore(base, &noopCore{})

	if result != base {
		t.Fatal("WithPersistCore should return the same object for non-*logger input")
	}
}

// TestLoggerImplementsInterface is a compile-time check.
var _ Log = (*logger)(nil)
