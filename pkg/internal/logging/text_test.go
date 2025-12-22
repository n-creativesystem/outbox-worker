package logging

import (
	"bytes"
	"testing"

	"log/slog"

	"github.com/stretchr/testify/require"
)

func TestText(t *testing.T) {
	buf := &bytes.Buffer{}
	f := func() []slog.Attr {
		return []slog.Attr{
			slog.String("key", "test"),
		}
	}
	handler := NewTextHandler(WithWriter(buf))
	slog.SetDefault(slog.New(handler))
	slog.With(ToInterface(f())...).Info("This is test.")
	require.Contains(t, buf.String(), "msg=\"This is test.\" key=test")
}
