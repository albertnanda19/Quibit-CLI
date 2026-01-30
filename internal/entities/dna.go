package entities

import (
	"time"

	"github.com/google/uuid"
)

type ProjectDNA struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index"`
	DNAHash   string    `gorm:"not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"not null"`

	Project Project `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`
}

func (ProjectDNA) TableName() string {
	return "project_dna"
}
