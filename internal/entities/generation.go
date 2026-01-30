package entities

import (
	"time"

	"github.com/google/uuid"
)

type Generation struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index"`
	Prompt    *string
	Result    *string
	CreatedAt time.Time `gorm:"not null"`

	Project Project `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`
}

func (Generation) TableName() string {
	return "generations"
}
