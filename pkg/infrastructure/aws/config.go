package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

func NewConfig(ctx context.Context, opts ...func(*config.LoadOptions) error) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return cfg, err
	}
	otelaws.AppendMiddlewares(&cfg.APIOptions)
	return cfg, nil
}
