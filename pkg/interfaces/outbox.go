package interfaces

type Outbox struct {
	AggregateId   string `json:"aggregate_id"`
	AggregateType string `json:"aggregate_type"`
	EventType     string `json:"event_type"`
	Payload       string `json:"payload"`
	ProducerName  string `json:"producer"`
}
