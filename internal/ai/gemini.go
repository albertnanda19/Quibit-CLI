package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"quibit/internal/config"
	"quibit/internal/model"
)

type GeminiProvider struct {
	apiKey string
}

func NewGeminiProvider(cfg config.AIConfig) *GeminiProvider {
	return &GeminiProvider{apiKey: cfg.GeminiAPIKey}
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) Generate(ctx context.Context, prompt PromptPayload) (AIResult, error) {
	if ctx == nil {
		return AIResult{}, fmt.Errorf("gemini: ctx is nil")
	}
	if p == nil {
		return AIResult{}, fmt.Errorf("gemini: not initialized")
	}
	if strings.TrimSpace(prompt.Prompt) == "" {
		return AIResult{}, fmt.Errorf("gemini: prompt is empty")
	}

	client, err := NewGeminiClient(ctx, p.apiKey)
	if err != nil {
		return AIResult{}, err
	}

	g, err := NewGenerator(client)
	if err != nil {
		return AIResult{}, err
	}

	text, err := g.GenerateText(ctx, prompt.Prompt)
	if err != nil {
		return AIResult{}, err
	}

	return AIResult{Text: text, ProviderUsed: p.Name()}, nil
}

type staticErrorProvider struct {
	name string
	err  error
}

func (p staticErrorProvider) Name() string { return p.name }
func (p staticErrorProvider) Generate(ctx context.Context, prompt PromptPayload) (AIResult, error) {
	return AIResult{}, p.err
}

func newDefaultProviderManager() (*ProviderManager, error) {
	cfg := config.LoadAIConfig()

	primary := AIProvider(NewGeminiProvider(cfg))
	var fallback AIProvider
	hf, err := NewHuggingFaceProvider(cfg)
	if err != nil {
		fallback = staticErrorProvider{name: "huggingface", err: err}
	} else {
		fallback = hf
	}

	return NewProviderManager(primary, fallback)
}

func GenerateProjectIdea(ctx context.Context, in model.ProjectInput) (ProjectIdea, string, error) {
	idea, raw, _, err := GenerateProjectIdeaWithMeta(ctx, in)
	return idea, raw, err
}

func GenerateProjectIdeaOnceWithMeta(ctx context.Context, in model.ProjectInput) (ProjectIdea, string, AIResult, error) {
	m, err := newDefaultProviderManager()
	if err != nil {
		return ProjectIdea{}, "", AIResult{}, err
	}

	prompt := BuildProjectIdeaPrompt(in)
	res, err := m.Generate(ctx, PromptPayload{Prompt: prompt})
	if err != nil {
		return ProjectIdea{}, "", AIResult{}, err
	}

	raw := normalizePromptContractJSON(res.Text)
	idea, err := decodeProjectIdea(raw, in)
	if err != nil {
		return ProjectIdea{}, "", res, err
	}
	return idea, raw, res, nil
}

func GenerateProjectIdeaWithMeta(ctx context.Context, in model.ProjectInput) (ProjectIdea, string, AIResult, error) {
	m, err := newDefaultProviderManager()
	if err != nil {
		return ProjectIdea{}, "", AIResult{}, err
	}

	const maxQualityAttempts = 4
	var lastErr error
	var lastMeta AIResult
	var lastVerdict *qualityVerdict
	for attempt := 0; attempt < maxQualityAttempts; attempt++ {
		prompt := BuildProjectIdeaPrompt(in)
		if attempt > 0 {
			strategy := rotatePivotStrategy(attempt)
			if lastVerdict != nil {
				switch lastVerdict.decision {
				case qualityRefine:
					strategy = PivotRefineDepth
				case qualityPivot:
					strategy = PivotContextShift
				case qualityRegenerate:
					strategy = rotatePivotStrategy(attempt)
				}
			}
			prompt = BuildProjectIdeaPivotPrompt(in, RetryQualityTooGeneric, strategy)
		}

		idea, raw, meta, err := generateProjectIdeaWithPrompt(ctx, m, prompt, in)
		lastMeta = meta
		if err != nil {
			lastErr = err
			continue
		}

		v := evaluateIdeaQuality(idea)
		if v.ok() {
			return idea, raw, meta, nil
		}

		lastVerdict = &v
		lastErr = fmt.Errorf("generate project idea: quality gate failed: %s", v.summary())
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("generate project idea: quality gate failed")
	}
	return ProjectIdea{}, "", lastMeta, lastErr
}

func GenerateProjectIdeaWithPivot(ctx context.Context, in model.ProjectInput, reason RetryReason, strategy PivotStrategy) (ProjectIdea, string, error) {
	idea, raw, _, err := GenerateProjectIdeaWithPivotMeta(ctx, in, reason, strategy)
	return idea, raw, err
}

func GenerateProjectIdeaWithPivotOnceMeta(ctx context.Context, in model.ProjectInput, reason RetryReason, strategy PivotStrategy) (ProjectIdea, string, AIResult, error) {
	m, err := newDefaultProviderManager()
	if err != nil {
		return ProjectIdea{}, "", AIResult{}, err
	}

	prompt := BuildProjectIdeaPivotPrompt(in, reason, strategy)
	res, err := m.Generate(ctx, PromptPayload{Prompt: prompt})
	if err != nil {
		return ProjectIdea{}, "", AIResult{}, err
	}

	raw := normalizePromptContractJSON(res.Text)
	idea, err := decodeProjectIdea(raw, in)
	if err != nil {
		return ProjectIdea{}, "", res, err
	}
	return idea, raw, res, nil
}

func GenerateProjectIdeaWithPivotMeta(ctx context.Context, in model.ProjectInput, reason RetryReason, strategy PivotStrategy) (ProjectIdea, string, AIResult, error) {
	m, err := newDefaultProviderManager()
	if err != nil {
		return ProjectIdea{}, "", AIResult{}, err
	}

	const maxQualityAttempts = 4
	var lastErr error
	var lastMeta AIResult
	var lastVerdict *qualityVerdict
	for attempt := 0; attempt < maxQualityAttempts; attempt++ {
		var prompt string
		if attempt == 0 {
			prompt = BuildProjectIdeaPivotPrompt(in, reason, strategy)
		} else {
			nextStrategy := rotatePivotStrategy(attempt)
			if lastVerdict != nil {
				switch lastVerdict.decision {
				case qualityRefine:
					nextStrategy = PivotRefineDepth
				case qualityPivot:
					nextStrategy = PivotContextShift
				case qualityRegenerate:
					nextStrategy = rotatePivotStrategy(attempt)
				}
			}
			prompt = BuildProjectIdeaPivotPrompt(in, RetryQualityTooGeneric, nextStrategy)
		}

		idea, raw, meta, err := generateProjectIdeaWithPrompt(ctx, m, prompt, in)
		lastMeta = meta
		if err != nil {
			lastErr = err
			continue
		}

		v := evaluateIdeaQuality(idea)
		if v.ok() {
			return idea, raw, meta, nil
		}
		lastVerdict = &v
		lastErr = fmt.Errorf("generate project idea: quality gate failed: %s", v.summary())
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("generate project idea: quality gate failed")
	}
	return ProjectIdea{}, "", lastMeta, lastErr
}

func GenerateProjectEvolution(ctx context.Context, in EvolutionInput) (ProjectEvolution, string, error) {
	evo, raw, _, err := GenerateProjectEvolutionWithMeta(ctx, in)
	return evo, raw, err
}

func rotatePivotStrategy(attempt int) PivotStrategy {
	switch attempt % 3 {
	case 1:
		return PivotChangeTargetUser
	case 2:
		return PivotContextShift
	default:
		return PivotFeatureReplacement
	}
}

func GenerateProjectEvolutionWithMeta(ctx context.Context, in EvolutionInput) (ProjectEvolution, string, AIResult, error) {
	m, err := newDefaultProviderManager()
	if err != nil {
		return ProjectEvolution{}, "", AIResult{}, err
	}

	res, err := m.Generate(ctx, PromptPayload{Prompt: BuildProjectEvolutionPrompt(in)})
	if err != nil {
		return ProjectEvolution{}, "", AIResult{}, err
	}

	raw := normalizePromptContractJSON(res.Text)
	evo, err := decodeProjectEvolution(raw)
	if err != nil {
		return ProjectEvolution{}, "", AIResult{}, err
	}

	return evo, raw, res, nil
}

func decodeProjectIdea(raw string, in model.ProjectInput) (ProjectIdea, error) {
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()

	var payload json.RawMessage
	if err := dec.Decode(&payload); err != nil {
		return ProjectIdea{}, fmt.Errorf("generate project idea: invalid JSON: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err == nil {
		return ProjectIdea{}, fmt.Errorf("generate project idea: invalid JSON: trailing content")
	}

	dec2 := json.NewDecoder(strings.NewReader(string(payload)))
	dec2.DisallowUnknownFields()

	var idea ProjectIdea
	if err := dec2.Decode(&idea); err != nil {
		return ProjectIdea{}, fmt.Errorf("generate project idea: invalid JSON: %w", err)
	}

	if err := validateProjectIdea(idea, in); err != nil {
		return ProjectIdea{}, err
	}

	return idea, nil
}

func generateProjectIdeaWithPrompt(ctx context.Context, m *ProviderManager, prompt string, in model.ProjectInput) (ProjectIdea, string, AIResult, error) {
	const maxAttempts = 3
	var lastErr error
	var lastMeta AIResult
	for i := 0; i < maxAttempts; i++ {
		res, err := m.Generate(ctx, PromptPayload{Prompt: prompt})
		if err != nil {
			lastErr = err
			continue
		}
		lastMeta = res
		raw := normalizePromptContractJSON(res.Text)
		idea, err := decodeProjectIdea(raw, in)
		if err != nil {
			lastErr = err
			continue
		}
		return idea, raw, res, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("generate project idea: empty response")
	}
	return ProjectIdea{}, "", lastMeta, lastErr
}

func validateProjectIdea(idea ProjectIdea, in model.ProjectInput) error {
	p := idea.Project
	if isTooShort(p.Name, 5) {
		return fmt.Errorf("generate project idea: invalid JSON: name is too short")
	}
	if isTooShort(p.Tagline, 8) {
		return fmt.Errorf("generate project idea: invalid JSON: tagline is too short")
	}
	if isTooShort(p.Description.Summary, 40) {
		return fmt.Errorf("generate project idea: invalid JSON: summary is too short")
	}
	if isTooShort(p.Description.DetailedExplanation, 120) {
		return fmt.Errorf("generate project idea: invalid JSON: detailed_explanation is too short")
	}
	if isTooShort(p.Problem.Problem, 40) {
		return fmt.Errorf("generate project idea: invalid JSON: problem is too short")
	}
	if isTooShort(p.Problem.WhyItMatters, 40) {
		return fmt.Errorf("generate project idea: invalid JSON: why_it_matters is too short")
	}
	if isTooShort(p.Problem.CurrentSolutionsAndGaps, 40) {
		return fmt.Errorf("generate project idea: invalid JSON: current_solutions_and_gaps is too short")
	}
	if countNonEmpty(p.TargetUsers.Primary) == 0 {
		return fmt.Errorf("generate project idea: invalid JSON: target_users.primary is required")
	}
	if countNonEmpty(p.TargetUsers.UseCases) == 0 {
		return fmt.Errorf("generate project idea: invalid JSON: target_users.use_cases is required")
	}
	if countNonEmpty(p.ValueProp.KeyBenefits) < 2 {
		return fmt.Errorf("generate project idea: invalid JSON: value_proposition.key_benefits is too short")
	}
	if isTooShort(p.ValueProp.WhyThisProjectIsInteresting, 40) {
		return fmt.Errorf("generate project idea: invalid JSON: why_this_project_is_interesting is too short")
	}
	if isTooShort(p.ValueProp.PortfolioValue, 40) {
		return fmt.Errorf("generate project idea: invalid JSON: portfolio_value is too short")
	}
	if isTooShort(p.MVP.Goal, 30) {
		return fmt.Errorf("generate project idea: invalid JSON: mvp.goal is too short")
	}
	if countNonEmpty(p.MVP.MustHave) < 3 {
		return fmt.Errorf("generate project idea: invalid JSON: mvp.must_have_features is too short")
	}
	if countNonEmpty(p.MVP.NiceToHave) == 0 {
		return fmt.Errorf("generate project idea: invalid JSON: mvp.nice_to_have_features is required")
	}
	if countNonEmpty(p.MVP.OutOfScope) == 0 {
		return fmt.Errorf("generate project idea: invalid JSON: mvp.out_of_scope is required")
	}
	if isTooShort(p.TechStack.Backend, 2) ||
		isTooShort(p.TechStack.Frontend, 2) ||
		isTooShort(p.TechStack.Database, 2) ||
		isTooShort(p.TechStack.Infra, 2) {
		return fmt.Errorf("generate project idea: invalid JSON: recommended_tech_stack fields are required")
	}
	if isTooShort(p.TechStack.Justification, 60) {
		return fmt.Errorf("generate project idea: invalid JSON: recommended_tech_stack.justification is too short")
	}
	if normalizeWhitespace(p.Complexity) != in.Complexity {
		return fmt.Errorf("generate project idea: invalid JSON: complexity must match input")
	}
	if normalizeWhitespace(p.Duration.Range) != in.Timeframe {
		return fmt.Errorf("generate project idea: invalid JSON: estimated_duration.range must match input")
	}
	if countNonEmpty(p.Future) < 2 {
		return fmt.Errorf("generate project idea: invalid JSON: future_extensions is too short")
	}
	if countNonEmpty(p.Learning) < 3 {
		return fmt.Errorf("generate project idea: invalid JSON: learning_outcomes is too short")
	}
	if !matchesInputTechStack(p.TechStack, in.TechStack) {
		return fmt.Errorf("generate project idea: invalid JSON: recommended_tech_stack must respect input tech_stack")
	}
	return nil
}

func isTooShort(v string, min int) bool {
	return len(strings.TrimSpace(v)) < min
}

func countNonEmpty(items []string) int {
	n := 0
	for _, v := range items {
		if strings.TrimSpace(v) == "" {
			continue
		}
		n++
	}
	return n
}

func matchesInputTechStack(stack ProjectTechStack, input []string) bool {
	if len(input) == 0 {
		return true
	}
	joined := strings.ToLower(strings.Join([]string{
		stack.Backend,
		stack.Frontend,
		stack.Database,
		stack.Infra,
		stack.Justification,
	}, " "))
	joinedNorm := normalizeToken(joined)
	for _, v := range input {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		required := requiredTechTokens(v)
		if len(required) == 0 {
			return false
		}
		for _, t := range required {
			if t == "" {
				continue
			}
			if !strings.Contains(joinedNorm, t) {
				return false
			}
		}
	}
	return true
}

func normalizeToken(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func requiredTechTokens(v string) []string {
	parts := splitAlphaNumTokens(v)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = normalizeToken(p)
		if p == "" {
			continue
		}
		if isTechDescriptor(p) {
			continue
		}
		out = append(out, p)
	}

	if len(out) == 0 {
		fallback := normalizeToken(v)
		if fallback != "" && !isTechDescriptor(fallback) {
			out = append(out, fallback)
		}
	}
	return out
}

func splitAlphaNumTokens(s string) []string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return nil
	}
	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			cur.WriteRune(r)
			continue
		}
		flush()
	}
	flush()
	return out
}

func isTechDescriptor(t string) bool {
	switch t {
	case "frontend", "backend", "api", "mvc", "fullstack", "monolith", "service":
		return true
	default:
		return false
	}
}

func decodeProjectEvolution(raw string) (ProjectEvolution, error) {
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.DisallowUnknownFields()

	var payload json.RawMessage
	if err := dec.Decode(&payload); err != nil {
		return ProjectEvolution{}, fmt.Errorf("generate project evolution: invalid JSON: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err == nil {
		return ProjectEvolution{}, fmt.Errorf("generate project evolution: invalid JSON: trailing content")
	}

	dec2 := json.NewDecoder(strings.NewReader(string(payload)))
	dec2.DisallowUnknownFields()

	var evo ProjectEvolution
	if err := dec2.Decode(&evo); err != nil {
		return ProjectEvolution{}, fmt.Errorf("generate project evolution: invalid JSON: %w", err)
	}

	if normalizeWhitespace(evo.EvolutionOverview) == "" {
		return ProjectEvolution{}, fmt.Errorf("generate project evolution: invalid JSON: evolution_overview is required")
	}
	if normalizeWhitespace(evo.ProductRationale) == "" {
		return ProjectEvolution{}, fmt.Errorf("generate project evolution: invalid JSON: product_rationale is required")
	}
	if normalizeWhitespace(evo.TechnicalRationale) == "" {
		return ProjectEvolution{}, fmt.Errorf("generate project evolution: invalid JSON: technical_rationale is required")
	}
	if len(evo.ProposedEnhancements) == 0 {
		return ProjectEvolution{}, fmt.Errorf("generate project evolution: invalid JSON: proposed_enhancements is required")
	}

	return evo, nil
}
