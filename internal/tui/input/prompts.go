package input

type Option struct {
	Label string
	Value string
}

type SelectPrompt struct {
	Title       string
	Description string
	Options     []Option
	CustomLabel string
	Default     Option
}

var ApplicationTypePrompt = SelectPrompt{
	Title:       "Application Type",
	Description: "Select the type of project you want to generate.",
	Options: []Option{
		{Label: "Web Application", Value: "web"},
		{Label: "CLI Tool", Value: "cli"},
		{Label: "Mobile Application", Value: "mobile"},
	},
	CustomLabel: "Custom",
	Default:     Option{Label: "Web Application", Value: "web"},
}

var ComplexityPrompt = SelectPrompt{
	Title:       "Complexity Level",
	Description: "Define the expected difficulty and depth of the project.",
	Options: []Option{
		{Label: "Beginner", Value: "beginner"},
		{Label: "Intermediate", Value: "intermediate"},
		{Label: "Advanced", Value: "advanced"},
	},
	CustomLabel: "Custom",
	Default:     Option{Label: "Intermediate", Value: "intermediate"},
}

var TechnologyStackPrompt = SelectPrompt{
	Title:       "Technology Stack",
	Description: "Select or customize the primary technologies.",
	Options: []Option{
		{Label: "Go", Value: "go"},
		{Label: "Node.js", Value: "nodejs"},
		{Label: "Python", Value: "python"},
		{Label: "Rust", Value: "rust"},
	},
	CustomLabel: "Custom (comma separated)",
	Default:     Option{Label: "Go", Value: "go"},
}

var ProjectGoalPrompt = SelectPrompt{
	Title:       "Project Goal",
	Description: "What is the main purpose of this project?",
	Options: []Option{
		{Label: "Portfolio Project", Value: "portfolio project"},
		{Label: "Learning Experiment", Value: "learning experiment"},
		{Label: "Open Source Tool", Value: "open source tool"},
	},
	CustomLabel: "Custom",
	Default:     Option{Label: "Portfolio Project", Value: "portfolio project"},
}

var EstimatedTimeframePrompt = SelectPrompt{
	Title:       "Estimated Timeframe",
	Description: "Expected development duration.",
	Options: []Option{
		{Label: "1-2 weeks", Value: "1-2 weeks"},
		{Label: "2-4 weeks", Value: "2-4 weeks"},
		{Label: "1-3 months", Value: "1-3 months"},
	},
	CustomLabel: "Custom",
	Default:     Option{Label: "2-4 weeks", Value: "2-4 weeks"},
}
