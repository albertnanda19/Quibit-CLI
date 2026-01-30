package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"encoding/json"

	"google.golang.org/genai"
)

type ProjectIdea struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Complexity   string   `json:"complexity"`
	TechStack    []string `json:"tech_stack"`
	CoreFeatures []string `json:"core_features"`
	Twist        string   `json:"twist"`
}

var modelCandidates = []string{
	"gemini-3-flash-preview",
	"gemini-2.5-flash",
}

const projectIdeaPrompt = "Return ONLY valid JSON. Do not include explanation or formatting. Do not include markdown.\n\n" +
	"Schema (must include ALL fields):\n" +
	"{\n" +
	"  \"title\": string,\n" +
	"  \"description\": string,\n" +
	"  \"complexity\": \"beginner\" | \"intermediate\" | \"advanced\",\n" +
	"  \"tech_stack\": string[],\n" +
	"  \"core_features\": string[],\n" +
	"  \"twist\": string\n" +
	"}\n\n" +
	"Task: Generate one simple software project idea. If unsure, still fill all fields."

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

func (g *Generator) GenerateProjectIdea(ctx context.Context) (ProjectIdea, error) {
	if g == nil || g.client == nil {
		return ProjectIdea{}, fmt.Errorf("generate project idea: client is nil")
	}

	raw, err := g.GenerateText(ctx, projectIdeaPrompt)
	if err != nil {
		return ProjectIdea{}, fmt.Errorf("generate project idea: %w", err)
	}

	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()

	var idea ProjectIdea
	if err := dec.Decode(&idea); err != nil {
		return ProjectIdea{}, fmt.Errorf("generate project idea: invalid JSON: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err == nil {
		return ProjectIdea{}, fmt.Errorf("generate project idea: invalid JSON: trailing content")
	}

	if idea.Title == "" {
		return ProjectIdea{}, fmt.Errorf("generate project idea: invalid JSON: missing title")
	}
	if idea.Description == "" {
		return ProjectIdea{}, fmt.Errorf("generate project idea: invalid JSON: missing description")
	}
	switch idea.Complexity {
	case "beginner", "intermediate", "advanced":
	default:
		return ProjectIdea{}, fmt.Errorf("generate project idea: invalid JSON: complexity must be beginner|intermediate|advanced")
	}
	if len(idea.TechStack) == 0 {
		return ProjectIdea{}, fmt.Errorf("generate project idea: invalid JSON: missing tech_stack")
	}
	if len(idea.CoreFeatures) == 0 {
		return ProjectIdea{}, fmt.Errorf("generate project idea: invalid JSON: missing core_features")
	}
	if idea.Twist == "" {
		return ProjectIdea{}, fmt.Errorf("generate project idea: invalid JSON: missing twist")
	}

	return idea, nil
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
