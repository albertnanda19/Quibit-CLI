package entities

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Title       string    `gorm:"not null"`
	Description *string
	Complexity  *string
	CreatedAt   time.Time `gorm:"not null"`
}

func (Project) TableName() string {
	return "projects"
}
