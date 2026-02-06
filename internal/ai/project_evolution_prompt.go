package ai

import (
	"encoding/json"
	"strings"

	"quibit/internal/project"
)

type NextPhaseEvolutionPromptInput struct {
	CurrentProjectOverview string
	TechStack             []string
	MVPScope              []string
	ProjectDNA            project.ProjectDNA
}

const nextPhaseEvolutionPromptTemplate = "" +
	"You are a Product + Engineering Lead.\n" +
	"Return ONLY valid JSON. Do not include explanation, formatting, markdown, or extra text.\n" +
	"You MUST return exactly one JSON object and nothing else.\n\n" +
	"Current Project Context (do not rewrite the core idea; evolve it):\n" +
	"- overview: {{overview}}\n" +
	"- tech_stack: {{tech_stack_json}}\n" +
	"- mvp_scope: {{mvp_scope_json}}\n\n" +
	"Project DNA (treat as technical identity constraints):\n" +
	"- app_type: {{dna_app_type}}\n" +
	"- primary_domain: {{dna_primary_domain}}\n" +
	"- core_tech_stack: {{dna_core_tech_stack_csv}}\n" +
	"- architectural_style: {{dna_architectural_style}}\n" +
	"- complexity_level: {{dna_complexity_level}}\n\n" +
	"Rules:\n" +
	"- Produce exactly ONE next-phase evolution. No alternatives.\n" +
	"- This evolution MUST meaningfully change how the system is built (architecture/process/ops/security/reliability).\n" +
	"- This evolution MUST introduce at least 2 new engineering concerns (e.g., scaling, reliability, observability, security, performance, cost, DX, data quality).\n" +
	"- Keep it realistic for a single phase by one engineer; avoid a full rewrite.\n" +
	"- Updated MVP scope must be concise and bounded.\n" +
	"- Skills learned must be explicit and tied to the architectural changes and concerns introduced.\n" +
	"- Fill EVERY field in the schema.\n" +
	"- Do NOT add, remove, or rename any fields.\n\n" +
	"Schema (must include ALL fields):\n" +
	"{\n" +
	"  \"evolution_goal\": string,\n" +
	"  \"architectural_changes\": string[],\n" +
	"  \"new_engineering_concerns\": string[],\n" +
	"  \"updated_mvp_scope\": {\n" +
	"    \"must_have\": string[],\n" +
	"    \"out_of_scope\": string[]\n" +
	"  },\n" +
	"  \"skills_and_concepts_learned\": string[]\n" +
	"}\n"

func BuildNextPhaseEvolutionPrompt(in NextPhaseEvolutionPromptInput) string {
	techStackJSON, err := json.Marshal(in.TechStack)
	if err != nil {
		techStackJSON = []byte("[]")
	}
	mvpScopeJSON, err := json.Marshal(in.MVPScope)
	if err != nil {
		mvpScopeJSON = []byte("[]")
	}

	dna := in.ProjectDNA.Canonical()

	return renderTemplate(nextPhaseEvolutionPromptTemplate, map[string]string{
		"{{overview}}":                safePromptValue(in.CurrentProjectOverview),
		"{{tech_stack_json}}":         string(techStackJSON),
		"{{mvp_scope_json}}":          string(mvpScopeJSON),
		"{{dna_app_type}}":            safePromptValue(dna.AppType),
		"{{dna_primary_domain}}":      safePromptValue(dna.PrimaryDomain),
		"{{dna_core_tech_stack_csv}}": safePromptValue(strings.Join(dna.CoreTechStack, ", ")),
		"{{dna_architectural_style}}": safePromptValue(dna.ArchitecturalStyle),
		"{{dna_complexity_level}}":    safePromptValue(dna.ComplexityLevel),
	})
}
