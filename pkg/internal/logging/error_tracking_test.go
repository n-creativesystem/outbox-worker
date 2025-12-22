package logging

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorTracking(t *testing.T) {
	require := require.New(t)
	buf := bytes.Buffer{}
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	h := NewErrorTracking(handler)
	log := slog.New(h)
	log.Error("aaa")
	require.Contains(buf.String(), "msg=aaa")
}

func TestIgnoreErrorTracking(t *testing.T) {
	require := require.New(t)
	buf := bytes.Buffer{}
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	h := NewErrorTracking(handler)
	log := slog.New(h)
	log.With(slog.Any("aaa", "bbb")).With(IgnoreTracing).Error("aaa")
	require.Equal(buf.String(), "")
}
