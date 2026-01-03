package nats

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/n-creativesystem/outbox-worker/pkg/interfaces"
	"github.com/n-creativesystem/outbox-worker/pkg/internal/tests"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublish(t *testing.T) {
	ctx := t.Context()
	natsServer, err := tests.RunNatsContainer()
	require.NoError(t, err)
	defer func() {
		natsServer.Shutdown()
		natsServer.WaitForShutdown()
	}()
	err = natsServer.AddStream("TEST_EVENTS", "tests.>")
	require.NoError(t, err)

	err = natsServer.AddConsumer("TEST_EVENTS", "TestProcessor", "tests.>")
	require.NoError(t, err)

	cfg := &config.NATS{
		Server:     natsServer.ClientURL(),
		ClientName: "test-server",
		UserInfo: &config.NatsUserInfo{
			User:     "user",
			Password: "password",
		},
	}
	payload := fmt.Sprintf(`{"key": "%s"}`, uuid.NewString())
	c, err := newNatsClient(cfg, nats.InProcessServer(natsServer))
	require.NoError(t, err)
	ret, err := c.Publish(ctx, interfaces.Outbox{
		AggregateId:   "1:1",
		AggregateType: "tests",
		EventType:     "event",
		Payload:       payload,
		ProducerName:  "test",
	})
	require.NoError(t, err)
	require.NotEmpty(t, ret)
	cc, err := c.js.Consumer(context.Background(), "TEST_EVENTS", "TestProcessor")
	require.NoError(t, err)
	var result bool
	batchMsg, err := cc.Fetch(1, jetstream.FetchMaxWait(10*time.Second))
	require.NoError(t, err)
	for msg := range batchMsg.Messages() {
		err := msg.DoubleAck(t.Context())
		require.NoError(t, err)
		fmt.Println(string(msg.Data()))
		assert.Equal(t, payload, string(msg.Data()))
		result = true
	}
	require.NoError(t, batchMsg.Error())
	require.True(t, result)
}

// func test_consumer(t *testing.T, conn *nats.Conn) <-chan struct{} {
// 	done := make(chan struct{})
// 	js, err := jetstream.New(conn)
// 	require.NoError(t, err)
// 	cc, err := js.Consumer(context.Background(), "TEST_EVENTS", "TestProcessor")
// 	require.NoError(t, err)
// 	wg := sync.WaitGroup{}
// 	wg.Add(1)
// 	cc.Consume(func(msg jetstream.Msg) {
// 		err := msg.Ack()
// 		assert.NoError(t, err)
// 		assert.Equal(t, `{"key": "value"}`, string(msg.Data()))
// 		fmt.Println(string(msg.Data()))
// 		wg.Done()
// 	})
// 	go func() {
// 		defer close(done)
// 		wg.Wait()
// 	}()
// 	return done
// }
