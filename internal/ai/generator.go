package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"quibit/internal/model"

	"google.golang.org/genai"
)

var modelCandidates = []string{
	"gemini-3-flash-preview",
	"gemini-2.5-flash",
}

func isValidComplexity(v string) bool {
	switch v {
	case "beginner", "intermediate", "advanced":
		return true
	default:
		return false
	}
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if normalizeToken(a[i]) != normalizeToken(b[i]) {
			return false
		}
	}
	return true
}

func validateRequiredObjectKeys(obj map[string]json.RawMessage, required []string) error {
	if len(obj) != len(required) {
		return fmt.Errorf("invalid JSON: wrong number of fields")
	}
	for _, k := range required {
		if _, ok := obj[k]; !ok {
			return fmt.Errorf("invalid JSON: missing %s", k)
		}
	}
	return nil
}

type Generator struct {
	client *genai.Client
}

func NewGenerator(client *genai.Client) (*Generator, error) {
	if client == nil {
		return nil, fmt.Errorf("ai generator: client is nil")
	}

	return &Generator{client: client}, nil
}

func (g *Generator) GenerateText(ctx context.Context, prompt string) (string, error) {
	if g == nil || g.client == nil {
		return "", fmt.Errorf("generate text: client is nil")
	}

	for i, model := range modelCandidates {
		resp, err := g.client.Models.GenerateContent(ctx, model, []*genai.Content{{
			Role: genai.RoleUser,
			Parts: []*genai.Part{{
				Text: prompt,
			}},
		}}, nil)
		if err != nil {
			if i < len(modelCandidates)-1 && isOverloadedError(err) {
				continue
			}
			return "", fmt.Errorf("generate text (model %s): %w", model, err)
		}
		if resp == nil {
			return "", fmt.Errorf("generate text (model %s): empty response", model)
		}

		out := resp.Text()
		if out == "" {
			return "", fmt.Errorf("generate text (model %s): empty text", model)
		}

		return out, nil
	}

	return "", fmt.Errorf("generate text: no model candidates available")
}

func (g *Generator) GenerateProjectPlan(ctx context.Context, in model.ProjectInput) (model.ProjectPlan, error) {
	if g == nil || g.client == nil {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: client is nil")
	}
	if !isValidComplexity(in.Complexity) {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: invalid complexity: %s", in.Complexity)
	}
	if in.TechStack == nil {
		in.TechStack = []string{}
	}

	raw, err := g.GenerateText(ctx, buildProjectPlanPrompt(in))
	if err != nil {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: %w", err)
	}

	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()

	var payload json.RawMessage
	if err := dec.Decode(&payload); err != nil {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: invalid JSON: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err == nil {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: invalid JSON: trailing content")
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(payload, &top); err != nil {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: invalid JSON: %w", err)
	}
	if err := validateRequiredObjectKeys(top, []string{
		"title",
		"description",
		"app_type",
		"complexity",
		"tech_stack",
		"goal",
		"estimated_time",
		"mvp",
		"extended_ideas",
		"possible_challenges",
		"next_steps",
	}); err != nil {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: %w", err)
	}

	var mvpObj map[string]json.RawMessage
	if err := json.Unmarshal(top["mvp"], &mvpObj); err != nil {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: invalid JSON: mvp: %w", err)
	}
	if err := validateRequiredObjectKeys(mvpObj, []string{"features", "user_flow", "success_criteria"}); err != nil {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: mvp: %w", err)
	}

	dec2 := json.NewDecoder(bytes.NewReader(payload))
	dec2.DisallowUnknownFields()

	var plan model.ProjectPlan
	if err := dec2.Decode(&plan); err != nil {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: invalid JSON: %w", err)
	}

	if !isValidComplexity(plan.Complexity) {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: invalid JSON: complexity must be beginner|intermediate|advanced")
	}
	if plan.AppType != in.AppType {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: invalid JSON: app_type must match input")
	}
	if plan.Complexity != in.Complexity {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: invalid JSON: complexity must match input")
	}
	if !equalStringSlice(plan.TechStack, in.TechStack) {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: invalid JSON: tech_stack must match input")
	}
	if plan.Goal != in.Goal {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: invalid JSON: goal must match input")
	}
	if plan.EstimatedTime != in.Timeframe {
		return model.ProjectPlan{}, fmt.Errorf("generate project plan: invalid JSON: estimated_time must match input")
	}

	return plan, nil
}

func isOverloadedError(err error) bool {
	var apiErrPtr *genai.APIError
	if errors.As(err, &apiErrPtr) {
		if apiErrPtr.Code == 503 {
			return true
		}
		if apiErrPtr.Status == "UNAVAILABLE" {
			return true
		}
	}

	var apiErr genai.APIError
	if errors.As(err, &apiErr) {
		if apiErr.Code == 503 {
			return true
		}
		if apiErr.Status == "UNAVAILABLE" {
			return true
		}
	}

	return false
}
