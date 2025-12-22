package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOutboxPollingValidation(t *testing.T) {
	v := OutboxPollingConfig{}
	err := Validate(&v)
	require.Equal(t, "producerName: cannot be blank; tableName: cannot be blank.", err.Error())
}
