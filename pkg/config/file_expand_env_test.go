package config

import (
	"io"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

type textScanner struct {
	io.Reader
}

func (t *textScanner) Close() error {
	return nil
}

func TestExpandEnvReader(t *testing.T) {
	data := `key: ${TEST}`
	t.Setenv("TEST", "Value")
	var value = struct {
		Key string `yaml:"key"`
	}{}
	scanner := &textScanner{Reader: strings.NewReader(data)}
	f, err := NewExpandEnvWithReader(scanner)
	require.NoError(t, err)
	err = yaml.NewDecoder(f).Decode(&value)
	require.NoError(t, err)
	require.Equal(t, "Value", value.Key)
}
