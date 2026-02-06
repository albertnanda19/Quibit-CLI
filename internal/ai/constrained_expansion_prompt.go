package ai

import (
	"strings"

	"quibit/internal/project"
)

type ConstrainedExpansionPromptInput struct {
	IdeaSpec project.IdeaSpec
}

const constrainedExpansionPromptTemplate = "" +
	"You are a Senior Engineering Lead.\n" +
	"You must follow the input specification as a FINAL contract.\n" +
	"Do NOT ask questions. Do NOT provide alternatives.\n" +
	"Do NOT change scope, direction, or complexity.\n\n" +
	"Final IdeaSpec (do NOT modify):\n" +
	"- app_type: {{app_type}}\n" +
	"- domain_focus: {{domain_focus}}\n" +
	"- core_problem: {{core_problem}}\n" +
	"- architectural_axis: {{architectural_axis}}\n" +
	"- complexity_level: {{complexity_level}}\n\n" +
	"Rules (hard constraints):\n" +
	"- You MUST keep app_type exactly the same.\n" +
	"- You MUST keep domain_focus exactly the same.\n" +
	"- You MUST keep core_problem exactly the same (only clarify wording, do not broaden/narrow).\n" +
	"- You MUST keep architectural_axis exactly the same (do not introduce a new architecture).\n" +
	"- You MUST keep complexity_level exactly the same.\n" +
	"- Do NOT add new product scope beyond what is strictly required to solve core_problem.\n" +
	"- Be conservative: prefer simpler assumptions and fewer moving parts.\n" +
	"- No fluff, no motivation, no speculative opinions.\n\n" +
	"Output format requirements:\n" +
	"- Use ONLY the section headings below, in the same order, with no extra sections.\n" +
	"- Each section must be concrete, professional, and ready for normalization.\n" +
	"- Lists must use '-' bullet lines.\n" +
	"- Do NOT use markdown code fences.\n\n" +
	"SECTION: Overview\n" +
	"- Project Name: <string>\n" +
	"- Tagline: <string>\n" +
	"- Problem: <1-3 sentences; keep the same core_problem>\n" +
	"- Target Users:\n" +
	"  - <string>\n" +
	"  - <string>\n" +
	"- Success Metrics:\n" +
	"  - <string>\n" +
	"  - <string>\n\n" +
	"SECTION: Tech Stack\n" +
	"- Backend: <string>\n" +
	"- Frontend: <string>\n" +
	"- Database: <string>\n" +
	"- Infra: <string>\n" +
	"- Justification: <3-6 sentences; must align with architectural_axis and complexity_level>\n\n" +
	"SECTION: MVP Scope\n" +
	"- Goal: <string; must align with core_problem>\n" +
	"- Must Have Features (exactly 6 items):\n" +
	"  - <string>\n" +
	"  - <string>\n" +
	"  - <string>\n" +
	"  - <string>\n" +
	"  - <string>\n" +
	"  - <string>\n" +
	"- Out Of Scope (exactly 4 items):\n" +
	"  - <string>\n" +
	"  - <string>\n" +
	"  - <string>\n" +
	"  - <string>\n\n" +
	"SECTION: Engineering Focus Areas\n" +
	"- <3-5 items; must be engineering/architecture concerns implied by architectural_axis>\n\n" +
	"SECTION: Learning Outcomes\n" +
	"- <5-7 items; must be tied to chosen architecture and focus areas>\n"

func BuildConstrainedExpansionPrompt(in ConstrainedExpansionPromptInput) string {
	s := in.IdeaSpec.Canonical()

	return renderTemplate(constrainedExpansionPromptTemplate, map[string]string{
		"{{app_type}}":           safePromptValue(s.AppType),
		"{{domain_focus}}":       safePromptValue(s.DomainFocus),
		"{{core_problem}}":       safePromptValue(s.CoreProblem),
		"{{architectural_axis}}": safePromptValue(s.ArchitecturalAxis),
		"{{complexity_level}}":   safePromptValue(s.ComplexityLevel),
	})
}

func (in ConstrainedExpansionPromptInput) Canonical() ConstrainedExpansionPromptInput {
	out := ConstrainedExpansionPromptInput{}
	out.IdeaSpec = in.IdeaSpec.Canonical()
	out.IdeaSpec.ArchitecturalAxis = strings.TrimSpace(out.IdeaSpec.ArchitecturalAxis)
	return out
}
