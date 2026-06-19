// Package shared holds domain primitives that are not tied to a single aggregate.
package shared

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent is the contract every domain event satisfies. Events are the
// ubiquitous-language record of "something that happened" in the domain. They
// are raised by aggregates and later carried out of the write side through the
// transactional outbox.
type DomainEvent interface {
	// EventID is a globally unique id used for at-least-once de-duplication.
	EventID() uuid.UUID
	// EventType is the stable name used as the messaging subject (e.g. "link.created").
	EventType() string
	// AggregateID is the id of the aggregate that produced the event.
	AggregateID() uuid.UUID
	// OccurredAt is the wall-clock time the event happened.
	OccurredAt() time.Time
}
