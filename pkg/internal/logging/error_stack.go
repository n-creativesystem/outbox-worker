package logging

import (
	"fmt"
	"log/slog"

	"github.com/cockroachdb/errors"
)

func WithStack(err error) slog.Attr {
	if err == nil {
		return slog.Attr{}
	}
	return slog.String("stack", fmt.Sprintf("%+v", errors.WithStack(err)))
}
