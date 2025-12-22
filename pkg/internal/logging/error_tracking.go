package logging

import (
	"context"
	"io"
	"log/slog"
)

var (
	IgnoreTracing = slog.Attr{
		Key:   "Ignore",
		Value: slog.BoolValue(true),
	}
)

type ErrorTracking struct {
	slog.Handler
	ignore bool
}

func NewErrorTracking(handler slog.Handler) Handle {
	return newErrorTracking(handler, false)
}

func newErrorTracking(handler slog.Handler, ignore bool) Handle {
	return &ErrorTracking{
		Handler: handler,
		ignore:  ignore,
	}
}

func (h *ErrorTracking) Handle(ctx context.Context, record slog.Record) error {
	if h.ignore {
		return nil
	}
	_ = h.Handler.Handle(ctx, record) // nolint: errcheck
	return nil
}

func (h *ErrorTracking) WithAttrs(attrs []slog.Attr) slog.Handler {
	if !h.ignore {
		for _, attr := range attrs {
			if attr.Equal(IgnoreTracing) {
				h.ignore = true
			}
		}
	}
	return newErrorTracking(h.Handler.WithAttrs(attrs), h.ignore)
}

func (h *ErrorTracking) WithGroup(name string) slog.Handler {
	return newErrorTracking(h.Handler.WithGroup(name), h.ignore)
}

func (h *ErrorTracking) Close() error {
	if v, ok := h.Handler.(io.Closer); ok {
		return v.Close()
	}
	return nil
}
