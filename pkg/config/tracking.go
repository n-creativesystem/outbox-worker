package config

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/n-creativesystem/outbox-worker/pkg/internal/logging"
	"github.com/n-creativesystem/outbox-worker/pkg/internal/version"
)

type Tracking struct {
	Enabled     bool   `yaml:"enabled" default:"false"`
	AgentAddr   string `yaml:"agentAddr"`
	ServiceName string `yaml:"serviceName" default:"outbox-worker"`
	Environment string `yaml:"environment" default:"prod"`
}

func (t Tracking) Validate() error {
	return validation.ValidateStruct(&t,
		validation.Field(&t.AgentAddr, validation.When(t.Enabled, validation.Required)),
	)
}

func (t Tracking) Datadog() logging.Service {
	return &datadogService{tracking: t}
}

type datadogService struct {
	tracking Tracking
}

func (s *datadogService) ServiceName() string {
	return s.tracking.ServiceName
}

func (s *datadogService) Environment() string {
	return s.tracking.Environment
}

func (s *datadogService) Version() string {
	return version.Version
}
