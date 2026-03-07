package repository

import (
	"context"
	"url-shortener/internal/domain"

	"gorm.io/gorm"
)

type LinkRepository struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) *LinkRepository {
	return &LinkRepository{db: db}
}

func (r *LinkRepository) Create(ctx context.Context, link *domain.Link) error {
	return r.db.WithContext(ctx).Create(link).Error
}

func (r *LinkRepository) Disable(ctx context.Context, short string) error {
	return r.db.
		WithContext(ctx).
		Model(&domain.Link{}).
		Where("short_code = ?", short).
		Update("disabled", true).
		Error
}

func (r *LinkRepository) ListAll(ctx context.Context) ([]domain.Link, error) {
	var links []domain.Link
	err := r.db.WithContext(ctx).Find(&links).Error
	return links, err
}
