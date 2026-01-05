package logger

import (
	"context"
	"os"
	"strings"
	"sync"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey string

const (
	// KeyRequestID is the context key for Request ID.
	KeyRequestID ctxKey = "request_id"
)

var (
	instance *Logger
	once     sync.Once
)

// Interface -.
type Log interface {
	Debug(args ...any)
	Info(args ...any)
	Warn(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Panic(args ...any)

	Debugf(template string, args ...any)
	Infof(template string, args ...any)
	Warnf(template string, args ...any)
	Errorf(template string, args ...any)
	Fatalf(template string, args ...any)
	Panicf(template string, args ...any)

	Debugw(msg string, keysAndValues ...any)
	Infow(msg string, keysAndValues ...any)
	Warnw(msg string, keysAndValues ...any)
	Errorw(msg string, keysAndValues ...any)
	Fatalw(msg string, keysAndValues ...any)
	Panicw(msg string, keysAndValues ...any)

	WithField(key string, value any) *Logger
	WithFields(fields map[string]any) *Logger
	WithContext(ctx context.Context) Log
	GetZap() *zap.Logger
	Sync()
}

// Logger wraps zap.SugaredLogger for structured logging.
type Logger struct {
	Entity *zap.SugaredLogger
}

// GetLogger returns the singleton logger instance.
func GetLogger() Log {
	once.Do(func() {
		instance = New(os.Getenv("LOG_LEVEL"))
	})
	return instance
}

// New creates a new Logger with specific level.
func New(level string) *Logger {
	var logLevel zapcore.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewCore(
		consoleEncoder,
		zapcore.Lock(os.Stdout),
		logLevel,
	)

	baseLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

	return &Logger{
		Entity: baseLogger.Sugar(),
	}
}

// WithContext returns a logger with context fields (Request ID, Trace ID, Span ID).
func (log *Logger) WithContext(ctx context.Context) Log {
	var args []interface{}
	if ctx == nil {
		return log
	}

	if id, ok := ctx.Value(KeyRequestID).(string); ok {
		args = append(args, zap.String("request_id", id))
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		args = append(args, zap.String("trace_id", span.SpanContext().TraceID().String()))
		args = append(args, zap.String("span_id", span.SpanContext().SpanID().String()))
	}

	if len(args) > 0 {
		return &Logger{Entity: log.Entity.With(args...)}
	}

	return log
}

// WithField creates a logger with an additional field.
func (log *Logger) WithField(key string, value any) *Logger {
	return &Logger{log.Entity.With(key, value)}
}

// WithFields creates a logger with multiple additional fields.
func (log *Logger) WithFields(fields map[string]any) *Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{Entity: log.Entity.With(args...)}
}

// Sync flushes any buffered log entries.
func (log *Logger) Sync() {
	_ = log.Entity.Sync()
}

// Info logs an info level message.
func (log *Logger) Info(args ...any) {
	log.Entity.Info(args...)
}

// Debug logs a debug level message.
func (log *Logger) Debug(args ...any) {
	log.Entity.Debug(args...)
}

// Warn logs a warning level message.
func (log *Logger) Warn(args ...any) {
	log.Entity.Warn(args...)
}

// Error logs an error level message.
func (log *Logger) Error(args ...any) {
	log.Entity.Error(args...)
}

// Fatal logs a fatal level message and exits.
func (log *Logger) Fatal(args ...any) {
	log.Entity.Fatal(args...)
}

// Panic logs a panic level message and panics.
func (log *Logger) Panic(args ...any) {
	log.Entity.Panic(args...)
}

// Infof logs a formatted info level message.
func (log *Logger) Infof(template string, args ...any) {
	log.Entity.Infof(template, args...)
}

// Debugf logs a formatted debug level message.
func (log *Logger) Debugf(template string, args ...any) {
	log.Entity.Debugf(template, args...)
}

// Warnf logs a formatted warning level message.
func (log *Logger) Warnf(template string, args ...any) {
	log.Entity.Warnf(template, args...)
}

// Errorf logs a formatted error level message.
func (log *Logger) Errorf(template string, args ...any) {
	log.Entity.Errorf(template, args...)
}

// Fatalf logs a formatted fatal level message and exits.
func (log *Logger) Fatalf(template string, args ...any) {
	log.Entity.Fatalf(template, args...)
}

// Panicf logs a formatted panic level message and panics.
func (log *Logger) Panicf(template string, args ...any) {
	log.Entity.Panicf(template, args...)
}

// Infow logs a message with structured key-value pairs at info level.
func (log *Logger) Infow(msg string, keysAndValues ...any) {
	log.Entity.Infow(msg, keysAndValues...)
}

// Debugw logs a message with structured key-value pairs at debug level.
func (log *Logger) Debugw(msg string, keysAndValues ...any) {
	log.Entity.Debugw(msg, keysAndValues...)
}

// Warnw logs a message with structured key-value pairs at warning level.
func (log *Logger) Warnw(msg string, keysAndValues ...any) {
	log.Entity.Warnw(msg, keysAndValues...)
}

// Errorw logs a message with structured key-value pairs at error level.
func (log *Logger) Errorw(msg string, keysAndValues ...any) {
	log.Entity.Errorw(msg, keysAndValues...)
}

// Fatalw logs a message with structured key-value pairs at fatal level and exits.
func (log *Logger) Fatalw(msg string, keysAndValues ...any) {
	log.Entity.Fatalw(msg, keysAndValues...)
}

// Panicw logs a message with structured key-value pairs at panic level and panics.
func (log *Logger) Panicw(msg string, keysAndValues ...any) {
	log.Entity.Panicw(msg, keysAndValues...)
}

// GetZap returns the underlying zap logger.
func (log *Logger) GetZap() *zap.Logger {
	return log.Entity.Desugar()
}
