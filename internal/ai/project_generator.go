package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"quibit/internal/domain"
	"quibit/internal/model"
)

type ProjectGenerator struct {
	g *Generator
}

func NewProjectGenerator(client *genai.Client) (*ProjectGenerator, error) {
	g, err := NewGenerator(client)
	if err != nil {
		return nil, fmt.Errorf("project generator: %w", err)
	}

	return &ProjectGenerator{g: g}, nil
}

func buildProjectProposalPrompt(in model.ProjectInput, pivot string) string {
	techStackJSON, err := json.Marshal(in.TechStack)
	if err != nil {
		techStackJSON = []byte("[]")
	}
	projectKind := strings.TrimSpace(in.ProjectKind)
	projectKindLine := ""
	if projectKind != "" {
		projectKindLine = "- project_kind: " + projectKind + "\n"
	}
	projectKindRule := ""
	if projectKind == "" {
		projectKindRule = "- If project_kind is not provided, you MUST infer a suitable software category based on tech_stack and typical real-world use.\n"
	}
	dbPref := strings.TrimSpace(in.Database)
	dbLine := ""
	if dbPref != "" && strings.ToLower(dbPref) != "none" {
		dbLine = "- database_preference: " + dbPref + "\n"
	}
	if strings.ToLower(dbPref) == "none" {
		dbLine = "- database_preference: none\n"
	}

	pivot = strings.TrimSpace(pivot)
	pivotBlock := ""
	if pivot != "" {
		pivotBlock = "Additional Constraints:\n" + pivot + "\n\n"
	}

	return "Return ONLY valid JSON. Do not include explanation, formatting, markdown, or extra text.\n" +
		"You MUST return exactly one JSON object and nothing else.\n\n" +
		"User Input (use these as constraints):\n" +
		"- app_type: " + in.AppType + "\n" +
		projectKindLine +
		dbLine +
		"- complexity: " + in.Complexity + "\n" +
		"- tech_stack: " + string(techStackJSON) + "\n" +
		"- goal: " + in.Goal + "\n" +
		"- estimated_duration: " + in.Timeframe + "\n\n" +
		"Rules:\n" +
		projectKindRule +
		"- estimated_complexity must match complexity exactly as provided (beginner|intermediate|advanced).\n" +
		"- recommended_stack must reflect tech_stack (do not invent technologies outside tech_stack).\n" +
		"- estimated_duration must match estimated_duration exactly as provided.\n" +
		"- Fill EVERY field in the schema.\n" +
		"- Do NOT add, remove, or rename any fields.\n\n" +
		pivotBlock +
		"Schema (must include ALL fields):\n" +
		"{\n" +
		"  \"title\": string,\n" +
		"  \"summary\": string,\n" +
		"  \"problem_statement\": string,\n" +
		"  \"target_users\": string[],\n" +
		"  \"core_features\": string[],\n" +
		"  \"mvp_scope\": string[],\n" +
		"  \"optional_extensions\": string[],\n" +
		"  \"recommended_stack\": string,\n" +
		"  \"estimated_complexity\": \"beginner\" | \"intermediate\" | \"advanced\",\n" +
		"  \"estimated_duration\": string\n" +
		"}\n"
}

type GeneratedProject struct {
	Project domain.Project
	RawJSON string
}

func (pg *ProjectGenerator) Generate(ctx context.Context, in model.ProjectInput) (GeneratedProject, error) {
	return pg.generate(ctx, in, "")
}

func (pg *ProjectGenerator) GenerateWithPivot(ctx context.Context, in model.ProjectInput, pivot string) (GeneratedProject, error) {
	return pg.generate(ctx, in, pivot)
}

func (pg *ProjectGenerator) generate(ctx context.Context, in model.ProjectInput, pivot string) (GeneratedProject, error) {
	if pg == nil || pg.g == nil {
		return GeneratedProject{}, fmt.Errorf("project generator: not initialized")
	}

	raw, err := pg.g.GenerateText(ctx, buildProjectProposalPrompt(in, pivot))
	if err != nil {
		return GeneratedProject{}, fmt.Errorf("generate project proposal: %w", err)
	}

	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()

	var payload json.RawMessage
	if err := dec.Decode(&payload); err != nil {
		return GeneratedProject{}, fmt.Errorf("generate project proposal: invalid JSON: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err == nil {
		return GeneratedProject{}, fmt.Errorf("generate project proposal: invalid JSON: trailing content")
	}

	dec2 := json.NewDecoder(bytes.NewReader(payload))
	dec2.DisallowUnknownFields()

	var p domain.Project
	if err := dec2.Decode(&p); err != nil {
		return GeneratedProject{}, fmt.Errorf("generate project proposal: invalid JSON: %w", err)
	}

	if strings.TrimSpace(p.Title) == "" {
		return GeneratedProject{}, fmt.Errorf("generate project proposal: invalid JSON: title is required")
	}
	if strings.TrimSpace(p.ProblemStatement) == "" {
		return GeneratedProject{}, fmt.Errorf("generate project proposal: invalid JSON: problem_statement is required")
	}
	if strings.TrimSpace(p.EstimatedComplexity) != in.Complexity {
		return GeneratedProject{}, fmt.Errorf("generate project proposal: invalid JSON: estimated_complexity must match input")
	}
	if strings.TrimSpace(p.EstimatedDuration) != in.Timeframe {
		return GeneratedProject{}, fmt.Errorf("generate project proposal: invalid JSON: estimated_duration must match input")
	}

	return GeneratedProject{Project: p, RawJSON: string(payload)}, nil
}
