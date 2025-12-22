package config

import (
	"fmt"
	"log/slog"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type OutboxPollingConfig struct {
	Schema          string        `yaml:"schema"`
	TableName       string        `yaml:"tableName" default:"outbox"`
	ProducerName    string        `yaml:"producerName"`
	MaxRetryCount   int           `yaml:"retryCount" default:"10"`
	PollingInterval time.Duration `yaml:"pollingInterval" default:"5s"`
	RetryBackOff    time.Duration `yaml:"retryBackoff" default:"20s"`
	Throughput      int           `yaml:"throughput" default:"3000"`
	FindEventLimit  uint          `yaml:"findEventLimit" default:"1000"`
}

var (
	_ validation.Validatable = (*OutboxPollingConfig)(nil)
)

func (o OutboxPollingConfig) Validate() error {
	err := ValidateStruct(&o,
		validation.Field(&o.ProducerName, validation.Required),
		validation.Field(&o.TableName, validation.Required),
		validation.Field(&o.FindEventLimit, validation.Max(uint(10000))),
	)
	if err != nil {
		return err
	}
	if err := Validate(o.Throughput, validation.Max(3000)); err != nil {
		slog.Warn(fmt.Sprintf("throughput: %s", err.Error()))
	}
	return nil
}

func (o *OutboxPollingConfig) Table() string {
	if o.Schema != "" {
		return fmt.Sprintf("%s.%s", o.Schema, o.TableName)
	}
	return o.TableName
}

type OutboxPolling struct {
	Config       `yaml:",inline"`
	OutboxConfig *OutboxPollingConfig `yaml:"outbox"`
}

var (
	_ validation.Validatable = (*OutboxPolling)(nil)
)

func LoadOutboxPollingConfig(filePath string) (*OutboxPolling, error) {
	config, err := loadConfig(filePath)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (conf OutboxPolling) Validate() error {
	return validation.ValidateStruct(&conf,
		validation.Field(&conf.Config),
		validation.Field(&conf.OutboxConfig, validation.NotNil),
	)
}
