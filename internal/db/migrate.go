package db

import (
	"context"
	"fmt"

	"quibit/internal/entities"

	"gorm.io/gorm"
)

func AutoMigrate(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("auto-migrate: db is nil")
	}

	if err := db.WithContext(ctx).AutoMigrate(
		&entities.Project{},
		&entities.Generation{},
		&entities.ProjectDNA{},
		&entities.ProjectSimilarity{},
	); err != nil {
		return fmt.Errorf("auto-migrate: %w", err)
	}

	return nil
}
