package logger

import (
	"context"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const callerSeparator = " → "

type logger struct {
	zap    *zap.SugaredLogger
	ctxZap *zap.SugaredLogger
}

var globalLog = New(LevelInfo)

// customColorLevelEncoder encodes log levels with custom colors and backgrounds:
// ERROR: bright red background with white bold text
// WARN: bright yellow background with black bold text
// INFO: bright blue background with white bold text
// DEBUG: bright green background with white bold text
func customColorLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var coloredLevel string
	switch l {
	case zapcore.DebugLevel:
		// DEBUG: Matrix-style dim green
		coloredLevel = BgGray + ColorBlack + Bold + "[*] " + LevelDebugDisplay + ColorReset
	case zapcore.InfoLevel:
		// INFO: Classic hacker green
		coloredLevel = BgGreen + ColorBlack + Bold + "[+] " + LevelInfoDisplay + ColorReset
	case zapcore.WarnLevel:
		// WARN: Amber alert
		coloredLevel = BgYellow + ColorBlack + Bold + "[!] " + LevelWarnDisplay + ColorReset
	case zapcore.ErrorLevel:
		// ERROR: Critical failure
		coloredLevel = BgRed + ColorBrightWhite + Bold + " [X] " + LevelErrorDisplay + " " + ColorReset
	case zapcore.PanicLevel:
		coloredLevel = BgBrightRed + ColorBlack + Bold + " [!] CRITICAL_PANIC " + ColorReset
	case zapcore.FatalLevel:
		coloredLevel = BgRed + ColorBlack + Bold + " [💀] TERMINAL_FATAL " + ColorReset
	default:
		coloredLevel = l.CapitalString()
	}
	enc.AppendString(coloredLevel)
}

// customTimeEncoder wraps time encoding with terminal-green dim style
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(ColorGreen + Dim + "[" + t.Format(TimeFormat) + "]" + ColorReset)
}

// customCallerEncoder encodes caller with file:line and short function name
func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	fn := caller.Function
	if idx := strings.LastIndex(fn, "/"); idx >= 0 {
		fn = fn[idx+1:]
	}
	if idx := strings.Index(fn, "."); idx >= 0 {
		fn = fn[idx+1:]
	}
	fn = strings.NewReplacer("(*", "", ")", "").Replace(fn)
	enc.AppendString(ColorBrightCyan + Dim + Italic + "# " + caller.TrimmedPath() + callerSeparator + fn + ColorReset)
}

// customDurationEncoder wraps duration encoding with blue color
func customDurationEncoder(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(ColorCyan + d.String() + ColorReset)
}

const (
	FormatConsole = "console"
	FormatJSON    = "json"
)

// New creates a new logger instance with console format (default).
func New(level string) Log {
	return NewWithFormat(level, FormatConsole)
}

// NewWithFormat creates a new logger instance with the specified format.
// format: "console" for colored human-readable output, "json" for structured JSON output.
func NewWithFormat(level, format string) Log {
	logLevel := parseLevel(level)

	var encoder zapcore.Encoder
	if strings.ToLower(format) == FormatJSON {
		encoder = zapcore.NewJSONEncoder(jsonEncoderConfig())
	} else {
		encoder = zapcore.NewConsoleEncoder(consoleEncoderConfig())
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.Lock(os.Stdout),
		logLevel,
	)

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &logger{
		zap:    zapLogger.Sugar(),
		ctxZap: zapLogger.WithOptions(zap.AddCallerSkip(1)).Sugar(),
	}
}

func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func consoleEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:          KeyTimestamp,
		LevelKey:         KeyLevel,
		NameKey:          KeyLogger,
		CallerKey:        KeyCaller,
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       KeyMessage,
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      customColorLevelEncoder,
		EncodeTime:       customTimeEncoder,
		EncodeDuration:   customDurationEncoder,
		EncodeCaller:     customCallerEncoder,
		ConsoleSeparator: ConsoleSeparator,
		EncodeName: func(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(ColorGray + loggerName + ColorReset)
		},
	}
}

func jsonEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        KeyTimestamp,
		LevelKey:       KeyLevel,
		NameKey:        KeyLogger,
		CallerKey:      KeyCaller,
		FunctionKey:    "function",
		MessageKey:     KeyMessage,
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

func (l *logger) withCtx(ctx context.Context) map[string]any {
	fields := extractFields(ctx)
	if len(fields) == 0 {
		return nil
	}
	return fields
}

// WithPersistCore adds an additional zap core (e.g., RedisSink) to the logger
// so that log entries are tee'd to both console/JSON output and the persist sink.
func WithPersistCore(base Log, extra zapcore.Core) Log {
	l, ok := base.(*logger)
	if !ok {
		return base
	}
	// Rebuild underlying zap logger with a tee core
	baseZap := l.zap.Desugar()
	tee := zapcore.NewTee(baseZap.Core(), extra)
	newZap := zap.New(tee, zap.AddCaller(), zap.AddCallerSkip(1))

	return &logger{
		zap:    newZap.Sugar(),
		ctxZap: newZap.WithOptions(zap.AddCallerSkip(1)).Sugar(),
	}
}
