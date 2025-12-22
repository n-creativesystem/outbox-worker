//go:generate go run -mod=mod github.com/golang/mock/mockgen -source=$GOFILE -package=$GOPACKAGE -destination=../mock/$GOPACKAGE/$GOFILE
package interfaces

import "context"

type Publisher interface {
	PublishOutbox(ctx context.Context, outbox Outbox) error
	FindResources(ctx context.Context) error
	RefetchResources(ctx context.Context) chan error
}

type BackendPublisher interface {
	PublishOutbox(ctx context.Context, outbox Outbox) (string, error)
	FindBackendResources(ctx context.Context) error
}
