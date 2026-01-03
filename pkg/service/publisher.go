package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/n-creativesystem/go-packages/lib/trace"
	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/n-creativesystem/outbox-worker/pkg/interfaces"
	"github.com/n-creativesystem/outbox-worker/pkg/internal/rn"
)

type service struct {
	cfg              *config.Publisher
	publisher        map[string]interfaces.BackendPublisher
	defaultPublisher interfaces.BackendPublisher
}

func New(cfg *config.Publisher, publisher map[string]interfaces.BackendPublisher) interfaces.Publisher {
	def, _ := publisher["default"]
	return &service{
		cfg:              cfg,
		publisher:        publisher,
		defaultPublisher: def,
	}
}

func (svc *service) Publish(ctx context.Context, outbox interfaces.Outbox) (rErr error) {
	ctx = trace.StartSpan(ctx, "PublishOutbox")
	defer func() { trace.EndSpan(ctx, rErr) }()
	slog.With(
		slog.String("AggregateType", outbox.AggregateType),
		slog.String("AggregateId", outbox.AggregateId),
	).InfoContext(ctx, "Receive outbox.")
	publisher := svc.defaultPublisher
	if rn.IsRN(outbox.AggregateType) {
		rn, _ := rn.Parse(outbox.AggregateType)
		if v, ok := svc.publisher[rn.Service]; ok {
			publisher = v
		}
	}
	if publisher == nil {
		return fmt.Errorf("failed to find publisher")
	}
	msgId, err := publisher.Publish(ctx, outbox)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Published outbox.", "MessageId", msgId)
	return nil
}

func (svc *service) FindResources(ctx context.Context) error {
	var err error
	for _, publisher := range svc.publisher {
		err = errors.Join(err, publisher.FindResources(ctx))
	}
	return err
}

func (svc *service) RefetchResources(ctx context.Context) chan error {
	if !svc.cfg.RefetchTimer.Enabled {
		return nil
	}
	errCh := make(chan error)
	go svc.refetchResources(ctx, errCh)
	return errCh
}

func (svc *service) refetchResources(ctx context.Context, errCh chan error) {
	defer close(errCh)
	interval := svc.cfg.RefetchTimer.Interval
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := svc.FindResources(ctx); err != nil {
				errCh <- errors.Wrap(err, "service.refetchResources")
			}
		}
	}
}
