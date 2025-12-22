package healthcheck

import "context"

type Ping interface {
	PingContext(ctx context.Context) error
}
