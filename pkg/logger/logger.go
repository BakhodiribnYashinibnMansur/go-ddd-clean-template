package logger

import (
	"context"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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

// customCallerEncoder encodes caller with hacker-cyan and italics
func customCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(ColorBrightCyan + Dim + Italic + "# " + caller.TrimmedPath() + ColorReset)
}

// customDurationEncoder wraps duration encoding with blue color
func customDurationEncoder(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(ColorCyan + d.String() + ColorReset)
}

// New creates a new logger instance.
func New(level string) Log {
	var logLevel zapcore.Level

	switch strings.ToLower(level) {
	case LevelDebug:
		logLevel = zapcore.DebugLevel
	case LevelInfo:
		logLevel = zapcore.InfoLevel
	case LevelWarn:
		logLevel = zapcore.WarnLevel
	case LevelError:
		logLevel = zapcore.ErrorLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:          KeyTimestamp,
		LevelKey:         KeyLevel,
		NameKey:          KeyLogger,
		CallerKey:        KeyCaller,
		MessageKey:       KeyMessage,
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      customColorLevelEncoder,
		EncodeTime:       customTimeEncoder,
		EncodeDuration:   customDurationEncoder,
		EncodeCaller:     customCallerEncoder,
		ConsoleSeparator: ConsoleSeparator,
		// Colorize field keys
		EncodeName: func(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(ColorGray + loggerName + ColorReset)
		},
	}

	encoder := zapcore.NewConsoleEncoder(encoderConfig)

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

func (l *logger) withCtx(ctx context.Context) map[string]any {
	fields := extractFields(ctx)
	if len(fields) == 0 {
		return nil
	}
	return fields
}
