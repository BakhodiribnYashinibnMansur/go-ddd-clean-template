package logger

const (
	// Log levels
	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"
	LevelFatal = "fatal"

	// Encoder keys
	KeyTimestamp = "timestamp"
	KeyLevel     = "level"
	KeyLogger    = "logger"
	KeyCaller    = "caller"
	KeyMessage   = "msg"
)
