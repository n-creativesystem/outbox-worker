package logging

import (
	"context"
	"io"
	"log/slog"
	"slices"

	"go.uber.org/multierr"
)

func NewHandler(handlers ...slog.Handler) Handle {
	return &handler{
		handlers: handlers,
	}
}

type handler struct {
	handlers []slog.Handler
}

var (
	_ slog.Handler = (*handler)(nil)
)

func (h *handler) handler(fn func(h slog.Handler)) {
	for _, handler := range h.handlers {
		if handler != nil {
			fn(handler)
		}
	}
}

func (h *handler) Enabled(ctx context.Context, level slog.Level) bool {
	flags := make([]bool, 0, len(h.handlers))
	h.handler(func(h slog.Handler) {
		flags = append(flags, h.Enabled(ctx, level))
	})
	return slices.Contains(flags, true)
}

func (h *handler) Handle(ctx context.Context, record slog.Record) error {
	var err error
	h.handler(func(h slog.Handler) {
		if h.Enabled(ctx, record.Level) {
			if e := h.Handle(ctx, record); e != nil {
				err = multierr.Append(err, e)
			}
		}
	})
	return err
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	h.handler(func(h slog.Handler) {
		handlers = append(handlers, h.WithAttrs(attrs))
	})
	return NewHandler(handlers...)
}

func (h *handler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.handlers))
	h.handler(func(h slog.Handler) {
		handlers = append(handlers, h.WithGroup(name))
	})
	return NewHandler(handlers...)
}

func (h *handler) Close() error {
	h.handler(func(h slog.Handler) {
		if v, ok := h.(io.Closer); ok {
			_ = v.Close() // nolint: errcheck
		}
	})
	return nil
}
