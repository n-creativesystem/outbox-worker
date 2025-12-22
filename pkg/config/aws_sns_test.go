package config

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

func TestAWSSNSConfig(t *testing.T) {
	data := `endpoint: http://localstack:4566
`
	var config SNS
	err := yaml.Unmarshal([]byte(data), &config)
	require.NoError(t, err)
	require.Equal(t, config.Endpoint, "http://localstack:4566")
}

func TestAWSSNSConfigWithGoTemplate(t *testing.T) {
	data := `endpoint: http://localstack:4566
`
	var config SNS
	err := yaml.Unmarshal([]byte(data), &config)
	require.NoError(t, err)
	require.Equal(t, config.Endpoint, "http://localstack:4566")
}
