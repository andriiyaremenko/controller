package controller

import (
	"context"
	"log/slog"
	"sync/atomic"
)

type Logger interface {
	ErrorContext(ctx context.Context, msg string, args ...any)
}

func SetLogger(l Logger) {
	if l != nil {
		loggerPtr.Store(&l)
	}
}

func init() {
	loggerPtr.Store(func() *Logger { var l Logger = slog.Default(); return &l }())
}

var loggerPtr atomic.Pointer[Logger]

func logger() Logger {
	return *loggerPtr.Load()
}
