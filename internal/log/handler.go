package log

import (
	"context"
	"fmt"
	"log/slog"
)

type TerminalHandler struct {
	lvl       slog.Level
	attrs     []slog.Attr
	component string
}

func (h *TerminalHandler) Enabled(_ context.Context, lvl slog.Level) bool {
	return lvl >= h.lvl
}

func (h *TerminalHandler) Handle(_ context.Context, r slog.Record) error {
	msg := r.Message
	lvl := r.Level.String()

	color := ""
	switch r.Level {
	case slog.LevelInfo:
		color = "\x1b[32m" // green
	case slog.LevelWarn:
		color = "\x1b[33m" // yellow
	case slog.LevelError:
		color = "\x1b[31m" // red
	}

	time := ""
	if !r.Time.IsZero() {
		time = fmt.Sprintf("[%s]", r.Time.Format("Jan 02|15:04:05.000"))
	}

	attrs := ""
	r.Attrs(func(a slog.Attr) bool {
		attrs += fmt.Sprintf("[%s=%s] ", a.Key, a.Value)
		return true
	})

	_, err := fmt.Println(color, time, lvl, h.component, msg, attrs)

	return err
}

func (h *TerminalHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	component := "[]"
	for _, attr := range attrs {
		if attr.Key == "component" {
			component = fmt.Sprintf("[%s]", attr.Value)
		}
	}

	return &TerminalHandler{
		lvl:       h.lvl,
		attrs:     append(h.attrs, attrs...),
		component: component,
	}
}

func (h *TerminalHandler) WithGroup(_ string) slog.Handler {
	panic("not implemented")
}

// NewTerminalHandler creates a new terminal
// log handler that prints colorful messages
// to stdout.
func NewTerminalHandler() *TerminalHandler {
	return &TerminalHandler{
		lvl:       slog.LevelDebug,
		attrs:     []slog.Attr{},
		component: "[]",
	}
}
