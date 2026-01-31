package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectOverview   string    `gorm:"type:text;not null"`
	MVPScopeJSON      string    `gorm:"type:jsonb;not null;column:mvp_scope"`
	TechStackJSON     string    `gorm:"type:jsonb;not null;column:tech_stack"`
	Complexity        string    `gorm:"not null"`
	EstimatedDuration string    `gorm:"not null"`
	DNAHash           string    `gorm:"not null;uniqueIndex"`
	AIProvider        string    `gorm:"not null"`
	ProviderUsed      string    `gorm:"type:text;not null;column:provider_used"`
	FallbackUsed      bool      `gorm:"not null;default:false;column:fallback_used"`
	ProviderError     *string   `gorm:"type:text;column:provider_error"`
	LatencyMS         int64     `gorm:"not null;default:0;column:latency_ms"`
	RetryReason       *string   `gorm:"type:text;column:retry_reason"`
	RawAIOutput       string    `gorm:"type:jsonb;not null"`
	AppType           string    `gorm:"not null"`
	Goal              string    `gorm:"not null"`
	CreatedAt         time.Time `gorm:"not null"`
}

func (Project) TableName() string {
	return "legacy_projects"
}
