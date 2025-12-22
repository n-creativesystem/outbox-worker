package polling

import "context"

type Poller interface {
	Start(ctx context.Context) error
}
