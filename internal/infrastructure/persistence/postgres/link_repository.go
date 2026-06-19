package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"url-shortener/internal/application/port"
	"url-shortener/internal/domain/link"
)

// LinkRepository implements link.Repository and port.LinkReader.
type LinkRepository struct {
	db *gorm.DB
}

// NewLinkRepository constructs the repository.
func NewLinkRepository(db *gorm.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

// Save persists the aggregate state and drains its pending domain events into the
// outbox within a single transaction — the heart of the transactional outbox.
func (r *LinkRepository) Save(ctx context.Context, l *link.Link) error {
	model := linkModel{
		ID:        l.ID().UUID(),
		ShortCode: l.ShortCode().String(),
		LongURL:   l.LongURL().String(),
		Status:    string(l.Status()),
		Version:   l.Version(),
		CreatedAt: l.CreatedAt(),
	}
	events := l.PullEvents()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Upsert by primary key: insert on create, update on disable.
		if err := tx.Save(&model).Error; err != nil {
			return err
		}
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

// FindByShortCode loads and reconstitutes an aggregate.
func (r *LinkRepository) FindByShortCode(ctx context.Context, code link.ShortCode) (*link.Link, error) {
	var m linkModel
	err := r.db.WithContext(ctx).Where("short_code = ?", code.String()).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, link.ErrLinkNotFound
	}
	if err != nil {
		return nil, err
	}
	return toAggregate(m)
}

// ExistsByShortCode reports whether a short code is already taken.
func (r *LinkRepository) ExistsByShortCode(ctx context.Context, code link.ShortCode) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&linkModel{}).Where("short_code = ?", code.String()).Count(&count).Error
	return count > 0, err
}

// ListAll returns flat snapshots of every link for read-model rebuilds.
func (r *LinkRepository) ListAll(ctx context.Context) ([]port.LinkSnapshot, error) {
	var models []linkModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	snapshots := make([]port.LinkSnapshot, 0, len(models))
	for _, m := range models {
		snapshots = append(snapshots, port.LinkSnapshot{
			ShortCode: m.ShortCode,
			LongURL:   m.LongURL,
			Disabled:  link.LinkStatus(m.Status).IsDisabled(),
		})
	}
	return snapshots, nil
}

func toAggregate(m linkModel) (*link.Link, error) {
	id := link.LinkIDFromUUID(m.ID)
	code, err := link.NewShortCode(m.ShortCode)
	if err != nil {
		return nil, err
	}
	longURL, err := link.NewURL(m.LongURL)
	if err != nil {
		return nil, err
	}
	return link.Reconstitute(id, code, longURL, link.LinkStatus(m.Status), m.CreatedAt, m.Version), nil
}

// Compile-time interface checks.
var (
	_ link.Repository = (*LinkRepository)(nil)
	_ port.LinkReader = (*LinkRepository)(nil)
)
