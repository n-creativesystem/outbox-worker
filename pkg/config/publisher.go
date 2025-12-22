package config

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	ErrNotFoundProducer = errors.New("not found producer")
)

type Publisher struct {
	AWS          AWS          `yaml:"aws"`
	RefetchTimer RefetchTimer `yaml:"refetchTimer"`
}

func (p *Publisher) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.AWS),
	)
}

func (p *Publisher) WhitelistResources() []string {
	return p.AWS.WhitelistResources()
}

func (p *Publisher) BlacklistResources() []string {
	return p.AWS.BlacklistResources()
}

func (p *Publisher) SkipEvents() []string {
	return p.AWS.SkipEvents()
}
