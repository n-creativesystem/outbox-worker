package outbox

import (
	"sort"
	"time"
)

type Event struct {
	id            int64
	aggregateType string
	aggregateID   string
	event         string
	producerName  string
	payload       string
	retryAt       *time.Time
	retryCount    int
	sentAt        *time.Time
}

func NewEvent(
	id int64,
	aggregateType string,
	aggregateID string,
	event string,
	producerName string,
	payload string,
	retryAt *time.Time,
	retryCount int,
	sentAt *time.Time,
) *Event {
	return &Event{
		id:            id,
		aggregateType: aggregateType,
		aggregateID:   aggregateID,
		event:         event,
		payload:       payload,
		retryAt:       retryAt,
		retryCount:    retryCount,
		producerName:  producerName,
		sentAt:        sentAt,
	}
}

func (e *Event) ID() int64             { return e.id }
func (e *Event) AggregateID() string   { return e.aggregateID }
func (e *Event) AggregateType() string { return e.aggregateType }
func (e *Event) Event() string         { return e.event }
func (e *Event) Payload() string       { return e.payload }
func (e *Event) ProducerName() string  { return e.producerName }
func (e *Event) RetryCount() int       { return e.retryCount }
func (e *Event) RetryAt() *time.Time   { return e.retryAt }
func (e *Event) RetryAtString() string {
	if e.retryAt != nil {
		return e.retryAt.Format(time.RFC3339)
	}
	return ""
}

func (e *Event) IncrementToRetryCount() {
	e.retryCount++
}

func (e *Event) CheckMaxRetryCount(retryCount int) bool {
	return e.retryCount >= retryCount
}

func (e *Event) CanNotRetry(value time.Time) bool {
	return e.retryAt != nil && e.retryAt.After(value)
}

func (e *Event) CanPublish() bool {
	return e.sentAt == nil
}

func (e *Event) SetSentAt(sentAt time.Time) {
	e.sentAt = &sentAt
}

func (e *Event) Retry(retryBackOff time.Duration, timeFn func() time.Time) {
	e.IncrementToRetryCount()
	duration := getNextRetryDuration(retryBackOff, e.retryCount)
	nextRetryTime := timeFn().Add(duration)
	e.retryAt = &nextRetryTime
}

type Events []*Event

func (es Events) SortByID() {
	sort.Slice(es, func(i, j int) bool {
		return es[i].id < es[j].id
	})
}
