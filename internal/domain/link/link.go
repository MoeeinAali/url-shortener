package link

import (
	"time"

	"url-shortener/internal/domain/shared"
)

// Link is the aggregate root of the URL-shortening bounded context. It owns its
// invariants and is the only place link state may change. All fields are private
// so the only way to mutate a Link is through its behavior.
type Link struct {
	id        LinkID
	shortCode ShortCode
	longURL   URL
	status    LinkStatus
	createdAt time.Time
	version   int

	events []shared.DomainEvent
}

// NewLink creates a brand-new, active link and records a LinkCreated event.
func NewLink(code ShortCode, longURL URL) (*Link, error) {
	if code.IsZero() {
		return nil, ErrInvalidShortCode
	}
	if longURL.IsZero() {
		return nil, ErrInvalidURL
	}

	id := NewLinkID()
	l := &Link{
		id:        id,
		shortCode: code,
		longURL:   longURL,
		status:    StatusActive,
		createdAt: time.Now().UTC(),
		version:   1,
	}
	l.record(LinkCreated{
		baseEvent: newBaseEvent(id.UUID()),
		ShortCode: code.String(),
		LongURL:   longURL.String(),
	})
	return l, nil
}

// Reconstitute rebuilds an aggregate from persisted state without raising any
// events. It is used by repositories when loading from the write store.
func Reconstitute(id LinkID, code ShortCode, longURL URL, status LinkStatus, createdAt time.Time, version int) *Link {
	return &Link{
		id:        id,
		shortCode: code,
		longURL:   longURL,
		status:    status,
		createdAt: createdAt,
		version:   version,
	}
}

// Disable transitions the link to the disabled state. It enforces the invariant
// that a disabled link cannot be disabled again.
func (l *Link) Disable() error {
	if l.status.IsDisabled() {
		return ErrLinkAlreadyDisabled
	}
	l.status = StatusDisabled
	l.version++
	l.record(LinkDisabled{
		baseEvent: newBaseEvent(l.id.UUID()),
		ShortCode: l.shortCode.String(),
	})
	return nil
}

// PullEvents returns and clears the aggregate's pending domain events. The
// repository drains these into the outbox in the same transaction as the state.
func (l *Link) PullEvents() []shared.DomainEvent {
	events := l.events
	l.events = nil
	return events
}

func (l *Link) record(e shared.DomainEvent) { l.events = append(l.events, e) }

// Accessors (read-only projections of internal state).
func (l *Link) ID() LinkID            { return l.id }
func (l *Link) ShortCode() ShortCode  { return l.shortCode }
func (l *Link) LongURL() URL          { return l.longURL }
func (l *Link) Status() LinkStatus    { return l.status }
func (l *Link) CreatedAt() time.Time  { return l.createdAt }
func (l *Link) Version() int          { return l.version }
