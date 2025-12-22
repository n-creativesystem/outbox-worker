package config

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

func TestAWSConfig(t *testing.T) {
	data := `accessKey: dummy
secretKey: dummy
`
	var config AWS
	err := yaml.Unmarshal([]byte(data), &config)
	require.NoError(t, err)
	require.Equal(t, config.AccessKey, "dummy")
	require.Equal(t, config.SecretKet, "dummy")
}
