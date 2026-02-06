package ai

import (
	"strings"
)

type GenerateProjectIdeaPromptInput struct {
	AppType     string
	DomainGoal  string
	Complexity  string
}

const generateProjectIdeaPromptTemplate = "" +
	"You are an Expert Engineering Lead.\n" +
	"Return ONLY valid JSON. Do not include explanation, formatting, markdown, or extra text.\n" +
	"You MUST return exactly one JSON object and nothing else.\n\n" +
	"User Input (use these as strict constraints):\n" +
	"- app_type: {{app_type}}\n" +
	"- domain_goal: {{domain_goal}}\n" +
	"- complexity: {{complexity}}\n\n" +
	"Rules:\n" +
	"- Be deterministic: avoid vague language, avoid multiple alternative options for the same decision.\n" +
	"- Provide an engineering-led, portfolio-ready idea with clear trade-offs.\n" +
	"- Keep scope realistic for a single engineer and aligned to the requested complexity.\n" +
	"- MVP must be minimal and tightly focused (5-8 features max).\n" +
	"- Recommended tech stack must be rational and justified; do not list trendy tools without reason.\n" +
	"- Learning outcomes must be explicit and directly connected to the chosen architecture/stack.\n" +
	"- Fill EVERY field in the schema.\n" +
	"- Do NOT add, remove, or rename any fields.\n\n" +
	"Schema (must include ALL fields):\n" +
	"{\n" +
	"  \"overview\": {\n" +
	"    \"project_name\": string,\n" +
	"    \"tagline\": string,\n" +
	"    \"problem\": string,\n" +
	"    \"target_users\": string[],\n" +
	"    \"success_metrics\": string[]\n" +
	"  },\n" +
	"  \"recommended_tech_stack\": {\n" +
	"    \"backend\": string,\n" +
	"    \"frontend\": string,\n" +
	"    \"database\": string,\n" +
	"    \"infra\": string,\n" +
	"    \"justification\": string\n" +
	"  },\n" +
	"  \"mvp_scope\": {\n" +
	"    \"goal\": string,\n" +
	"    \"must_have_features\": string[],\n" +
	"    \"out_of_scope\": string[]\n" +
	"  },\n" +
	"  \"learning_outcomes\": string[]\n" +
	"}\n"

func BuildGenerateProjectIdeaPrompt(in GenerateProjectIdeaPromptInput) string {
	return renderTemplate(generateProjectIdeaPromptTemplate, map[string]string{
		"{{app_type}}":    safePromptValue(in.AppType),
		"{{domain_goal}}": safePromptValue(in.DomainGoal),
		"{{complexity}}":  safePromptValue(in.Complexity),
	})
}

func renderTemplate(tpl string, values map[string]string) string {
	out := tpl
	for k, v := range values {
		out = strings.ReplaceAll(out, k, v)
	}
	return out
}

func safePromptValue(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.TrimSpace(s)
}
