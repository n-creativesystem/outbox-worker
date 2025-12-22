package aws

import (
	"context"
	"errors"
	"testing"

	svcerr "github.com/n-creativesystem/outbox-worker/pkg/service/errors"
	"github.com/stretchr/testify/require"
)

func TestNotFoundStore(t *testing.T) {
	s := Store[string]{}
	s.Add(context.Background(), "test", "value")
	_, err := s.Load(context.Background(), "t")
	require.Error(t, err)
	var e svcerr.NotFoundKeyError
	require.True(t, errors.As(err, &e))
}
