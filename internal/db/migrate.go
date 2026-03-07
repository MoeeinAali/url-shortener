package db

import (
	"url-shortener/internal/domain"

	"gorm.io/gorm"
)

func RunMigrations(db *gorm.DB) error {

	return db.AutoMigrate(
		&domain.Link{},
	)
}
