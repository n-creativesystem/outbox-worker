package config

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	ErrNotFoundProducer = errors.New("not found producer")
)

type Publisher struct {
	AWS          *AWS         `yaml:"aws"`
	Nats         *NATS        `yaml:"nats"`
	RefetchTimer RefetchTimer `yaml:"refetchTimer"`
}

func (p *Publisher) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.AWS),
	)
}

func (p *Publisher) WhitelistResources() []string {
	values := make([]string, 0, 20)
	if p.AWS != nil {
		values = append(values, p.AWS.WhitelistResources()...)
	}
	return values
}

func (p *Publisher) BlacklistResources() []string {
	values := make([]string, 0, 20)
	if p.AWS != nil {
		values = append(values, p.AWS.BlacklistResources()...)
	}
	return values
}

func (p *Publisher) SkipEvents() []string {
	values := make([]string, 0, 20)
	if p.AWS != nil {
		values = append(values, p.AWS.SkipEvents()...)
	}
	return values
}
