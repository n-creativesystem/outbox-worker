package outbox

import (
	"context"
)

type Repository interface {
	FindEvents(
		ctx context.Context,
		whitelistResources []string,
		blacklistResources []string,
		skipEvents []string,
	) ([]*Event, error)
	Transaction(ctx context.Context, isOk bool, event *Event) (err error)
}
