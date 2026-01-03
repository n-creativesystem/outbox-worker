package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/golang/mock/gomock"
	"github.com/n-creativesystem/outbox-worker/pkg/config"
	"github.com/n-creativesystem/outbox-worker/pkg/interfaces"
	mockAws "github.com/n-creativesystem/outbox-worker/pkg/mock/aws"
	"github.com/stretchr/testify/require"
)

func TestAWSSNSFindBackendResourcesWithNotPagination(t *testing.T) {
	output := sns.ListTopicsOutput{
		NextToken: nil,
		Topics: []types.Topic{
			{
				TopicArn: aws.String("arn:aws:sns:ap-northeast-1:000000000000:test-sns"),
			},
		},
	}
	outputAttribute := sns.GetTopicAttributesOutput{
		Attributes: map[string]string{},
	}
	require := require.New(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	client := mockAws.NewMockSNSClient(mockCtl)
	client.EXPECT().ListTopics(gomock.Any(), gomock.Any(), gomock.Any()).Return(&output, nil)
	client.EXPECT().GetTopicAttributes(gomock.Any(), gomock.Any(), gomock.Any()).Return(&outputAttribute, nil)
	p := newAWSSNS(context.Background(), client, &config.AWS{})
	err := p.FindResources(context.Background())
	require.NoError(err)
	v, err := p.mpResourceNameToArn.Load(context.Background(), "test-sns")
	require.NoError(err)
	require.Equal(v.Arn, "arn:aws:sns:ap-northeast-1:000000000000:test-sns")
	require.Equal(v.IsMessageDeduplicationSetting, false)
}

func TestAWSSNSFindBackendResourcesWithPagination(t *testing.T) {
	require := require.New(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()
	client := mockAws.NewMockSNSClient(mockCtl)
	client.EXPECT().ListTopics(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(
			func(ctx context.Context, input *sns.ListTopicsInput, opts ...func(*sns.Options)) (*sns.ListTopicsOutput, error) {
				if input.NextToken != nil && *input.NextToken == "NextToken" {
					return &sns.ListTopicsOutput{
						NextToken: nil,
						Topics: []types.Topic{
							{
								TopicArn: aws.String("arn:aws:sns:ap-northeast-1:000000000000:test-sns2"),
							},
						},
					}, nil
				} else {
					return &sns.ListTopicsOutput{
						NextToken: aws.String("NextToken"),
						Topics: []types.Topic{
							{
								TopicArn: aws.String("arn:aws:sns:ap-northeast-1:000000000000:test-sns"),
							},
							{
								TopicArn: aws.String("arn:aws:sns:ap-northeast-1:000000000000:test-sns.fifo"),
							},
						},
					}, nil
				}
			})

	outputAttribute := map[string]sns.GetTopicAttributesOutput{
		"arn:aws:sns:ap-northeast-1:000000000000:test-sns2": {
			Attributes: map[string]string{},
		},
		"arn:aws:sns:ap-northeast-1:000000000000:test-sns": {
			Attributes: map[string]string{},
		},
		"arn:aws:sns:ap-northeast-1:000000000000:test-sns.fifo": {
			Attributes: map[string]string{
				"FifoTopic":                 "true",
				"ContentBasedDeduplication": "false",
			},
		},
	}
	client.EXPECT().
		GetTopicAttributes(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(ctx context.Context, params *sns.GetTopicAttributesInput, optFns ...func(*sns.Options)) (*sns.GetTopicAttributesOutput, error) {
			v := outputAttribute[*params.TopicArn]
			return &v, nil
		})

	p := newAWSSNS(context.Background(), client, &config.AWS{})
	err := p.FindResources(context.Background())
	require.NoError(err)
	v, err := p.mpResourceNameToArn.Load(context.Background(), "test-sns")
	require.NoError(err)
	require.Equal(v.Arn, "arn:aws:sns:ap-northeast-1:000000000000:test-sns")
	require.Equal(v.IsMessageDeduplicationSetting, false)
	v, err = p.mpResourceNameToArn.Load(context.Background(), "test-sns.fifo")
	require.NoError(err)
	require.Equal(v.Arn, "arn:aws:sns:ap-northeast-1:000000000000:test-sns.fifo")
	require.Equal(v.IsMessageDeduplicationSetting, true)
	v, err = p.mpResourceNameToArn.Load(context.Background(), "test-sns2")
	require.NoError(err)
	require.Equal(v.Arn, "arn:aws:sns:ap-northeast-1:000000000000:test-sns2")
	require.Equal(v.IsMessageDeduplicationSetting, false)
}

func TestAWSSNSPublish_WithSNSTopicARN_AggregateType(t *testing.T) {
	require := require.New(t)
	mockCtl := gomock.NewController(t)
	defer mockCtl.Finish()

	client := mockAws.NewMockSNSClient(mockCtl)

	// Cache is keyed by topic name (ARN resource), Publish should accept full ARN and resolve to that key.
	awsSNS := newAWSSNS(context.Background(), client, &config.AWS{})

	topicArn := "arn:aws:sns:ap-northeast-1:123456789012:my-topic"
	awsSNS.mpResourceNameToArn.Add(context.Background(), "my-topic", &snsInfo{Arn: topicArn})

	client.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, input *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
			require.NotNil(input.TargetArn)
			require.Equal(topicArn, *input.TargetArn)
			require.NotNil(input.Message)
			require.Equal(`{"a":1}`, *input.Message)
			require.NotNil(input.MessageGroupId)
			require.Equal("A1", *input.MessageGroupId)
			return &sns.PublishOutput{MessageId: aws.String("msg-1")}, nil
		},
	)

	out := interfaces.Outbox{
		AggregateId:   "A1",
		AggregateType: topicArn,
		EventType:     "created",
		Payload:       `{"a":1}`,
		ProducerName:  "producer",
	}

	msgID, err := awsSNS.Publish(context.Background(), out)
	require.NoError(err)
	require.Equal("msg-1", fmt.Sprintf("%s", msgID))
}
