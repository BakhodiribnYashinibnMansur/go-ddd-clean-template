package logger

// Simple methods

func (l *logger) Debug(args ...any) {
	l.zap.Debug(args...)
}

func (l *logger) Debugf(template string, args ...any) {
	l.zap.Debugf(template, args...)
}

func (l *logger) Info(args ...any) {
	l.zap.Info(args...)
}

func (l *logger) Infof(template string, args ...any) {
	l.zap.Infof(template, args...)
}

func (l *logger) Warn(args ...any) {
	l.zap.Warn(args...)
}

func (l *logger) Warnf(template string, args ...any) {
	l.zap.Warnf(template, args...)
}

func (l *logger) Error(args ...any) {
	l.zap.Error(args...)
}

func (l *logger) Errorf(template string, args ...any) {
	l.zap.Errorf(template, args...)
}

func (l *logger) Fatal(args ...any) {
	l.zap.Fatal(args...)
}

func (l *logger) Fatalf(template string, args ...any) {
	l.zap.Fatalf(template, args...)
}
