package logs

import (
	"context"
	"errors"
	"log/slog"
)

var _ slog.Handler = (*fanoutHandler)(nil)

type fanoutHandler struct {
	handlers []slog.Handler
}

// SlogFanout returns a new slog.Handler that fans out log records to all the given handlers.
func SlogFanout(handlers ...slog.Handler) slog.Handler {
	return &fanoutHandler{
		handlers: handlers,
	}
}

func (h *fanoutHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, l) {
			return true
		}
	}

	return false
}

func (h *fanoutHandler) Handle(ctx context.Context, r slog.Record) (err error) {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, r.Level) {
			err = errors.Join(err,
				h.handlers[i].Handle(ctx, r.Clone()),
			)
		}
	}
	return err
}

func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlers[i] = h.handlers[i].WithAttrs(attrs)
	}
	return &fanoutHandler{
		handlers: handlers,
	}
}

func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlers[i] = h.handlers[i].WithGroup(name)
	}
	return &fanoutHandler{
		handlers: handlers,
	}
}
