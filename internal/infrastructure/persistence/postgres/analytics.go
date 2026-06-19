package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"url-shortener/internal/application/port"
)

// AnalyticsStore implements port.AnalyticsStore: the durable source of truth for
// click counts, with idempotent application via the processed_events table.
type AnalyticsStore struct {
	db *gorm.DB
}

// NewAnalyticsStore constructs the store.
func NewAnalyticsStore(db *gorm.DB) *AnalyticsStore {
	return &AnalyticsStore{db: db}
}

// RecordClick applies a click exactly once per eventID and returns the resulting
// total. The dedup insert and the counter increment happen in one transaction so
// a retry can never double-count.
func (s *AnalyticsStore) RecordClick(ctx context.Context, eventID uuid.UUID, shortCode string, at time.Time) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Idempotency gate: only the first writer of this eventID inserts a row.
		res := tx.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&processedEventModel{EventID: eventID, ProcessedAt: at})
		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected > 0 {
			// New event → increment the durable counter (upsert).
			if err := tx.Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "short_code"}},
				DoUpdates: clause.Assignments(map[string]interface{}{
					"click_count":     gorm.Expr("link_analytics.click_count + 1"),
					"last_clicked_at": at,
				}),
			}).Create(&analyticsModel{ShortCode: shortCode, ClickCount: 1, LastClickedAt: &at}).Error; err != nil {
				return err
			}
		}

		// Read back the authoritative total (also covers the duplicate path so
		// the read model can be self-healed to the correct value).
		var m analyticsModel
		err := tx.Where("short_code = ?", shortCode).First(&m).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			count = 0
			return nil
		}
		if err != nil {
			return err
		}
		count = m.ClickCount
		return nil
	})
	return count, err
}

// ListAll returns every analytics row for read-model rebuilds.
func (s *AnalyticsStore) ListAll(ctx context.Context) ([]port.ClickCount, error) {
	var models []analyticsModel
	if err := s.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	counts := make([]port.ClickCount, 0, len(models))
	for _, m := range models {
		counts = append(counts, port.ClickCount{ShortCode: m.ShortCode, Count: m.ClickCount})
	}
	return counts, nil
}

var _ port.AnalyticsStore = (*AnalyticsStore)(nil)
