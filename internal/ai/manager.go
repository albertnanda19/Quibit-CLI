package ai

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"google.golang.org/genai"
)

type ProviderManager struct {
	primary  AIProvider
	fallback AIProvider
}

func NewProviderManager(primary AIProvider, fallback AIProvider) (*ProviderManager, error) {
	if primary == nil {
		return nil, fmt.Errorf("ai manager: primary provider is nil")
	}
	if fallback == nil {
		return nil, fmt.Errorf("ai manager: fallback provider is nil")
	}
	return &ProviderManager{primary: primary, fallback: fallback}, nil
}

func (m *ProviderManager) Generate(ctx context.Context, prompt PromptPayload) (AIResult, error) {
	if ctx == nil {
		return AIResult{}, fmt.Errorf("ai manager: ctx is nil")
	}
	if m == nil || m.primary == nil || m.fallback == nil {
		return AIResult{}, fmt.Errorf("ai manager: not initialized")
	}

	start := time.Now()

	res, err := m.primary.Generate(ctx, prompt)
	if err == nil {
		res.LatencyMS = time.Since(start).Milliseconds()
		return res, nil
	}

	if !shouldFallbackFromPrimary(err) {
		return AIResult{}, err
	}

	primaryErr := err

	res2, err2 := m.fallback.Generate(ctx, prompt)
	if err2 != nil {
		return AIResult{}, fmt.Errorf("ai manager: primary(%s) err=%s; fallback(%s) err=%w",
			m.primary.Name(),
			sanitizeErr(primaryErr),
			m.fallback.Name(),
			err2,
		)
	}

	res2.FallbackUsed = true
	res2.ProviderError = sanitizeErr(primaryErr)
	res2.LatencyMS = time.Since(start).Milliseconds()
	return res2, nil
}

func sanitizeErr(err error) string {
	if err == nil {
		return ""
	}
	s := strings.TrimSpace(err.Error())
	if len(s) > 2000 {
		return s[:2000]
	}
	return s
}

func shouldFallbackFromPrimary(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	if isRateLimitedError(err) {
		return true
	}

	return true
}

func isRateLimitedError(err error) bool {
	if err == nil {
		return false
	}

	var apiErrPtr *genai.APIError
	if errors.As(err, &apiErrPtr) && apiErrPtr != nil {
		if apiErrPtr.Code == 429 {
			return true
		}
		switch strings.ToUpper(strings.TrimSpace(apiErrPtr.Status)) {
		case "RESOURCE_EXHAUSTED", "TOO_MANY_REQUESTS":
			return true
		}
	}
	var apiErr genai.APIError
	if errors.As(err, &apiErr) {
		if apiErr.Code == 429 {
			return true
		}
		switch strings.ToUpper(strings.TrimSpace(apiErr.Status)) {
		case "RESOURCE_EXHAUSTED", "TOO_MANY_REQUESTS":
			return true
		}
	}

	s := strings.ToLower(err.Error())
	if strings.Contains(s, "429") && strings.Contains(s, "rate") {
		return true
	}
	if strings.Contains(s, "too many requests") || strings.Contains(s, "resource exhausted") || strings.Contains(s, "rate limit") {
		return true
	}
	return false
}
