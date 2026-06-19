package link

import "context"

// Repository is the port (interface) for persisting Link aggregates. The
// implementation lives in the infrastructure layer (dependency inversion).
//
// Save MUST persist the aggregate state and drain its pending domain events into
// the transactional outbox atomically (single transaction). This is what makes
// the write side and the event log impossible to diverge.
type Repository interface {
	Save(ctx context.Context, l *Link) error
	FindByShortCode(ctx context.Context, code ShortCode) (*Link, error)
	ExistsByShortCode(ctx context.Context, code ShortCode) (bool, error)
}
