package cmd

import (
	"context"
	"errors"

	"github.com/n-creativesystem/outbox-worker/pkg/config"
	infaws "github.com/n-creativesystem/outbox-worker/pkg/infrastructure/aws"
	"github.com/n-creativesystem/outbox-worker/pkg/interfaces"
	"github.com/n-creativesystem/outbox-worker/pkg/service"
	"github.com/n-creativesystem/outbox-worker/pkg/service/backend/aws"
	"github.com/n-creativesystem/outbox-worker/pkg/service/backend/nats"
)

func getPublisher(ctx context.Context, conf *config.Publisher) (interfaces.Publisher, error) {
	noSelectedErr := errors.New("set the Publisher")
	if conf == nil {
		return nil, noSelectedErr
	}
	publisher := map[string]interfaces.BackendPublisher{}
	if conf.AWS != nil {
		if v, err := snsPublisher(ctx, conf.AWS); err != nil {
			return nil, err
		} else {
			publisher["sns"] = v
		}
	}
	if conf.Nats != nil {
		if pub, err := nats.NewNatsClient(conf.Nats); err != nil {
			return nil, err
		} else {
			publisher["nats"] = pub
		}
	}

	if len(publisher) == 0 {
		return nil, noSelectedErr
	}
	return service.New(conf, publisher), nil
}

func snsPublisher(ctx context.Context, awsConf *config.AWS) (interfaces.BackendPublisher, error) {
	if client, err := infaws.NewSNSClient(ctx, awsConf); err != nil {
		return nil, err
	} else {
		return aws.NewAWSSNS(ctx, client, awsConf), nil
	}
}
