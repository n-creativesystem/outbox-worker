package schema

import (
	"database/sql"
	"time"
)

type Outbox struct {
	// ID of the primary key.
	ID int64 `db:"id" fieldtag:"pk"`
	// AggregateType holds the value of the "aggregate_type" field.
	AggregateType string `db:"aggregate_type"`
	// AggregateID holds the value of the "aggregate_id" field.
	AggregateID string `db:"aggregate_id"`
	// Event holds the value of the "event" field.
	Event string `db:"event"`
	// Payload holds the value of the "payload" field.
	Payload []byte `db:"payload"`
	// RetryAt holds the value of the "retry_at" field.
	RetryAt sql.NullTime `db:"retry_at"`
	// RetryCount holds the value of the "retry_count" field.
	RetryCount sql.NullInt32 `db:"retry_count"`
	// SentAt holds the value of the "sent_at" field.
	SentAt sql.NullTime `db:"sent_at"`
}

func (v *Outbox) RetryAtValue() *time.Time {
	if v.RetryAt.Valid {
		return &v.RetryAt.Time
	}
	return nil
}

func (v *Outbox) RetryCountValue() int {
	if v.RetryCount.Valid {
		return int(v.RetryCount.Int32)
	}
	return 0
}

func (v *Outbox) SentAtValue() *time.Time {
	if v.SentAt.Valid {
		return &v.SentAt.Time
	}
	return nil
}
