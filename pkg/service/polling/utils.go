package polling

import (
	"github.com/n-creativesystem/outbox-worker/pkg/domain/outbox"
)

func eventToGroupingByAggregateType(events outbox.Events) map[string]outbox.Events {
	mp := map[string]outbox.Events{}
	for _, event := range events {
		mp[event.AggregateType()] = append(mp[event.AggregateType()], event)
	}
	return mp
}

func eventToGroupingAggregateId(events outbox.Events) map[string]outbox.Events {
	mp := map[string]outbox.Events{}
	for _, event := range events {
		mp[event.AggregateID()] = append(mp[event.AggregateID()], event)
	}
	return mp
}
