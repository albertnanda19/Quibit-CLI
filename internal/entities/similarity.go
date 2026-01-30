package entities

import (
	"time"

	"github.com/google/uuid"
)

type ProjectSimilarity struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID         uuid.UUID `gorm:"type:uuid;not null;index"`
	ComparedProjectID uuid.UUID `gorm:"type:uuid;not null;index"`
	SimilarityScore   float64   `gorm:"not null"`
	CreatedAt         time.Time `gorm:"not null"`

	Project         Project `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`
	ComparedProject Project `gorm:"foreignKey:ComparedProjectID;references:ID;constraint:OnDelete:CASCADE"`
}

func (ProjectSimilarity) TableName() string {
	return "project_similarity"
}
