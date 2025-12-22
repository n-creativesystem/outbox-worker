package logging

import (
	"log/slog"
	"time"
)

func NewTextHandler(opts ...Option) slog.Handler {
	o := defaultOptions(opts...)
	return slog.NewTextHandler(o.writer, &slog.HandlerOptions{
		Level: o.level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "time" { // nolint: goconst
				return slog.String(a.Key, time.Now().Format(time.RFC3339))
			}
			return a
		},
	})
}
