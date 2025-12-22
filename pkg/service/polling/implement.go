package polling

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/n-creativesystem/go-packages/lib/trace"
	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/n-creativesystem/outbox-worker/pkg/domain/outbox"
	"github.com/n-creativesystem/outbox-worker/pkg/interfaces"
	svcerr "github.com/n-creativesystem/outbox-worker/pkg/service/errors"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
	"golang.org/x/time/rate"
)

type OutboxPolling struct {
	cfg       *config.OutboxPolling
	repo      outbox.Repository
	timeNow   func() time.Time
	publisher interfaces.Publisher
}

type OutboxPollingOption interface {
	apply(opt *OutboxPolling)
}

type outboxPollingOptionFn func(opt *OutboxPolling)

func (fn outboxPollingOptionFn) apply(opt *OutboxPolling) {
	fn(opt)
}

func WithTimeNow(fn func() time.Time) OutboxPollingOption {
	return outboxPollingOptionFn(func(opt *OutboxPolling) {
		opt.timeNow = fn
	})
}

func NewPoller(
	ctx context.Context,
	cfg *config.OutboxPolling,
	repo outbox.Repository,
	publisher interfaces.Publisher,
	opts ...OutboxPollingOption,
) (*OutboxPolling, error) {
	opt := &OutboxPolling{
		cfg:       cfg,
		repo:      repo,
		publisher: publisher,
		timeNow:   time.Now,
	}
	for _, o := range opts {
		o.apply(opt)
	}

	return opt, nil
}

func (p *OutboxPolling) Start(ctx context.Context) error {
	slog.InfoContext(ctx, "Start polling process.")
	return p.processing(ctx)
}

func (p *OutboxPolling) processing(ctx context.Context) error {
	pollingTimer := time.NewTicker(p.cfg.OutboxConfig.PollingInterval)
	for {
		events, err := p.repo.FindEvents(
			ctx,
			p.cfg.Publisher.WhitelistResources(),
			p.cfg.Publisher.BlacklistResources(),
			p.cfg.Publisher.SkipEvents(),
		)
		if err != nil {
			slog.ErrorContext(ctx, err.Error())
		}
		if len(events) > 0 {
			p.processingEvent(ctx, events)
		}

		select {
		case <-pollingTimer.C:
			slog.InfoContext(ctx, "Resume polling.")
			continue
		case <-ctx.Done():
			pollingTimer.Stop()
			return ctx.Err()
		}
	}
}

func (p *OutboxPolling) processingEvent(ctx context.Context, events outbox.Events) {
	ctx = trace.StartSpan(ctx, "processingEvent")
	defer func() { trace.EndSpan(ctx, nil) }()

	var wg sync.WaitGroup
	mp := eventToGroupingByAggregateType(events)
	for aggregateType := range mp {
		wg.Add(1)
		event := mp[aggregateType]
		go func() {
			defer wg.Done()
			p.processingGroupingByAggregateTypeEvents(ctx, event)
		}()
	}
	wg.Wait()
}

// processingGroupingByAggregateTypeEvents aggregate type別に処理を行う
func (p *OutboxPolling) processingGroupingByAggregateTypeEvents(ctx context.Context, events outbox.Events) {
	ctx = trace.StartSpan(ctx, "processingGroupingByAggregateTypeEvents")
	defer func() { trace.EndSpan(ctx, nil) }()
	// TPSを準拠するために rate limit 設定を行う
	messageRate := p.cfg.OutboxConfig.Throughput
	// rate := time.Tick(3*time.Millisecond)
	rate := rate.NewLimiter(rate.Every(time.Second/time.Duration(messageRate)), messageRate)

	// aggregateIdをキーにメッセージの順番を生成
	mpGroupEvent := eventToGroupingAggregateId(events)
	// aggregateId別にgoroutineを管理する
	wg := sync.WaitGroup{}
	for mGroupId := range mpGroupEvent {
		wg.Add(1)
		go func(mGroupId string, groupEvents outbox.Events) {
			defer wg.Done()
			if mGroupId != "" {
				groupEvents.SortByID()
			}
			eventCount := len(groupEvents)
			for idx := 0; idx < eventCount; idx++ {
				// rate limit が処理できる場合はそのTopicに対して300TPS未満なので処理を行う
				event := groupEvents[idx]
				_ = rate.Wait(context.Background()) // nolint: errcheck
				isOk := p.publishEvent(ctx, event)
				if mGroupId != "" && !isOk {
					// メッセージグループがありかつ失敗の場合はこれ以上送らない
					break
				}
			}
		}(mGroupId, mpGroupEvent[mGroupId])
	}
	wg.Wait()
}

func (p *OutboxPolling) publishEvent(ctx context.Context, event *outbox.Event) (isOk bool) {
	ctx = trace.StartSpan(ctx, "publishEvent",
		attribute.Int64("ID", event.ID()),
		attribute.String("AggregateID", event.AggregateID()),
		attribute.String("AggregateType", event.AggregateType()),
		attribute.String("Event", event.Event()),
		attribute.String("ProducerName", event.ProducerName()),
		attribute.Int("RetryCount", event.RetryCount()),
	)
	defer func() { trace.EndSpan(ctx, nil) }()
	span := trace.SpanFromContext(ctx)
	log := slog.With(slog.Int64("ID", event.ID()))
	if event.CanNotRetry(p.timeNow()) {
		span.SetAttributes(
			attribute.Bool("Skip", true),
		)
		log.Warn("waiting for the retry")
		return false
	}
	span.SetAttributes(
		attribute.Bool("Skip", false),
	)
	if v := event.RetryAtString(); v != "" {
		span.SetAttributes(
			attribute.String("RetryTime", v),
		)
	}

	var skip bool
	defer func() {
		if skip {
			isOk = false
			return
		}

		if isOk {
			event.SetSentAt(p.timeNow())
		} else {
			event.Retry(p.cfg.OutboxConfig.RetryBackOff, p.timeNow)
			maxRetryErr := event.CheckMaxRetryCount(p.cfg.OutboxConfig.MaxRetryCount)
			if maxRetryErr {
				slog.With(
					slog.Int64("ID", event.ID()),
					slog.Int("RetryCount", event.RetryCount()),
				).ErrorContext(ctx, "RetryCount has reached its maximum value.")
				maxRetryErr = true
			}
			span.SetAttributes(
				attribute.Bool("MaxRetry", maxRetryErr),
				attribute.Int("RetryCount", event.RetryCount()),
				attribute.String("RetryTime", event.RetryAtString()),
			)
		}

		err := p.repo.Transaction(ctx, isOk, event)
		if err != nil {
			isOk = false
			log.Error(err.Error())
		}
	}()

	err := p.publisher.PublishOutbox(ctx, interfaces.Outbox{
		AggregateId:   event.AggregateID(),
		AggregateType: event.AggregateType(),
		EventType:     event.Event(),
		Payload:       event.Payload(),
		ProducerName:  p.cfg.OutboxConfig.ProducerName,
	})
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		var e svcerr.NotFoundKeyError
		if errors.As(err, &e) {
			// メッセージ送信先が取得できない場合はログ上エラーにはするがリトライなどは行わない様にする
			span.AddEvent("exception",
				oteltrace.WithStackTrace(true),
				oteltrace.WithAttributes(
					attribute.String("Error", err.Error()),
					attribute.String("Key", event.AggregateType()),
				),
			)
			slog.WarnContext(ctx, e.Error())
			skip = true
			return false
		}
	}
	return err == nil
}
