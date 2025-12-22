package logging

import (
	"io"
	"log/slog"
	"os"
	"time"
)

type Option interface {
	apply(opt *option)
}

type optionFn func(opt *option)

func (fn optionFn) apply(opt *option) {
	fn(opt)
}

type ReplaceAttr func(groups []string, a slog.Attr) slog.Attr

type option struct {
	writer      io.Writer
	level       slog.Leveler
	handler     slog.Handler
	replaceAttr ReplaceAttr
}

func defaultReplaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "time" {
		return slog.String(a.Key, time.Now().Format(time.RFC3339))
	}
	return a
}

var (
	defaultOption = &option{
		writer: os.Stdout,
		level:  slog.LevelInfo,
		handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:       slog.LevelInfo,
			ReplaceAttr: defaultReplaceAttr,
		}),
		replaceAttr: defaultReplaceAttr,
	}
)

func defaultOptions(opts ...Option) *option {
	o := *defaultOption
	for _, opt := range opts {
		opt.apply(&o)
	}
	return &o
}

func WithWriter(writer io.Writer) Option {
	return optionFn(func(opt *option) {
		opt.writer = writer
	})
}

func WithLevel(level slog.Leveler) Option {
	return optionFn(func(opt *option) {
		opt.level = level
	})
}

func WithHandler(handler slog.Handler) Option {
	return optionFn(func(opt *option) {
		opt.handler = handler
	})
}

func WithReplaceAttr(attrFn ReplaceAttr) Option {
	return optionFn(func(opt *option) {
		opt.replaceAttr = attrFn
	})
}
