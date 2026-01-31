package models

import (
	"time"

	"github.com/google/uuid"
)

type ProjectEvolution struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID uuid.UUID `gorm:"type:uuid;not null;index;column:project_id"`

	RawAIOutput string `gorm:"type:jsonb;not null;column:raw_ai_output"`

	ProviderUsed  string  `gorm:"type:text;not null;column:provider_used"`
	FallbackUsed  bool    `gorm:"not null;default:false;column:fallback_used"`
	ProviderError *string `gorm:"type:text;column:provider_error"`
	LatencyMS     int64   `gorm:"not null;default:0;column:latency_ms"`

	CreatedAt time.Time `gorm:"not null"`
}

func (ProjectEvolution) TableName() string {
	return "project_evolutions"
}
