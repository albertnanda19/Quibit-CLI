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
		return AIResult{}, fmt.Errorf("ai manager: generation failed\n\nPrimary provider (%s)\n- Error: %s\n- Diagnosis: %s\n- What you can do: %s\n\nFallback provider (%s)\n- Error: %s\n- Diagnosis: %s\n- What you can do: %s",
			m.primary.Name(),
			sanitizeErr(primaryErr),
			primaryDiagnosis(primaryErr),
			primaryActions(primaryErr),
			m.fallback.Name(),
			sanitizeErr(err2),
			fallbackDiagnosis(err2),
			fallbackActions(err2),
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

func primaryDiagnosis(err error) string {
	if err == nil {
		return "unknown"
	}
	if isRateLimitedError(err) {
		return "Gemini rejected the request due to rate limit / quota exhaustion (HTTP 429 / RESOURCE_EXHAUSTED)."
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "quota") && strings.Contains(s, "exceed") {
		return "Gemini quota exceeded."
	}
	if strings.Contains(s, "unauthorized") || strings.Contains(s, "permission") || strings.Contains(s, "api key") {
		return "Gemini authentication/authorization issue (API key missing/invalid or project permission)."
	}
	return "Primary provider failed."
}

func primaryActions(err error) string {
	if err == nil {
		return "Retry generation."
	}
	if isRateLimitedError(err) {
		retry := extractRetryHint(err)
		if retry != "" {
			return "Wait and retry (server suggested delay: " + retry + "). If this keeps happening, check Gemini API quotas/billing for the project behind GEMINI_API_KEY, or switch to another key/project/model."
		}
		return "Wait briefly and retry. If this keeps happening, check Gemini API quotas/billing for the project behind GEMINI_API_KEY, or switch to another key/project/model."
	}
	return "Check GEMINI_API_KEY in your .env, confirm billing/quota, then retry."
}

func fallbackDiagnosis(err error) string {
	if err == nil {
		return "unknown"
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "hf_token is required") {
		return "Fallback provider is configured but HF_TOKEN is missing."
	}
	if strings.Contains(s, "http 503") || strings.Contains(s, "service unavailable") {
		return "Hugging Face Router is temporarily unavailable (HTTP 503)."
	}
	if strings.Contains(s, "http 401") || strings.Contains(s, "http 403") {
		return "Hugging Face authentication/authorization failure (token invalid/insufficient)."
	}
	return "Fallback provider failed."
}

func fallbackActions(err error) string {
	if err == nil {
		return "Retry generation."
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "hf_token is required") {
		return "Set HF_TOKEN in your .env (Hugging Face access token), then retry."
	}
	if strings.Contains(s, "http 503") || strings.Contains(s, "service unavailable") {
		return "Retry after a short delay. If persistent, verify HF Router availability and ensure your HF_TOKEN is valid; you can also switch fallback provider/model if supported."
	}
	if strings.Contains(s, "http 401") || strings.Contains(s, "http 403") {
		return "Verify HF_TOKEN is correct and has access; then retry."
	}
	return "Retry. If it persists, check HF_TOKEN and network connectivity."
}

func extractRetryHint(err error) string {
	if err == nil {
		return ""
	}
	// Best-effort parsing: many provider errors include 'Please retry in ...s' or 'retryDelay:...'.
	s := err.Error()
	low := strings.ToLower(s)
	idx := strings.Index(low, "please retry in")
	if idx >= 0 {
		chunk := strings.TrimSpace(s[idx:])
		chunk = strings.TrimPrefix(strings.ToLower(chunk), "please retry in")
		chunk = strings.TrimSpace(chunk)
		// take until first period/comma/newline
		end := len(chunk)
		for i, r := range chunk {
			if r == '.' || r == ',' || r == '\n' || r == ';' {
				end = i
				break
			}
		}
		return strings.TrimSpace(chunk[:end])
	}
	idx = strings.Index(low, "retrydelay")
	if idx >= 0 {
		chunk := strings.TrimSpace(s[idx:])
		// e.g. 'retryDelay:20s'
		chunk = strings.ReplaceAll(chunk, " ", "")
		chunk = strings.ReplaceAll(chunk, "\t", "")
		pos := strings.Index(strings.ToLower(chunk), "retrydelay:")
		if pos >= 0 {
			chunk = chunk[pos+len("retrydelay:"):]
			end := len(chunk)
			for i, r := range chunk {
				if r == ',' || r == '\n' || r == ']' || r == '}' {
					end = i
					break
				}
			}
			return strings.TrimSpace(chunk[:end])
		}
	}
	return ""
}
