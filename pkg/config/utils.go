package config

import (
	"bytes"
	"io"
	"os"

	template "github.com/compose-spec/compose-go/v2/template"
	"github.com/creasty/defaults"
	"github.com/goccy/go-yaml"
)

type TConfig interface {
	OutboxPolling
}

func newConfig[T TConfig](r io.Reader) (*T, error) {
	var cfg T
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	by, err := template.Substitute(string(buf), os.LookupEnv)
	if err != nil {
		return nil, err
	}
	if err := yaml.NewDecoder(bytes.NewBufferString(by)).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func loadConfig[T TConfig](filePath string) (*T, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	config, err := newConfig[T](f)
	if err != nil {
		return nil, err
	}
	if err := defaults.Set(config); err != nil {
		return nil, err
	}
	if err := Validate(&config); err != nil {
		return nil, err
	}
	switch t := any(&config).(type) {
	case *Config:
		setConfig(t)
		t.Logging.SetupLog()
	case *OutboxPolling:
		setConfig(&t.Config)
		t.Logging.SetupLog()
	}
	return config, nil
}
