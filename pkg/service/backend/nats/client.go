package nats

import (
	"context"
	"crypto/sha256"
	"fmt"
	"runtime"

	"github.com/n-creativesystem/go-packages/lib/trace"
	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/n-creativesystem/outbox-worker/pkg/interfaces"
	"github.com/n-creativesystem/outbox-worker/pkg/internal/rn"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel/attribute"
)

type natsClient struct {
	js jetstream.JetStream
}

var (
	_ interfaces.BackendPublisher = (*natsClient)(nil)
)

func NewNatsClient(cfg *config.NATS) (interfaces.BackendPublisher, error) {
	return newNatsClient(cfg)
}

func newNatsConn(cfg *config.NATS, natsOpts ...nats.Option) (*nats.Conn, error) {
	opts, err := cfg.Options()
	if err != nil {
		return nil, err
	}
	opts = append(opts, natsOpts...)
	return nats.Connect(cfg.Server, opts...)
}

func newNatsClient(cfg *config.NATS, opts ...nats.Option) (*natsClient, error) {
	nc, err := newNatsConn(cfg, opts...)
	if err != nil {
		return nil, err
	}
	var (
		js     jetstream.JetStream
		domain string
	)
	if cfg.Domain != nil {
		domain = *cfg.Domain
	}
	if domain != "" {
		js, err = jetstream.NewWithDomain(nc, domain)
		if err != nil {
			return nil, err
		}
	} else {
		js, err = jetstream.New(nc)
		if err != nil {
			return nil, err
		}
	}
	client := &natsClient{
		js: js,
	}
	runtime.AddCleanup(client, func(con *nats.Conn) {
		con.Close()
	}, client.js.Conn())
	return client, nil
}

func (c *natsClient) Publish(ctx context.Context, outbox interfaces.Outbox) (_ string, rErr error) {
	resourceName, _ := rn.Parse(outbox.AggregateType)
	ctx = trace.StartSpan(ctx, "PublishOutbox",
		attribute.String("AggregateId", outbox.AggregateId),
		attribute.String("AggregateType", outbox.AggregateType),
		attribute.String("Event", outbox.EventType),
		attribute.String("ProducerName", outbox.ProducerName),
		attribute.String("Resource", resourceName.Resource),
	)
	defer func() { trace.EndSpan(ctx, rErr) }()

	subject := fmt.Sprintf("%s.%s", resourceName.Resource, outbox.AggregateId)
	msg := nats.NewMsg(subject)
	payload := []byte(outbox.Payload)
	msg.Data = payload
	msgId := fmt.Sprintf("%x", sha256.Sum256(payload)) // 16進数に変換
	msg.Header.Add("Event", outbox.EventType)
	msg.Header.Add("Producer", outbox.ProducerName)

	if pubAck, err := c.js.PublishMsg(
		ctx,
		msg,
		jetstream.WithMsgID(msgId),
	); err != nil {
		return "", err
	} else {
		return fmt.Sprintf("%d", pubAck.Sequence), nil
	}
}

func (c *natsClient) FindResources(ctx context.Context) error {
	return nil
}
