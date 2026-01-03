//go:generate go run -mod=mod go.uber.org/mock/mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../mock/$GOPACKAGE/$GOFILE
package interfaces

import "context"

type Publisher interface {
	Publish(ctx context.Context, outbox Outbox) error
	FindResources(ctx context.Context) error
	RefetchResources(ctx context.Context) chan error
}

type BackendPublisher interface {
	Publish(ctx context.Context, outbox Outbox) (string, error)
	FindResources(ctx context.Context) error
}
