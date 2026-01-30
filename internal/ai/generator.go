package ai

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/genai"
)

var modelCandidates = []string{
	"gemini-3-flash-preview",
	"gemini-2.5-flash",
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
