package models

import (
	"time"

	"github.com/google/uuid"
)

type Project struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Title            string     `gorm:"not null"`
	Summary          string     `gorm:"not null"`
	DNAHash          string     `gorm:"not null;uniqueIndex"`
	SimilarityScore  float64    `gorm:"not null;default:0"`
	SimilarProjectID *uuid.UUID `gorm:"type:uuid"`
	PivotReason      *string    `gorm:"type:text"`

	ProjectOverview string `gorm:"type:text;not null;column:project_overview"`
	MVPScopeJSON    string `gorm:"type:jsonb;not null;column:mvp_scope"`
	TechStackJSON   string `gorm:"type:jsonb;not null;column:tech_stack"`
	RawAIOutput     string `gorm:"type:jsonb;not null;column:raw_ai_output"`
	AppType         string `gorm:"not null;column:app_type"`
	Goal            string `gorm:"not null;column:goal"`

	Complexity string `gorm:"not null"`
	Duration   string `gorm:"not null"`

	AIProvider    string  `gorm:"not null"`
	ProviderUsed  string  `gorm:"type:text;not null;column:provider_used"`
	FallbackUsed  bool    `gorm:"not null;default:false;column:fallback_used"`
	ProviderError *string `gorm:"type:text;column:provider_error"`
	LatencyMS     int64   `gorm:"not null;default:0;column:latency_ms"`
	RetryReason   *string `gorm:"type:text;column:retry_reason"`

	CreatedAt time.Time `gorm:"not null"`
}

func (Project) TableName() string {
	return "projects"
}

type ProjectFeature struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProjectID   uuid.UUID `gorm:"type:uuid;not null;index"`
	Type        string    `gorm:"not null"`
	Description string    `gorm:"not null"`
}

func (ProjectFeature) TableName() string {
	return "project_features"
}

type ProjectMeta struct {
	ProjectID   uuid.UUID `gorm:"type:uuid;primaryKey;column:project_id"`
	TargetUsers string    `gorm:"type:jsonb;not null;column:target_users"`
	TechStack   string    `gorm:"type:text;not null;column:tech_stack"`
	RawAIOutput string    `gorm:"type:jsonb;not null;column:raw_ai_output"`
}

func (ProjectMeta) TableName() string {
	return "project_meta"
}
