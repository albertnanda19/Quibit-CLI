package model

type ProjectMVP struct {
	Features        []string `json:"features"`
	UserFlow        string   `json:"user_flow"`
	SuccessCriteria string   `json:"success_criteria"`
}

type ProjectPlan struct {
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	AppType            string     `json:"app_type"`
	Complexity         string     `json:"complexity"`
	TechStack          []string   `json:"tech_stack"`
	Goal               string     `json:"goal"`
	EstimatedTime      string     `json:"estimated_time"`
	MVP                ProjectMVP `json:"mvp"`
	ExtendedIdeas      []string   `json:"extended_ideas"`
	PossibleChallenges []string   `json:"possible_challenges"`
	NextSteps          []string   `json:"next_steps"`
}
