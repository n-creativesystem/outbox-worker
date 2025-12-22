package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/n-creativesystem/go-packages/lib/trace"
	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/n-creativesystem/outbox-worker/pkg/interfaces"
)

type service struct {
	cfg       *config.Publisher
	publisher interfaces.BackendPublisher
}

func New(cfg *config.Publisher, publisher interfaces.BackendPublisher) interfaces.Publisher {
	return &service{
		cfg:       cfg,
		publisher: publisher,
	}
}

func (svc *service) PublishOutbox(ctx context.Context, outbox interfaces.Outbox) (rErr error) {
	ctx = trace.StartSpan(ctx, "PublishOutbox")
	defer func() { trace.EndSpan(ctx, rErr) }()
	slog.With(
		slog.String("AggregateType", outbox.AggregateType),
		slog.String("AggregateId", outbox.AggregateId),
	).InfoContext(ctx, "Receive outbox.")
	msgId, err := svc.publisher.PublishOutbox(ctx, outbox)
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Published outbox.", "MessageId", msgId)
	return nil
}

func (svc *service) FindResources(ctx context.Context) error {
	return svc.publisher.FindBackendResources(ctx)
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
