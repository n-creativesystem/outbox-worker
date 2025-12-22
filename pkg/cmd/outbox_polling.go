package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	pkgErr "github.com/cockroachdb/errors"
	"github.com/n-creativesystem/go-packages/lib/trace"
	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/n-creativesystem/outbox-worker/pkg/infrastructure/repository/outbox"
	"github.com/n-creativesystem/outbox-worker/pkg/internal/logging"
	"github.com/n-creativesystem/outbox-worker/pkg/service"
	"github.com/n-creativesystem/outbox-worker/pkg/service/healthcheck"
	"github.com/n-creativesystem/outbox-worker/pkg/service/polling"
	"github.com/n-creativesystem/outbox-worker/pkg/utils"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

type outboxPollingArgs struct {
	mainArgs
}

func outboxPollingCommand() *cobra.Command {
	var (
		args outboxPollingArgs
	)

	cmd := cobra.Command{
		Use: "polling",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			return executePolling(ctx, &args)
		},
	}
	flag := cmd.Flags()
	args.setpflag(flag)
	return &cmd
}

func executePolling(parent context.Context, args *outboxPollingArgs) error {
	ctx, stop := signal.NotifyContext(parent, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	config, err := config.LoadOutboxPollingConfig(args.configFilePath)
	if err != nil {
		return pkgErr.Wrap(err, "config.LoadOutboxPollingConfig")
	}
	defer config.Logging.Close()

	_, cleanup, err := trace.NewTracerProvider(
		ctx,
		trace.WithEnabled(config.Tracking.Enabled),
		trace.WithAgentAddr(config.Tracking.AgentAddr),
		trace.WithServiceName(config.Tracking.ServiceName),
		trace.WithEnvironment(config.Tracking.Environment),
	)
	if err != nil {
		return pkgErr.Wrap(err, "tracking.NewTracerProvider")
	}
	defer cleanup()

	publisher, err := getPublisher(ctx, config.Publisher)
	if err != nil {
		return pkgErr.Wrap(err, "getPublisher")
	}
	if err := publisher.FindResources(ctx); err != nil {
		return pkgErr.Wrap(err, "publisher.FindResources")
	}
	errCh := publisher.RefetchResources(ctx)
	go func() {
		for err := range errCh {
			slog.With(logging.WithStack(err)).Error(fmt.Sprintf("Refetch resources: %v\n", err))
		}
	}()

	dbClient, err := outbox.NewClient(ctx, config, outbox.WithTimeNow(utils.NowInJST))
	if err != nil {
		return err
	}
	poller, err := polling.NewPoller(ctx, config, dbClient, publisher)
	if err != nil {
		return err
	}

	var healthCheckServer service.HealthCheck = healthcheck.New(dbClient, args.healthCheckAddr)
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return healthCheckServer.Start(ctx)
	})
	group.Go(func() error {
		return poller.Start(ctx)
	})

	if err := group.Wait(); err != nil && err != context.Canceled {
		slog.With(logging.WithStack(err)).ErrorContext(ctx, err.Error())
	}
	stop()
	slog.InfoContext(ctx, "shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	<-ctx.Done()
	return nil
}
