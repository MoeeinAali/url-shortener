package postgres

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// connectAttempts bounds how long Open waits for Postgres to accept connections,
// so services survive startup/restart ordering (depends_on health is only
// honored by `up`, not `restart`).
const connectAttempts = 30

// Open connects to Postgres using GORM, retrying until the database is reachable.
func Open(dsn string) (*gorm.DB, error) {
	cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}

	var lastErr error
	for attempt := 1; attempt <= connectAttempts; attempt++ {
		db, err := gorm.Open(postgres.Open(dsn), cfg)
		if err != nil {
			lastErr = err
			time.Sleep(time.Second)
			continue
		}
		sqlDB, err := db.DB()
		if err != nil {
			lastErr = err
			time.Sleep(time.Second)
			continue
		}
		if err := sqlDB.Ping(); err != nil {
			lastErr = err
			time.Sleep(time.Second)
			continue
		}
		return db, nil
	}
	return nil, fmt.Errorf("postgres unreachable after %d attempts: %w", connectAttempts, lastErr)
}

// AutoMigrate creates/updates all write-side tables.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&linkModel{},
		&outboxModel{},
		&analyticsModel{},
		&processedEventModel{},
	)
}
