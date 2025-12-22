package config

import (
	"errors"
)

var (
	ErrNotFoundAggregateTypeTopic = errors.New("not found aggregate type topic")
)

type SNS struct {
	IEndpoint `yaml:",inline"`
	Resources Resource `yaml:"resources"`
}
