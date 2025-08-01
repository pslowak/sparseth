package log

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
