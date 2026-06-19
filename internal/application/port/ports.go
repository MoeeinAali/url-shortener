// Package port declares the boundaries (ports) between the application core and
// the outside world. Infrastructure adapters implement these interfaces, and the
// application depends only on them — never on concrete adapters (dependency
// inversion / hexagonal architecture).
package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	"url-shortener/internal/domain/shared"
)

// Outbox is the write-side port for appending domain events to the transactional
// outbox outside an aggregate save (used for clicks recorded on the read path).
type Outbox interface {
	Append(ctx context.Context, events ...shared.DomainEvent) error
}

// ReadModel is the query-side port for the fast read store (Redis).
type ReadModel interface {
	// LongURL returns the destination for a short code, or link.ErrLinkNotFound.
	LongURL(ctx context.Context, shortCode string) (string, error)
	// Clicks returns the click counter for a short code (0 if none), or
	// link.ErrLinkNotFound if the link does not exist in the read model.
	Clicks(ctx context.Context, shortCode string) (int64, error)
}

// ReadModelWriter is the projector-side port for mutating the read store.
type ReadModelWriter interface {
	SaveLink(ctx context.Context, shortCode, longURL string) error
	DeleteLink(ctx context.Context, shortCode string) error
	SetClicks(ctx context.Context, shortCode string, count int64) error
}

// ClickCount is a durable analytics row.
type ClickCount struct {
	ShortCode string
	Count     int64
}

// AnalyticsStore is the durable source of truth for click counts (Postgres).
type AnalyticsStore interface {
	// RecordClick idempotently applies a click identified by eventID and returns
	// the resulting total count for the short code. Duplicate event ids are
	// counted at most once.
	RecordClick(ctx context.Context, eventID uuid.UUID, shortCode string, at time.Time) (int64, error)
	// ListAll returns every analytics row (used to rebuild the read model).
	ListAll(ctx context.Context) ([]ClickCount, error)
}

// LinkSnapshot is a flat read of an aggregate used for read-model rebuilds.
type LinkSnapshot struct {
	ShortCode string
	LongURL   string
	Disabled  bool
}

// LinkReader is a write-store read port used by the projector to rebuild the
// read model on startup.
type LinkReader interface {
	ListAll(ctx context.Context) ([]LinkSnapshot, error)
}
