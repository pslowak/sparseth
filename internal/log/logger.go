package log

import "log/slog"

type Logger interface {
	// With returns a logger that includes the
	// given attributes in each output operation
	With(ctx ...any) Logger

	// Debug logs a message at the debug level
	// with context key/value pairs.
	Debug(msg string, ctx ...any)

	// Info logs a message at the info level
	// with context key/value pairs.
	Info(msg string, ctx ...any)

	// Warn logs a message at the warn level
	// with context key/value pairs.
	Warn(msg string, ctx ...any)

	// Error logs a message at the error level
	// with context key/value pairs.
	Error(msg string, ctx ...any)
}

type logger struct {
	inner *slog.Logger
}

// New returns a new logger with the
// specified handler set.
func New(handler slog.Handler) Logger {
	return &logger{
		inner: slog.New(handler),
	}
}

// With returns a Logger that includes the
// given attributes in each output operation.
func (l *logger) With(ctx ...any) Logger {
	return &logger{l.inner.With(ctx...)}
}

// Debug logs the given message at Debug level.
func (l *logger) Debug(msg string, ctx ...any) {
	l.inner.Debug(msg, ctx...)
}

// Info logs the given message at Info level.
func (l *logger) Info(msg string, ctx ...any) {
	l.inner.Info(msg, ctx...)
}

// Warn logs the given message at Warn level.
func (l *logger) Warn(msg string, ctx ...any) {
	l.inner.Warn(msg, ctx...)
}

// Error logs the given message at Error level.
func (l *logger) Error(msg string, ctx ...any) {
	l.inner.Error(msg, ctx...)
}
