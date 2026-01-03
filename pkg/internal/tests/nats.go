package tests

import (
	"context"
	"fmt"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	nats "github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type NatsContainer struct {
	*natsserver.Server

	nc *nats.Conn
	js jetstream.JetStream
}

func RunNatsContainer() (*NatsContainer, error) {
	opts := &natsserver.Options{
		JetStream:       true,
		DontListen:      true,
		ServerName:      "embedded-nats",
		Debug:           true,
		Logtime:         true,
		Trace:           false,
		TraceVerbose:    false,
		NoSystemAccount: true,
		Username:        "user",
		Password:        "password",
		SystemAccount:   natsserver.DEFAULT_SYSTEM_ACCOUNT,
		Accounts: []*natsserver.Account{
			natsserver.NewAccount(natsserver.DEFAULT_SYSTEM_ACCOUNT),
			natsserver.NewAccount("TEST_SERVICE"),
		},
		Users: []*natsserver.User{
			{
				Username: "user",
				Password: "password",
				Account:  natsserver.NewAccount("TEST_SERVICE"),
			},
		},
	}
	srv := natsserver.New(opts)
	srv.ConfigureLogger()
	srv.Start()
	if !srv.ReadyForConnections(4 * time.Second) {
		return nil, fmt.Errorf("not ready for connection")
	}
	account, err := srv.LookupAccount("TEST_SERVICE")
	if err != nil {
		return nil, err
	}
	err = account.EnableJetStream(map[string]natsserver.JetStreamAccountLimits{
		"": {
			MaxMemory: 100 * 1024 * 1024, // 100Mi
			MaxStore:  100 * 1024 * 1024, // 100Mi
		},
	}, nil)
	if err != nil {
		srv.Shutdown()
		return nil, err
	}
	nc, err := nats.Connect(
		srv.ClientURL(),
		nats.InProcessServer(srv),
		nats.UserInfo("user", "password"),
	)
	if err != nil {
		srv.Shutdown()
		return nil, err
	}
	js, err := jetstream.New(nc)
	if err != nil {
		srv.Shutdown()
		return nil, err
	}

	container := &NatsContainer{
		Server: srv,
		nc:     nc,
		js:     js,
	}
	return container, nil
}

func (c *NatsContainer) AddStream(stream, subject string) error {
	_, err := c.js.CreateStream(context.Background(), jetstream.StreamConfig{
		Name:       stream,
		Subjects:   []string{subject},
		Retention:  jetstream.LimitsPolicy,
		Replicas:   1,
		Discard:    jetstream.DiscardOld,
		MaxAge:     1 * time.Hour,
		Duplicates: 2 * time.Minute,
		Storage:    jetstream.MemoryStorage,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *NatsContainer) AddConsumer(stream, consumer, subject string) error {
	_, err := c.js.CreateConsumer(context.Background(), stream, jetstream.ConsumerConfig{
		Name:           consumer,
		FilterSubjects: []string{subject},
		AckPolicy:      jetstream.AckExplicitPolicy,
		MaxDeliver:     6,
		AckWait:        30 * time.Second,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *NatsContainer) Shutdown() {
	c.nc.Close()
	c.Server.Shutdown()
}
