package ai

import "context"

type PromptPayload struct {
	Prompt string
}

type AIResult struct {
	Text         string
	ProviderUsed string

	FallbackUsed bool

	ProviderError string

	LatencyMS int64
}

type AIProvider interface {
	Generate(ctx context.Context, prompt PromptPayload) (AIResult, error)
	Name() string
}
