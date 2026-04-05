package logger

import (
	"testing"
)

func TestNew_ReturnsNonNil(t *testing.T) {
	l := New(LevelInfo)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNew_AllLevels(t *testing.T) {
	levels := []string{LevelDebug, LevelInfo, LevelWarn, LevelError, "unknown"}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			l := New(level)
			if l == nil {
				t.Fatalf("expected non-nil logger for level %q", level)
			}
		})
	}
}

func TestLogger_SimpleMethodsDoNotPanic(t *testing.T) {
	l := New(LevelDebug)
	impl, ok := l.(*logger)
	if !ok {
		t.Fatal("expected *logger type")
	}

	// These should not panic. We cannot easily capture output
	// since the logger writes to stdout, but we verify no panics.
	impl.Debug("test debug")
	impl.Debugf("test debugf %s", "arg")
	impl.Info("test info")
	impl.Infof("test infof %s", "arg")
	impl.Warn("test warn")
	impl.Warnf("test warnf %s", "arg")
	impl.Error("test error")
	impl.Errorf("test errorf %s", "arg")
	// Skip Fatal/Fatalf as they call os.Exit
}

func TestGlobalLog_NotNil(t *testing.T) {
	if globalLog == nil {
		t.Fatal("expected globalLog to be initialized")
	}
}
