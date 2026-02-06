package ai

import (
	"encoding/json"
	"strings"
)

type EvolutionReadinessPromptInput struct {
	CurrentProjectOverview string
	TechStackAndArchitecture string
	BuiltMVPScope          []string
	ProposedEvolution      string
}

const evolutionReadinessPromptTemplate = "" +
	"You are a Senior Engineering Manager. Be critical, skeptical, and direct.\n" +
	"Return ONLY valid JSON. Do not include explanation, formatting, markdown, or extra text.\n" +
	"You MUST return exactly one JSON object and nothing else.\n\n" +
	"Inputs:\n" +
	"- current_project_overview: {{overview}}\n" +
	"- tech_stack_and_architecture: {{tech_stack_and_architecture}}\n" +
	"- built_mvp_scope: {{built_mvp_scope_json}}\n" +
	"- proposed_evolution_next_phase: {{proposed_evolution}}\n\n" +
	"Evaluation Focus (gate criteria):\n" +
	"- System understanding: architecture, data flow, boundaries, failure modes.\n" +
	"- Engineering fundamentals: correctness, maintainability, testing strategy, dependency management.\n" +
	"- Operational readiness: observability, reliability, security posture, deployment/runbook readiness.\n" +
	"- Scope reality: the proposed evolution is feasible as a single phase and not a rewrite.\n\n" +
	"Rules:\n" +
	"- Verdict must be exactly one of: READY or NOT_READY.\n" +
	"- If NOT_READY, blocking_gaps MUST be non-empty and specific.\n" +
	"- prerequisites MUST be concrete, actionable, and verifiable (not motivational).\n" +
	"- risks_if_forced must describe realistic failure modes if evolution is attempted prematurely.\n" +
	"- Do NOT be diplomatic. Do NOT add fluff.\n" +
	"- Avoid abstract advice; every item must reference the provided context.\n" +
	"- Fill EVERY field in the schema.\n" +
	"- Do NOT add, remove, or rename any fields.\n\n" +
	"Schema (must include ALL fields):\n" +
	"{\n" +
	"  \"readiness_verdict\": \"READY\" | \"NOT_READY\",\n" +
	"  \"blocking_gaps\": string[],\n" +
	"  \"concrete_prerequisites\": string[],\n" +
	"  \"risks_if_forced\": string[]\n" +
	"}\n"

func BuildEvolutionReadinessPrompt(in EvolutionReadinessPromptInput) string {
	mvpJSON, err := json.Marshal(in.BuiltMVPScope)
	if err != nil {
		mvpJSON = []byte("[]")
	}

	return renderTemplate(evolutionReadinessPromptTemplate, map[string]string{
		"{{overview}}":                 safePromptValue(in.CurrentProjectOverview),
		"{{tech_stack_and_architecture}}": safePromptValue(in.TechStackAndArchitecture),
		"{{built_mvp_scope_json}}":     string(mvpJSON),
		"{{proposed_evolution}}":       safePromptValue(in.ProposedEvolution),
	})
}

func (in EvolutionReadinessPromptInput) Canonical() EvolutionReadinessPromptInput {
	out := EvolutionReadinessPromptInput{}
	out.CurrentProjectOverview = strings.TrimSpace(in.CurrentProjectOverview)
	out.TechStackAndArchitecture = strings.TrimSpace(in.TechStackAndArchitecture)
	out.ProposedEvolution = strings.TrimSpace(in.ProposedEvolution)
	out.BuiltMVPScope = in.BuiltMVPScope
	return out
}
