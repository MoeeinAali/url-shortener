// Package postgres is the write-side adapter. It implements the domain
// Repository port and the application's outbox/analytics ports on top of GORM.
// All GORM/persistence concerns are confined here; the domain stays persistence
// ignorant.
package postgres

import (
	"time"

	"github.com/google/uuid"
)

// linkModel is the persistence representation of the Link aggregate.
type linkModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ShortCode string    `gorm:"uniqueIndex;not null"`
	LongURL   string    `gorm:"not null"`
	Status    string    `gorm:"not null"`
	Version   int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"not null"`
}

func (linkModel) TableName() string { return "links" }

// outboxModel is one row of the transactional outbox. Seq gives a monotonic
// publish order; ID is the unique domain event id used for bus de-duplication.
type outboxModel struct {
	Seq         int64      `gorm:"primaryKey;autoIncrement"`
	ID          uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null"`
	AggregateID uuid.UUID  `gorm:"type:uuid;not null"`
	EventType   string     `gorm:"not null;index"`
	Payload     []byte     `gorm:"type:jsonb;not null"`
	OccurredAt  time.Time  `gorm:"not null"`
	PublishedAt *time.Time `gorm:"index"`
}

func (outboxModel) TableName() string { return "outbox_events" }

// analyticsModel is the durable source of truth for click counts.
type analyticsModel struct {
	ShortCode     string `gorm:"primaryKey"`
	ClickCount    int64  `gorm:"not null;default:0"`
	LastClickedAt *time.Time
}

func (analyticsModel) TableName() string { return "link_analytics" }

// processedEventModel records consumed event ids so the projector is idempotent.
type processedEventModel struct {
	EventID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProcessedAt time.Time `gorm:"not null"`
}

func (processedEventModel) TableName() string { return "processed_events" }
