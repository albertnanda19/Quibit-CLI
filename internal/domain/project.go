package domain

type Project struct {
	Title              string   `json:"title"`
	Summary            string   `json:"summary"`
	ProblemStatement   string   `json:"problem_statement"`
	TargetUsers        []string `json:"target_users"`
	CoreFeatures       []string `json:"core_features"`
	MVPScope           []string `json:"mvp_scope"`
	OptionalExtensions []string `json:"optional_extensions"`
	RecommendedStack   string   `json:"recommended_stack"`
	EstimatedComplexity string  `json:"estimated_complexity"`
	EstimatedDuration  string  `json:"estimated_duration"`
}
