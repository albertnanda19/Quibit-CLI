package ai

type GenerateProjectIdeaResponse struct {
	Overview          ProjectOverview `json:"overview"`
	TechStack         TechStack       `json:"recommended_tech_stack"`
	MVPScope          MVPScope        `json:"mvp_scope"`
	LearningOutcomes  []string        `json:"learning_outcomes"`
}

type ProjectOverview struct {
	ProjectName    string   `json:"project_name"`
	Tagline        string   `json:"tagline"`
	Problem        string   `json:"problem"`
	TargetUsers    []string `json:"target_users"`
	SuccessMetrics []string `json:"success_metrics"`
}

type TechStack struct {
	Backend       string `json:"backend"`
	Frontend      string `json:"frontend"`
	Database      string `json:"database"`
	Infra         string `json:"infra"`
	Justification string `json:"justification"`
}

type MVPScope struct {
	Goal            string   `json:"goal"`
	MustHaveFeatures []string `json:"must_have_features"`
	OutOfScope      []string `json:"out_of_scope"`
}
