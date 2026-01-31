package ai

import (
	"encoding/json"
	"strings"

	"quibit/internal/model"
)

type ProjectIdea struct {
	Project ProjectIdeaProject `json:"project"`
}

type ProjectIdeaProject struct {
	Name        string             `json:"name"`
	Tagline     string             `json:"tagline"`
	Description ProjectDescription `json:"description"`
	Problem     ProjectProblem     `json:"problem_statement"`
	TargetUsers ProjectTargetUsers `json:"target_users"`
	ValueProp   ProjectValueProp   `json:"value_proposition"`
	MVP         ProjectMVP         `json:"mvp"`
	TechStack   ProjectTechStack   `json:"recommended_tech_stack"`
	Complexity  string             `json:"complexity"`
	Duration    ProjectDuration    `json:"estimated_duration"`
	Future      []string           `json:"future_extensions"`
	Learning    []string           `json:"learning_outcomes"`
}

type ProjectDescription struct {
	Summary             string `json:"summary"`
	DetailedExplanation string `json:"detailed_explanation"`
}

type ProjectProblem struct {
	Problem                 string `json:"problem"`
	WhyItMatters            string `json:"why_it_matters"`
	CurrentSolutionsAndGaps string `json:"current_solutions_and_gaps"`
}

type ProjectTargetUsers struct {
	Primary   []string `json:"primary"`
	Secondary []string `json:"secondary"`
	UseCases  []string `json:"use_cases"`
}

type ProjectValueProp struct {
	KeyBenefits                 []string `json:"key_benefits"`
	WhyThisProjectIsInteresting string   `json:"why_this_project_is_interesting"`
	PortfolioValue              string   `json:"portfolio_value"`
}

type ProjectMVP struct {
	Goal       string   `json:"goal"`
	MustHave   []string `json:"must_have_features"`
	NiceToHave []string `json:"nice_to_have_features"`
	OutOfScope []string `json:"out_of_scope"`
}

type ProjectTechStack struct {
	Backend       string `json:"backend"`
	Frontend      string `json:"frontend"`
	Database      string `json:"database"`
	Infra         string `json:"infra"`
	Justification string `json:"justification"`
}

type ProjectDuration struct {
	Range       string `json:"range"`
	Assumptions string `json:"assumptions"`
}

type EvolutionInput struct {
	ProjectOverview   string
	MVPScope          []string
	TechStack         []string
	Complexity        string
	EstimatedDuration string
	AppType           string
	Goal              string
}

type ProjectEvolution struct {
	EvolutionOverview    string   `json:"evolution_overview"`
	ProductRationale     string   `json:"product_rationale"`
	TechnicalRationale   string   `json:"technical_rationale"`
	ProposedEnhancements []string `json:"proposed_enhancements"`
	RiskConsiderations   []string `json:"risk_considerations"`
}

type RetryReason string

const (
	RetrySimilarityTooHigh RetryReason = "SIMILARITY_TOO_HIGH"
	RetryUserRejected      RetryReason = "USER_REJECTED"
	RetryDuplicateDNA      RetryReason = "DUPLICATE_DNA"
)

type PivotStrategy string

const (
	PivotChangeTargetUser   PivotStrategy = "CHANGE_TARGET_USER"
	PivotFeatureReplacement PivotStrategy = "FEATURE_REPLACEMENT"
	PivotContextShift       PivotStrategy = "CONTEXT_SHIFT"
)

func BuildProjectIdeaPrompt(in model.ProjectInput) string {
	techStackJSON, err := json.Marshal(in.TechStack)
	if err != nil {
		techStackJSON = []byte("[]")
	}
	projectKind := strings.TrimSpace(in.ProjectKind)
	projectKindLine := ""
	if projectKind != "" {
		projectKindLine = "- project_kind: " + projectKind + "\n"
	}
	projectKindRule := ""
	if projectKind == "" {
		projectKindRule = "- If project_kind is not provided, you MUST infer a suitable software category based on tech_stack and typical real-world use.\n"
	}
	dbPref := strings.TrimSpace(in.Database)
	dbLine := ""
	if dbPref != "" && strings.ToLower(dbPref) != "none" {
		dbLine = "- database_preference: " + dbPref + "\n"
	}
	if strings.ToLower(dbPref) == "none" {
		dbLine = "- database_preference: none\n"
	}

	return "Return ONLY valid JSON. Do not include explanation, formatting, markdown, or extra text.\n" +
		"You MUST return exactly one JSON object and nothing else.\n\n" +
		"User Input (use these as constraints):\n" +
		"- app_type: " + in.AppType + "\n" +
		projectKindLine +
		dbLine +
		"- complexity: " + in.Complexity + "\n" +
		"- tech_stack: " + string(techStackJSON) + "\n" +
		"- goal: " + in.Goal + "\n" +
		"- estimated_duration: " + in.Timeframe + "\n\n" +
		"Rules:\n" +
		projectKindRule +
		"- complexity must match input exactly (beginner|intermediate|advanced).\n" +
		"- estimated_duration.range must match input exactly.\n" +
		"- recommended_tech_stack must respect tech_stack constraints (no unrelated additions).\n" +
		"- Provide concrete, professional, portfolio-ready content (no marketing fluff).\n" +
		"- MVP must be truly minimal and focused.\n" +
		"- Provide explicit product and technical reasoning.\n" +
		"- Fill EVERY field in the schema.\n" +
		"- Do NOT add, remove, or rename any fields.\n\n" +
		"Schema (must include ALL fields):\n" +
		"{\n" +
		"  \"project\": {\n" +
		"    \"name\": string,\n" +
		"    \"tagline\": string,\n" +
		"    \"description\": {\n" +
		"      \"summary\": string,\n" +
		"      \"detailed_explanation\": string\n" +
		"    },\n" +
		"    \"problem_statement\": {\n" +
		"      \"problem\": string,\n" +
		"      \"why_it_matters\": string,\n" +
		"      \"current_solutions_and_gaps\": string\n" +
		"    },\n" +
		"    \"target_users\": {\n" +
		"      \"primary\": string[],\n" +
		"      \"secondary\": string[],\n" +
		"      \"use_cases\": string[]\n" +
		"    },\n" +
		"    \"value_proposition\": {\n" +
		"      \"key_benefits\": string[],\n" +
		"      \"why_this_project_is_interesting\": string,\n" +
		"      \"portfolio_value\": string\n" +
		"    },\n" +
		"    \"mvp\": {\n" +
		"      \"goal\": string,\n" +
		"      \"must_have_features\": string[],\n" +
		"      \"nice_to_have_features\": string[],\n" +
		"      \"out_of_scope\": string[]\n" +
		"    },\n" +
		"    \"recommended_tech_stack\": {\n" +
		"      \"backend\": string,\n" +
		"      \"frontend\": string,\n" +
		"      \"database\": string,\n" +
		"      \"infra\": string,\n" +
		"      \"justification\": string\n" +
		"    },\n" +
		"    \"complexity\": \"beginner\" | \"intermediate\" | \"advanced\",\n" +
		"    \"estimated_duration\": {\n" +
		"      \"range\": string,\n" +
		"      \"assumptions\": string\n" +
		"    },\n" +
		"    \"future_extensions\": string[],\n" +
		"    \"learning_outcomes\": string[]\n" +
		"  }\n" +
		"}\n"
}

func BuildProjectEvolutionPrompt(in EvolutionInput) string {
	mvpJSON, err := json.Marshal(in.MVPScope)
	if err != nil {
		mvpJSON = []byte("[]")
	}
	techStackJSON, err := json.Marshal(in.TechStack)
	if err != nil {
		techStackJSON = []byte("[]")
	}

	return "Return ONLY valid JSON. Do not include explanation, formatting, markdown, or extra text.\n" +
		"You MUST return exactly one JSON object and nothing else.\n\n" +
		"Project Context (do not change core idea):\n" +
		"- project_overview: " + in.ProjectOverview + "\n" +
		"- mvp_scope: " + string(mvpJSON) + "\n" +
		"- tech_stack: " + string(techStackJSON) + "\n" +
		"- complexity: " + in.Complexity + "\n" +
		"- estimated_duration: " + in.EstimatedDuration + "\n" +
		"- app_type: " + in.AppType + "\n" +
		"- goal: " + in.Goal + "\n\n" +
		"Rules:\n" +
		"- Do NOT change the core idea or reframe the product.\n" +
		"- Focus on next-step evolution and advanced development.\n" +
		"- Provide clear product rationale and technical rationale.\n" +
		"- Fill EVERY field in the schema.\n" +
		"- Do NOT add, remove, or rename any fields.\n\n" +
		"Schema (must include ALL fields):\n" +
		"{\n" +
		"  \"evolution_overview\": string,\n" +
		"  \"product_rationale\": string,\n" +
		"  \"technical_rationale\": string,\n" +
		"  \"proposed_enhancements\": string[],\n" +
		"  \"risk_considerations\": string[]\n" +
		"}\n"
}

func BuildProjectIdeaPivotPrompt(in model.ProjectInput, reason RetryReason, strategy PivotStrategy) string {
	base := BuildProjectIdeaPrompt(in)
	return base + "\n" +
		"Regeneration:\n" +
		"- retry_reason: " + string(reason) + "\n" +
		"- pivot_strategy: " + string(strategy) + "\n\n" +
		"Pivot Strategy Instructions:\n" +
		pivotStrategyInstruction(strategy) +
		"\n" +
		"Rules:\n" +
		"- You MUST follow the pivot strategy.\n" +
		"- The new idea must be meaningfully different from the previous attempt.\n"
}

func pivotStrategyInstruction(strategy PivotStrategy) string {
	switch strategy {
	case PivotChangeTargetUser:
		return "- Change the target user segment and adjust the value proposition to fit the new audience."
	case PivotFeatureReplacement:
		return "- Replace 2-3 key MVP items with different capabilities and adjust the main workflow."
	case PivotContextShift:
		return "- Shift the domain context or problem framing while keeping the input constraints."
	default:
		return "- Replace 2-3 key MVP items with different capabilities and adjust the main workflow."
	}
}

func normalizeWhitespace(s string) string {
	return strings.TrimSpace(s)
}
