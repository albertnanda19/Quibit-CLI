package db

import (
	"context"
	"fmt"

	dbmodels "quibit/internal/db/models"

	"gorm.io/gorm"
)

func AutoMigrateLegacy(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("auto-migrate legacy: db is nil")
	}
	if err := db.WithContext(ctx).AutoMigrate(
		&dbmodels.Project{},
	); err != nil {
		return fmt.Errorf("auto-migrate legacy: %w", err)
	}
	return nil
}
