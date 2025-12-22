package logging

import (
	"log/slog"
	"strings"

	"github.com/getsentry/sentry-go"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	slogsentry "github.com/samber/slog-sentry/v2"
)

type SentryConfig struct {
	Level          string           `yaml:"level" default:"warn"`
	DSN            string           `yaml:"dsn"`
	SampleRate     float64          `yaml:"sampleRate" default:"1.0"`
	IgnoreErrors   []string         `yaml:"ignoreErrors"`
	SendDefaultPII bool             `yaml:"sendDefaultPII"`
	Transport      sentry.Transport `yaml:"-"`
}

func (c SentryConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DSN, validation.Required),
	)
}

func (c *SentryConfig) SentryOptions(environment string) sentry.ClientOptions {
	return sentry.ClientOptions{
		Dsn:            c.DSN,
		IgnoreErrors:   c.IgnoreErrors,
		Environment:    environment,
		SampleRate:     c.SampleRate,
		SendDefaultPII: c.SendDefaultPII,
		Transport:      c.Transport,
	}
}

func (c *SentryConfig) getLevel() slog.Level {
	switch strings.ToLower(c.Level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "err", "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func NewSentryHandler(conf *SentryConfig, environment string) Handle {
	if err := sentry.Init(conf.SentryOptions(environment)); err != nil {
		slog.Error("Sentry initialize")
	}

	option := slogsentry.Option{
		Level: conf.getLevel(),
	}
	return NewErrorTracking(NewAsyncHandler(option.NewSentryHandler()))
}
