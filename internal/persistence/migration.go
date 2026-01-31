package persistence

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"quibit/internal/persistence/models"
)

func AutoMigrate(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("auto-migrate: db is nil")
	}

	if err := db.WithContext(ctx).AutoMigrate(
		&models.Project{},
		&models.ProjectFeature{},
		&models.ProjectMeta{},
	); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}

	return nil
}
