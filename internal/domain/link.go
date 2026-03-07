package domain

import (
	"time"

	"github.com/google/uuid"
)

type Link struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ShortCode string    `gorm:"uniqueIndex;not null"`
	LongURL   string    `gorm:"not null"`
	Disabled  bool      `gorm:"default:false"`
	CreatedAt time.Time `gorm:"not null"`
}

func NewLink(short, long string) *Link {
	return &Link{
		ID:        uuid.New(),
		ShortCode: short,
		LongURL:   long,
		Disabled:  false,
		CreatedAt: time.Now(),
	}
}
