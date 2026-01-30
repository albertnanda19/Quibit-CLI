package ai

import (
	"encoding/json"

	"quibit/internal/model"
)

func buildProjectPlanPrompt(in model.ProjectInput) string {
	techStackJSON, err := json.Marshal(in.TechStack)
	if err != nil {
		techStackJSON = []byte("[]")
	}

	return "Return ONLY valid JSON. Do not include explanation, formatting, markdown, or extra text.\n" +
		"You MUST return exactly one JSON object and nothing else.\n\n" +
		"User Input (use these as constraints):\n" +
		"- app_type: " + in.AppType + "\n" +
		"- complexity: " + in.Complexity + "\n" +
		"- tech_stack: " + string(techStackJSON) + "\n" +
		"- goal: " + in.Goal + "\n" +
		"- estimated_time: " + in.Timeframe + "\n\n" +
		"Rules:\n" +
		"- Use app_type exactly as provided.\n" +
		"- Use complexity exactly as provided (must be beginner|intermediate|advanced).\n" +
		"- Use tech_stack exactly as provided (do not add new items).\n" +
		"- Use goal exactly as provided.\n" +
		"- Use estimated_time exactly as provided.\n" +
		"- Fill EVERY field in the schema.\n" +
		"- Do NOT add, remove, or rename any fields.\n\n" +
		"Schema (must include ALL fields):\n" +
		"{\n" +
		"  \"title\": string,\n" +
		"  \"description\": string,\n" +
		"  \"app_type\": string,\n" +
		"  \"complexity\": \"beginner\" | \"intermediate\" | \"advanced\",\n" +
		"  \"tech_stack\": string[],\n" +
		"  \"goal\": string,\n" +
		"  \"estimated_time\": string,\n\n" +
		"  \"mvp\": {\n" +
		"    \"features\": string[],\n" +
		"    \"user_flow\": string,\n" +
		"    \"success_criteria\": string\n" +
		"  },\n\n" +
		"  \"extended_ideas\": string[],\n" +
		"  \"possible_challenges\": string[],\n" +
		"  \"next_steps\": string[]\n" +
		"}\n"
}
