// Package event defines the transport representation of domain events as they
// travel out of the write side: serialized in the outbox and on the message bus.
// It depends on nothing else in the project, so adapters can share it freely
// without creating import cycles.
package event

import (
	"encoding/json"
	"time"
)

// Envelope is the wire format published to the message bus and consumed by the
// projector. It carries the metadata required for routing and idempotency.
type Envelope struct {
	EventID     string          `json:"event_id"`
	AggregateID string          `json:"aggregate_id"`
	EventType   string          `json:"event_type"`
	OccurredAt  time.Time       `json:"occurred_at"`
	Payload     json.RawMessage `json:"payload"`
}

// OutboxRecord is a row read from the transactional outbox by the relay.
type OutboxRecord struct {
	Seq         int64
	EventID     string
	AggregateID string
	EventType   string
	Payload     []byte
	OccurredAt  time.Time
}

// Payload is the per-event body stored as JSON. A single shape covers all event
// types; absent fields are simply empty.
type Payload struct {
	ShortCode string `json:"short_code"`
	LongURL   string `json:"long_url,omitempty"`
}
