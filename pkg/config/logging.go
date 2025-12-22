package config

import (
	"log/slog"
	"os"
	"slices"

	"github.com/creasty/defaults"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/n-creativesystem/outbox-worker/pkg/internal/logging"
	"github.com/n-creativesystem/outbox-worker/pkg/internal/version"
)

type Logging struct {
	Level   slog.Level             `yaml:"level"`
	Handler any                    `yaml:"handler" default:"text"`
	Sentry  *logging.SentryConfig  `yaml:"sentry"`
	Rollbar *logging.RollbarConfig `yaml:"rollbar"`

	handlers []logging.LoggingHandle
	closer   func()
}

func (l *Logging) setupLog(key logging.LoggingHandle, tracking Tracking) slog.Handler {
	switch key {
	case "sentry":
		return logging.NewSentryHandler(l.Sentry, tracking.Environment)
	case "rollbar":
		l.Rollbar.Init(tracking.Environment, version.Version, tracking.ServiceName)
		l.closer = l.Rollbar.Close
		return logging.NewRollbarHandler(l.Rollbar)
	case "json":
		return logging.NewJSONHandler(
			logging.WithWriter(os.Stdout),
			logging.WithLevel(l.Level),
		)
	case "text":
		return logging.NewTextHandler(
			logging.WithWriter(os.Stdout),
			logging.WithLevel(l.Level),
		)
	}
	return nil
}

func (l *Logging) SetupLog() {
	var (
		handlers = make([]slog.Handler, 0, 10)
		tracking = DefaultConfig().Tracking
	)
	for _, handle := range l.handlers {
		h := l.setupLog(handle, tracking)
		if h != nil {
			handlers = append(handlers, h)
		}
	}
	h := logging.NewHandler(handlers...)
	if slices.Contains(l.handlers, logging.DatadogHandler) {
		h = logging.NewDatadogHandler(tracking.Datadog(), h)
	}
	log := slog.New(h)
	slog.SetDefault(log)
}

func (l Logging) Validate() error {
	return validation.ValidateStruct(&l,
		validation.Field(&l.handlers, validation.Each(validation.In(logging.LoggingHandlersToInf()...))),
		validation.Field(&l.Sentry, validation.When(slices.Contains(l.handlers, logging.SentryHandler), validation.NotNil)),
		validation.Field(&l.Rollbar, validation.When(slices.Contains(l.handlers, logging.RollbarHandler), validation.NotNil)),
	)
}

func (l Logging) Close() {
	if l.closer != nil {
		l.closer()
	}
}

func (l *Logging) UnmarshalYAML(unmarshal func(any) error) error {
	if err := defaults.Set(l); err != nil {
		return err
	}
	type plain Logging
	ll := (plain)(*l)
	if err := unmarshal(&ll); err != nil {
		return err
	}
	*l = (Logging)(ll)
	switch v := l.Handler.(type) {
	case string:
		l.handlers = append(l.handlers, logging.LoggingHandle(v))
	case []any:
		for _, key := range v {
			v, ok := key.(string)
			if ok {
				l.handlers = append(l.handlers, logging.LoggingHandle(v))
			}
		}
	}
	return nil
}
