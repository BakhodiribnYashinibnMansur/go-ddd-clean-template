package logger

// Value methods (Key-Value pairs)

func (l *logger) Debugw(msg string, keysAndValues ...any) {
	l.zap.Debugw(msg, keysAndValues...)
}

func (l *logger) Infow(msg string, keysAndValues ...any) {
	l.zap.Infow(msg, keysAndValues...)
}

func (l *logger) Warnw(msg string, keysAndValues ...any) {
	l.zap.Warnw(msg, keysAndValues...)
}

func (l *logger) Errorw(msg string, keysAndValues ...any) {
	l.zap.Errorw(msg, keysAndValues...)
}

func (l *logger) Fatalw(msg string, keysAndValues ...any) {
	l.zap.Fatalw(msg, keysAndValues...)
}
