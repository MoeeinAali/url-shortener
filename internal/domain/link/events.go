package link

import (
	"time"

	"github.com/google/uuid"

	"url-shortener/internal/domain/shared"
)

// Stable event type names, also used as JetStream subjects.
const (
	EventTypeLinkCreated  = "link.created"
	EventTypeLinkDisabled = "link.disabled"
	EventTypeLinkClicked  = "link.clicked"
)

// baseEvent provides the common metadata every domain event carries.
type baseEvent struct {
	id          uuid.UUID
	aggregateID uuid.UUID
	occurredAt  time.Time
}

func newBaseEvent(aggregateID uuid.UUID) baseEvent {
	return baseEvent{
		id:          uuid.New(),
		aggregateID: aggregateID,
		occurredAt:  time.Now().UTC(),
	}
}

func (e baseEvent) EventID() uuid.UUID     { return e.id }
func (e baseEvent) AggregateID() uuid.UUID { return e.aggregateID }
func (e baseEvent) OccurredAt() time.Time  { return e.occurredAt }

// LinkCreated is raised when a new link is created.
type LinkCreated struct {
	baseEvent
	ShortCode string
	LongURL   string
}

// EventType implements shared.DomainEvent.
func (LinkCreated) EventType() string { return EventTypeLinkCreated }

// LinkDisabled is raised when a link is disabled.
type LinkDisabled struct {
	baseEvent
	ShortCode string
}

// EventType implements shared.DomainEvent.
func (LinkDisabled) EventType() string { return EventTypeLinkDisabled }

// LinkClicked is raised on every redirect. It is produced on the read path (the
// redirect endpoint) and recorded durably via the outbox so analytics never
// silently drops a click.
type LinkClicked struct {
	baseEvent
	ShortCode string
}

// EventType implements shared.DomainEvent.
func (LinkClicked) EventType() string { return EventTypeLinkClicked }

// NewLinkClicked builds a click event for a known short code. Unlike LinkCreated
// and LinkDisabled (which are raised inside the aggregate), clicks are recorded
// without loading the aggregate, so this constructor is exported.
func NewLinkClicked(code ShortCode) LinkClicked {
	// The aggregate id is not known on the hot redirect path, so a fresh id is
	// used as the event's aggregate reference; the short code is the projection key.
	return LinkClicked{
		baseEvent: newBaseEvent(uuid.New()),
		ShortCode: code.String(),
	}
}

// Compile-time guarantees that events satisfy the DomainEvent contract.
var (
	_ shared.DomainEvent = LinkCreated{}
	_ shared.DomainEvent = LinkDisabled{}
	_ shared.DomainEvent = LinkClicked{}
)
