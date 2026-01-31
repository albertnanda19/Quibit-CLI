package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"quibit/internal/domain"
	"quibit/internal/persistence/models"
)

var ErrDuplicateDNAHash = errors.New("duplicate dna hash")

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) (*ProjectRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("project repository: db is nil")
	}
	return &ProjectRepository{db: db}, nil
}

type SaveParams struct {
	Project          domain.Project
	DNAHash          string
	AIProvider       string
	RawAIJSON        string
	SimilarityScore  float64
	SimilarProjectID *uuid.UUID
	PivotReason      *string
}

func (r *ProjectRepository) Save(ctx context.Context, p SaveParams) (uuid.UUID, error) {
	if ctx == nil {
		return uuid.Nil, fmt.Errorf("save project: ctx is nil")
	}
	if r == nil || r.db == nil {
		return uuid.Nil, fmt.Errorf("save project: repository is not initialized")
	}
	if p.DNAHash == "" {
		return uuid.Nil, fmt.Errorf("save project: dna hash is required")
	}
	if p.AIProvider == "" {
		return uuid.Nil, fmt.Errorf("save project: ai provider is required")
	}

	projectID := uuid.New()
	now := time.Now()

	targetUsersJSON, err := json.Marshal(p.Project.TargetUsers)
	if err != nil {
		return uuid.Nil, fmt.Errorf("save project: marshal target users: %w", err)
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return uuid.Nil, fmt.Errorf("save project: begin transaction: %w", tx.Error)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	row := models.Project{
		ID:               projectID,
		Title:            p.Project.Title,
		Summary:          p.Project.Summary,
		DNAHash:          p.DNAHash,
		SimilarityScore:  p.SimilarityScore,
		SimilarProjectID: p.SimilarProjectID,
		PivotReason:      p.PivotReason,
		Complexity:       p.Project.EstimatedComplexity,
		Duration:         p.Project.EstimatedDuration,
		AIProvider:       p.AIProvider,
		CreatedAt:        now,
	}
	if err := tx.Create(&row).Error; err != nil {
		if isUniqueViolation(err) {
			return uuid.Nil, ErrDuplicateDNAHash
		}
		return uuid.Nil, fmt.Errorf("save project: insert projects: %w", err)
	}

	var features []models.ProjectFeature
	for _, v := range p.Project.CoreFeatures {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		features = append(features, models.ProjectFeature{ID: uuid.New(), ProjectID: projectID, Type: "core", Description: v})
	}
	for _, v := range p.Project.MVPScope {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		features = append(features, models.ProjectFeature{ID: uuid.New(), ProjectID: projectID, Type: "mvp", Description: v})
	}
	for _, v := range p.Project.OptionalExtensions {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		features = append(features, models.ProjectFeature{ID: uuid.New(), ProjectID: projectID, Type: "extension", Description: v})
	}
	if len(features) > 0 {
		if err := tx.Create(&features).Error; err != nil {
			return uuid.Nil, fmt.Errorf("save project: insert project_features: %w", err)
		}
	}

	meta := models.ProjectMeta{
		ProjectID:   projectID,
		TargetUsers: string(targetUsersJSON),
		TechStack:   p.Project.RecommendedStack,
		RawAIOutput: p.RawAIJSON,
	}
	if err := tx.Create(&meta).Error; err != nil {
		return uuid.Nil, fmt.Errorf("save project: insert project_meta: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return uuid.Nil, fmt.Errorf("save project: commit: %w", err)
	}

	return projectID, nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "sqlstate 23505") {
		return true
	}
	if strings.Contains(s, "duplicate key") {
		return true
	}
	if strings.Contains(s, "unique constraint") {
		return true
	}
	return false
}

type SimilarityCandidate struct {
	ID      uuid.UUID
	Project domain.Project
}

func (r *ProjectRepository) ListRecentForSimilarity(ctx context.Context, limit int) ([]SimilarityCandidate, error) {
	if ctx == nil {
		return nil, fmt.Errorf("list recent: ctx is nil")
	}
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("list recent: repository is not initialized")
	}
	if limit <= 0 {
		limit = 50
	}
	if v := os.Getenv("SIMILARITY_LOOKBACK_N"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			limit = n
		}
	}

	var rows []struct {
		ID        uuid.UUID
		RawAIJSON string
	}
	q := r.db.WithContext(ctx).
		Table("projects").
		Select("projects.id as id, project_meta.raw_ai_output as raw_ai_json").
		Joins("join project_meta on project_meta.project_id = projects.id").
		Order("projects.created_at desc").
		Limit(limit)

	if err := q.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("list recent: %w", err)
	}

	out := make([]SimilarityCandidate, 0, len(rows))
	for _, row := range rows {
		var p domain.Project
		if err := json.Unmarshal([]byte(row.RawAIJSON), &p); err != nil {
			return nil, fmt.Errorf("list recent: parse raw_ai_output: %w", err)
		}
		out = append(out, SimilarityCandidate{ID: row.ID, Project: p})
	}

	return out, nil
}
