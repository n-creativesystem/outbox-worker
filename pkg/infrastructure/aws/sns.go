package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/n-creativesystem/outbox-worker/pkg/interfaces/aws"
)

func NewSNSClient(ctx context.Context, conf *config.AWS) (aws.SNSClient, error) {
	endpoint := NewEndpoint(WithSNSEndpoint(conf.SNS.Endpoint))
	aConfig := conf.AWSConfig()
	aConfig = append(aConfig, endpoint.EndpointResolver())
	awsConfig, err := NewConfig(ctx, aConfig...)
	if err != nil {
		return nil, err
	}
	return sns.NewFromConfig(awsConfig), nil
}
