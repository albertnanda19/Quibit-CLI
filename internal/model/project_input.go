package model

type ProjectInput struct {
	UserIdea    string
	AppType     string
	ProjectKind string
	Complexity  string
	TechStack   []string
	Database    []string
	Goal        string
	Timeframe   string
}
