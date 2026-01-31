package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"quibit/internal/config"
)

const (
	hfRouterBaseURL = "https://router.huggingface.co/v1"
	hfDefaultModel  = "moonshotai/Kimi-K2-Instruct-0905"
)

type HuggingFaceProvider struct {
	baseURL string
	model   string
	token   string
	client  *http.Client
}

func NewHuggingFaceProvider(cfg config.AIConfig) (*HuggingFaceProvider, error) {
	if strings.TrimSpace(cfg.HFToken) == "" {
		return nil, fmt.Errorf("HF_TOKEN is required")
	}
	return &HuggingFaceProvider{
		baseURL: hfRouterBaseURL,
		model:   hfDefaultModel,
		token:   cfg.HFToken,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (p *HuggingFaceProvider) Name() string { return "huggingface" }

func (p *HuggingFaceProvider) Generate(ctx context.Context, prompt PromptPayload) (AIResult, error) {
	if ctx == nil {
		return AIResult{}, fmt.Errorf("huggingface: ctx is nil")
	}
	if p == nil || p.client == nil {
		return AIResult{}, fmt.Errorf("huggingface: not initialized")
	}
	if strings.TrimSpace(prompt.Prompt) == "" {
		return AIResult{}, fmt.Errorf("huggingface: prompt is empty")
	}

	type message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	type reqBody struct {
		Model    string    `json:"model"`
		Messages []message `json:"messages"`
	}

	body := reqBody{
		Model: p.model,
		Messages: []message{
			{Role: "system", Content: ""},
			{Role: "user", Content: prompt.Prompt},
		},
	}
	b, err := json.Marshal(body)
	if err != nil {
		return AIResult{}, fmt.Errorf("huggingface: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(p.baseURL, "/")+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return AIResult{}, fmt.Errorf("huggingface: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	start := time.Now()
	resp, err := p.client.Do(req)
	if err != nil {
		return AIResult{}, fmt.Errorf("huggingface: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	rawBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(rawBody))
		if msg == "" {
			msg = resp.Status
		}
		return AIResult{}, fmt.Errorf("huggingface: http %d: %s", resp.StatusCode, msg)
	}

	type chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	var out chatResp
	if err := json.Unmarshal(rawBody, &out); err != nil {
		return AIResult{}, fmt.Errorf("huggingface: decode response: %w", err)
	}
	if len(out.Choices) == 0 {
		return AIResult{}, fmt.Errorf("huggingface: empty choices")
	}
	text := strings.TrimSpace(out.Choices[0].Message.Content)
	if text == "" {
		return AIResult{}, fmt.Errorf("huggingface: empty content")
	}

	return AIResult{
		Text:         text,
		ProviderUsed: p.Name(),
		LatencyMS:    time.Since(start).Milliseconds(),
	}, nil
}
