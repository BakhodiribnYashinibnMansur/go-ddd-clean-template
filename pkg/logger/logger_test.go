package logger

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newBufferedLogger. Helper to replace underlying zap writer with a buffer and capture logs.
func newBufferedLogger(level string) (*Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	var logLevel zapcore.Level
	if err := logLevel.UnmarshalText([]byte(level)); err != nil {
		logLevel = zapcore.InfoLevel
	}

	core := zapcore.NewCore(encoder, zapcore.AddSync(buf), logLevel)
	l := &Logger{
		Entity: zap.New(core).Sugar(),
	}

	return l, buf
}

func TestNew(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		name string
	}{
		{"debug", "debug level"},
		{"info", "info level"},
		{"warn", "warn level"},
		{"error", "error level"},
		{"unknown", "default level"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := New(tc.in)
			if l == nil {
				t.Fatalf("New(%q) returned nil logger or entity", tc.in)
			}
		})
	}
}

func TestInfoAndWarn_LogMessageWithAndWithoutArgs(t *testing.T) {
	t.Parallel()

	l, buf := newBufferedLogger("info")

	l.Info("hello")
	l.Infof("hello %s", "world")
	l.Warnf("warn %d", 7)

	out := buf.String()

	// Expect level fields and messages present
	if !strings.Contains(out, "\"level\":\"info\"") || !strings.Contains(out, "\"message\":\"hello\"") {
		t.Fatalf("info without args not found in output: %s", out)
	}

	if !strings.Contains(out, "hello world") {
		t.Fatalf("formatted info not found in output: %s", out)
	}

	if !strings.Contains(out, "\"level\":\"warn\"") || !strings.Contains(out, "warn 7") {
		t.Fatalf("warn log not found in output: %s", out)
	}
}

func TestDebug_RespectsLevel(t *testing.T) {
	t.Parallel()

	// when level is info, debug should not emit
	l, buf := newBufferedLogger("info")
	l.Debug("dbg %d", 1)

	if got := buf.String(); got != "" {
		if strings.Contains(got, "\"level\":\"debug\"") {
			t.Fatalf("debug should be suppressed at info level, got: %s", got)
		}
	}

	// when level is debug, debug should emit
	l, buf = newBufferedLogger("debug")
	l.Debugf("dbg %d", 2)

	out := buf.String()

	if !strings.Contains(out, "\"level\":\"debug\"") || !strings.Contains(out, "dbg 2") {
		t.Fatalf("debug should appear at debug level, got: %s", out)
	}
}

func TestError_LogsError(t *testing.T) {
	t.Parallel()

	l, buf := newBufferedLogger("info")
	l.Error("boom")

	out := buf.String()

	if strings.Count(out, "\"level\":\"error\"") != 1 {
		t.Fatalf("expected 1 error log, got: %s", out)
	}
}

func TestStructuredLogging(t *testing.T) {
	t.Parallel()

	l, buf := newBufferedLogger("info")

	l.Infow("structured message", "key1", "value1", "key2", 42)

	out := buf.String()

	if !strings.Contains(out, "\"message\":\"structured message\"") {
		t.Fatalf("message not found: %s", out)
	}
	if !strings.Contains(out, "\"key1\":\"value1\"") {
		t.Fatalf("key1 not found: %s", out)
	}
	if !strings.Contains(out, "\"key2\":42") {
		t.Fatalf("key2 not found: %s", out)
	}
}

func TestWithFields(t *testing.T) {
	t.Parallel()

	l, buf := newBufferedLogger("info")

	l2 := l.WithFields(map[string]any{
		"component": "test",
		"version":   "1.0",
	})

	l2.Info("with fields")

	out := buf.String()

	if !strings.Contains(out, "\"component\":\"test\"") || !strings.Contains(out, "\"version\":\"1.0\"") {
		t.Fatalf("fields from WithFields not found: %s", out)
	}
}

func TestFatal_ExitsAndLogs(t *testing.T) {
	// Fatal test cannot be run in parallel with others because it might affect global state if any,
	// though here it's in a sub-process.

	if os.Getenv("LOGGER_FATAL_SUBPROC") == "1" {
		// child process: run Fatal and exit
		l := New("debug")
		// Fatal in zap will stop the process.
		// Since New("debug") writes to stdout, we can capture it.
		l.Fatal("fatal now")

		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestFatal_ExitsAndLogs")
	cmd.Env = append(os.Environ(), "LOGGER_FATAL_SUBPROC=1")

	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-nil error due to os.Exit in Fatal, got nil; output: %s", string(out))
	}

	// In zap, Fatal logs at fatal level and then calls os.Exit(1)
	if !strings.Contains(string(out), "fatal now") {
		t.Fatalf("expected fatal message in output, got: %s", string(out))
	}
}
