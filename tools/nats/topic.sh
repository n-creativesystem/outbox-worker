#!/bin/bash

export NATS_URL='nats://nats:4222'
nats stream add TEST_EVENTS \
  --subjects "tests.>" \
  --storage file \
  --retention limits \
  --replicas 1 \
  --discard old \
  --max-age 7d \
  --dupe-window 2m \
  --user user \
  --password password \
  --defaults || echo "exists stream"

nats consumer add TEST_EVENTS TestProcessor \
  --pull \
  --filter "tests.>" \
  --ack explicit \
  --max-deliver 6 \
  --wait 2s \
  --user user \
  --password password \
  --defaults || echo "exists consumer"
