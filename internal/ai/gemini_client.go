package ai

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

func NewGeminiClient(ctx context.Context, apiKey string) (*genai.Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("create gemini client: %w", err)
	}

	return client, nil
}
