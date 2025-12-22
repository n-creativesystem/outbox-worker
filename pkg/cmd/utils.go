package cmd

import (
	"context"
	"errors"

	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/n-creativesystem/outbox-worker/pkg/infrastructure/aws"
	"github.com/n-creativesystem/outbox-worker/pkg/interfaces"
	"github.com/n-creativesystem/outbox-worker/pkg/service"
	backend "github.com/n-creativesystem/outbox-worker/pkg/service/aws"
)

func getPublisher(ctx context.Context, conf *config.Publisher) (interfaces.Publisher, error) {
	noSelectedErr := errors.New("set the Publisher")
	var publisher interfaces.BackendPublisher
	if conf == nil {
		return nil, noSelectedErr
	}
	if client, err := aws.NewSNSClient(ctx, &conf.AWS); err != nil {
		return nil, err
	} else {
		publisher = backend.NewAWSSNS(ctx, client, &conf.AWS)
	}

	if publisher == nil {
		return nil, noSelectedErr
	}
	return service.New(conf, publisher), nil
}
