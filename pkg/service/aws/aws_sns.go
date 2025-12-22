package aws

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	originAWS "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/cockroachdb/errors"
	"github.com/google/uuid"
	"github.com/n-creativesystem/go-packages/lib/trace"
	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/n-creativesystem/outbox-worker/pkg/interfaces"
	"github.com/n-creativesystem/outbox-worker/pkg/interfaces/aws"
	"github.com/n-creativesystem/outbox-worker/pkg/internal/logging"
	"go.opentelemetry.io/otel/attribute"
)

type snsInfo struct {
	Arn                           string
	IsMessageDeduplicationSetting bool
}

type awsSNS struct {
	client aws.SNSClient
	conf   *config.AWS

	mpResourceNameToArn Store[*snsInfo]
}

var (
	_ interfaces.BackendPublisher = (*awsSNS)(nil)
)

func NewAWSSNS(ctx context.Context, client aws.SNSClient, conf *config.AWS) interfaces.BackendPublisher {
	return newAWSSNS(ctx, client, conf)
}

func newAWSSNS(_ context.Context, client aws.SNSClient, conf *config.AWS) *awsSNS {
	return &awsSNS{
		client:              client,
		conf:                conf,
		mpResourceNameToArn: Store[*snsInfo]{},
	}
}

func (p *awsSNS) PublishOutbox(ctx context.Context, outbox interfaces.Outbox) (_ string, rErr error) {
	ctx = trace.StartSpan(ctx, "PublishOutbox",
		attribute.String("AggregateId", outbox.AggregateId),
		attribute.String("AggregateType", outbox.AggregateType),
		attribute.String("Event", outbox.EventType),
		attribute.String("ProducerName", outbox.ProducerName),
	)
	defer func() { trace.EndSpan(ctx, rErr) }()
	resource, err := p.mpResourceNameToArn.Load(ctx, outbox.AggregateType)
	if err != nil {
		return "", err
	}
	input := &sns.PublishInput{
		Message:        originAWS.String(outbox.Payload),
		MessageGroupId: originAWS.String(outbox.AggregateId),
		TargetArn:      originAWS.String(resource.Arn),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Event": {
				DataType:    originAWS.String("String"),
				StringValue: originAWS.String(outbox.EventType),
			},
			"Producer": {
				DataType:    originAWS.String("String"),
				StringValue: originAWS.String(outbox.ProducerName),
			},
		},
	}
	if resource.IsMessageDeduplicationSetting {
		input.MessageDeduplicationId = originAWS.String(uuid.NewString())
	}
	if output, err := p.client.Publish(ctx, input); err != nil {
		return "", err
	} else {
		return *output.MessageId, nil
	}
}

func (p *awsSNS) FindBackendResources(ctx context.Context) (rErr error) {
	ctx = trace.StartSpan(ctx, "FindBackendResources")
	defer func() { trace.EndSpan(ctx, rErr) }()

	input := sns.ListTopicsInput{}
	token, err := p.findBackendResources(ctx, &input)
	if err != nil {
		return errors.Wrap(err, "sns.FindBackendResources")
	}
	for token != nil {
		input = sns.ListTopicsInput{
			NextToken: token,
		}
		token, err = p.findBackendResources(ctx, &input)
		if err != nil {
			return errors.Wrap(err, "sns.FindBackendResources")
		}
	}
	return nil
}

func (p *awsSNS) findBackendResources(ctx context.Context, input *sns.ListTopicsInput) (_ *string, rErr error) {
	ctx = trace.StartSpan(ctx, "FindBackendResources")
	defer func() { trace.EndSpan(ctx, rErr) }()

	output, err := p.client.ListTopics(ctx, input)
	if err != nil {
		return nil, err
	}
	for _, topic := range output.Topics {
		if topic.TopicArn == nil {
			continue
		}
		pArn, err := arn.Parse(*topic.TopicArn)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Unexpected Arn %s", *topic.TopicArn))
		}

		attrs, err := p.client.GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
			TopicArn: topic.TopicArn,
		})
		if err != nil {
			if !p.conf.SNS.Resources.IsIgnoreFetchErrorResources(pArn.Resource) {
				slog.With(logging.WithStack(err)).WarnContext(ctx, err.Error())
			}
			continue
		}
		info := snsInfo{
			Arn: *topic.TopicArn,
		}
		isFifo := false
		fifoTopicValue, ok := attrs.Attributes["FifoTopic"]
		if ok {
			v, _ := strconv.ParseBool(fifoTopicValue) // nolint: errcheck
			isFifo = v
		}
		if isFifo {
			contentBasedValue, ok := attrs.Attributes["ContentBasedDeduplication"]
			if ok {
				v, _ := strconv.ParseBool(contentBasedValue) // nolint: errcheck
				info.IsMessageDeduplicationSetting = !v
			}
		}
		p.mpResourceNameToArn.Add(ctx, pArn.Resource, &info)
	}
	return output.NextToken, nil
}
