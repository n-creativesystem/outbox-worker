package logging

import (
	"log/slog"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/rollbar/rollbar-go"
	slogrollbar "github.com/samber/slog-rollbar/v2"
)

type RollbarConfig struct {
	Level string `yaml:"level" default:"warn"`
	Token string `yaml:"token"`

	client *rollbar.Client
	Client *http.Client
}

func (c RollbarConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Token, validation.Required),
	)
}

func (c *RollbarConfig) Init(env, version, serverRoot string) {
	client := rollbar.NewAsync(c.Token, env, version, "", serverRoot)
	if c.Client != nil {
		client.SetHTTPClient(c.Client)
	}
	c.client = client
}

func (c *RollbarConfig) Close() {
	if c.client != nil {
		_ = c.client.Close() // nolint: errcheck
	}
}

func (c *RollbarConfig) getLevel() slog.Level {
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

func NewRollbarHandler(conf *RollbarConfig) Handle {
	option := slogrollbar.Option{
		Level:  conf.getLevel(),
		Client: conf.client,
	}
	return NewErrorTracking(NewAsyncHandler(option.NewRollbarHandler()))
}
