// Package relay implements the publishing half of the Transactional Outbox
// pattern: it polls the outbox table and forwards events to the message bus,
// solving the dual-write problem (DB commit and publish can never diverge).
package relay

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"url-shortener/internal/infrastructure/event"
)

// Source reads unpublished outbox rows and marks them published.
type Source interface {
	FetchUnpublished(ctx context.Context, limit int) ([]event.OutboxRecord, error)
	MarkPublished(ctx context.Context, seqs []int64) error
}

// Publisher publishes an envelope to the bus with a dedup id.
type Publisher interface {
	Publish(subject string, data []byte, msgID string) error
}

// Relay forwards outbox rows to the bus.
type Relay struct {
	source    Source
	publisher Publisher
	batchSize int
	log       *zap.Logger
}

// New constructs a relay.
func New(source Source, publisher Publisher, batchSize int, log *zap.Logger) *Relay {
	return &Relay{source: source, publisher: publisher, batchSize: batchSize, log: log}
}

// RunOnce publishes a single batch and returns how many rows were forwarded.
// A row is marked published only after a successful publish, so a crash mid-batch
// results in at-least-once redelivery (de-duplicated downstream), never loss.
func (r *Relay) RunOnce(ctx context.Context) (int, error) {
	records, err := r.source.FetchUnpublished(ctx, r.batchSize)
	if err != nil {
		return 0, err
	}
	if len(records) == 0 {
		return 0, nil
	}

	published := make([]int64, 0, len(records))
	for _, rec := range records {
		envelope := event.Envelope{
			EventID:     rec.EventID,
			AggregateID: rec.AggregateID,
			EventType:   rec.EventType,
			OccurredAt:  rec.OccurredAt,
			Payload:     json.RawMessage(rec.Payload),
		}
		data, err := json.Marshal(envelope)
		if err != nil {
			r.log.Error("relay: marshal envelope failed", zap.Error(err), zap.String("event_id", rec.EventID))
			continue
		}
		if err := r.publisher.Publish(rec.EventType, data, rec.EventID); err != nil {
			r.log.Warn("relay: publish failed, will retry", zap.Error(err), zap.String("event_id", rec.EventID))
			break // stop the batch to preserve ordering; retry next tick
		}
		published = append(published, rec.Seq)
	}

	if len(published) > 0 {
		if err := r.source.MarkPublished(ctx, published); err != nil {
			return len(published), err
		}
		r.log.Info("relay: published outbox batch", zap.Int("count", len(published)))
	}
	return len(published), nil
}

// Run polls the outbox on an interval until the context is cancelled.
func (r *Relay) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := r.RunOnce(ctx); err != nil {
				r.log.Error("relay: run once failed", zap.Error(err))
			}
		}
	}
}
