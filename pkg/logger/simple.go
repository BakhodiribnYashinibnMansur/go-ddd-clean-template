package logger

// Simple methods

func (l *logger) Debug(args ...interface{}) {
	l.zap.Debug(args...)
}

func (l *logger) Debugf(template string, args ...interface{}) {
	l.zap.Debugf(template, args...)
}

func (l *logger) Info(args ...interface{}) {
	l.zap.Info(args...)
}

func (l *logger) Infof(template string, args ...interface{}) {
	l.zap.Infof(template, args...)
}

func (l *logger) Warn(args ...interface{}) {
	l.zap.Warn(args...)
}

func (l *logger) Warnf(template string, args ...interface{}) {
	l.zap.Warnf(template, args...)
}

func (l *logger) Error(args ...interface{}) {
	l.zap.Error(args...)
}

func (l *logger) Errorf(template string, args ...interface{}) {
	l.zap.Errorf(template, args...)
}

func (l *logger) Fatal(args ...interface{}) {
	l.zap.Fatal(args...)
}

func (l *logger) Fatalf(template string, args ...interface{}) {
	l.zap.Fatalf(template, args...)
}
