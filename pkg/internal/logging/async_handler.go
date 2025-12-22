package logging

import (
	"context"
	"io"
	"log/slog"
	"sync"
)

type AsyncHandler struct {
	slog.Handler
	sync sync.WaitGroup
}

var (
	_ Handle = (*AsyncHandler)(nil)
)

func NewAsyncHandler(h slog.Handler) slog.Handler {
	return &AsyncHandler{Handler: h}
}

func (h *AsyncHandler) Handle(ctx context.Context, r slog.Record) error {
	h.sync.Add(1)
	go func() {
		defer h.sync.Done()
		_ = h.Handler.Handle(ctx, r) // nolint: errcheck
	}()
	return nil
}

func (h *AsyncHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewAsyncHandler(h.Handler.WithAttrs(attrs))
}

func (h *AsyncHandler) WithGroup(name string) slog.Handler {
	return NewAsyncHandler(h.Handler.WithGroup(name))
}

func (h *AsyncHandler) Close() error {
	h.sync.Wait()
	if v, ok := h.Handler.(io.Closer); ok {
		return v.Close()
	}
	return nil
}
