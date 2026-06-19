package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"url-shortener/internal/application/port"
	"url-shortener/internal/domain/shared"
	"url-shortener/internal/infrastructure/event"
)

// OutboxStore implements port.Outbox for standalone event appends (clicks) and
// exposes the relay's read/mark operations over the outbox table.
type OutboxStore struct {
	db *gorm.DB
}

// NewOutboxStore constructs the store.
func NewOutboxStore(db *gorm.DB) *OutboxStore {
	return &OutboxStore{db: db}
}

// Append writes one or more domain events to the outbox in a transaction.
func (o *OutboxStore) Append(ctx context.Context, events ...shared.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}
	return o.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, e := range events {
			row, err := toOutboxRow(e)
			if err != nil {
				return err
			}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FetchUnpublished returns the next batch of unpublished outbox rows in order.
func (o *OutboxStore) FetchUnpublished(ctx context.Context, limit int) ([]event.OutboxRecord, error) {
	var rows []outboxModel
	err := o.db.WithContext(ctx).
		Where("published_at IS NULL").
		Order("seq asc").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	records := make([]event.OutboxRecord, 0, len(rows))
	for _, r := range rows {
		records = append(records, event.OutboxRecord{
			Seq:         r.Seq,
			EventID:     r.ID.String(),
			AggregateID: r.AggregateID.String(),
			EventType:   r.EventType,
			Payload:     r.Payload,
			OccurredAt:  r.OccurredAt,
		})
	}
	return records, nil
}

// MarkPublished stamps the given outbox rows as published.
func (o *OutboxStore) MarkPublished(ctx context.Context, seqs []int64) error {
	if len(seqs) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return o.db.WithContext(ctx).
		Model(&outboxModel{}).
		Where("seq IN ?", seqs).
		Update("published_at", now).Error
}

var _ port.Outbox = (*OutboxStore)(nil)
