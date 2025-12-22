#!/bin/bash

awslocal sns create-topic --name test-sns
awslocal sqs create-queue --queue-name test-sns

awslocal sns subscribe \
  --topic-arn arn:aws:sns:ap-northeast-1:000000000000:test-sns \
  --protocol sqs \
  --notification-endpoint arn:aws:sqs:ap-northeast-1:000000000000:test-sns

awslocal sns create-topic --name test-sns.fifo --attributes FifoTopic=true,ContentBasedDeduplication=false
awslocal sqs create-queue --queue-name test-sns-dlq.fifo --attributes FifoQueue=true
awslocal sqs create-queue --queue-name test-sns.fifo --attributes \
  '{"FifoQueue": "true","ContentBasedDeduplication": "true","RedrivePolicy": "{\"deadLetterTargetArn\": \"arn:aws:sqs:ap-northeast-1:000000000000:test-sns-dlq.fifo\", \"maxReceiveCount\": 2}"}'

awslocal sns subscribe \
  --topic-arn arn:aws:sns:ap-northeast-1:000000000000:test-sns.fifo \
  --protocol sqs \
  --notification-endpoint arn:aws:sqs:ap-northeast-1:000000000000:test-sns.fifo


# awslocal sqs get-queue-attributes --queue-url http://localstack/000000000000/test-sns --attribute-names All
