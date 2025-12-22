package logging

import (
	"log/slog"
	"time"
)

func NewJSONHandler(opts ...Option) slog.Handler {
	o := defaultOptions(opts...)
	return slog.NewJSONHandler(o.writer, &slog.HandlerOptions{
		Level: o.level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "time" {
				return slog.String(a.Key, time.Now().Format(time.RFC3339))
			}
			return a
		},
	})
}
