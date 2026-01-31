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
		{Label: "Desktop Application", Value: "desktop"},
		{Label: "Machine Learning Project", Value: "ml"},
		{Label: "CLI Tool", Value: "cli"},
		{Label: "Mobile Application", Value: "mobile"},
		{Label: "Backend API / Service", Value: "backend-api"},
	},
	CustomLabel: "Custom",
	Default:     Option{Label: "Web Application", Value: "web"},
}

var ProjectKindPrompt = SelectPrompt{
	Title:       "Project Category (Optional)",
	Description: "Optional: pick a specific kind of software (or skip to keep results like before).",
	Options: []Option{
		{Label: "Skip (no preference)", Value: ""},
		{Label: "LMS (Learning Management System)", Value: "lms"},
		{Label: "ERP (Enterprise Resource Planning)", Value: "erp"},
		{Label: "CRM (Customer Relationship Management)", Value: "crm"},
		{Label: "SCM (Supply Chain Management)", Value: "scm"},
		{Label: "E-commerce", Value: "ecommerce"},
		{Label: "FinTech / Accounting", Value: "fintech"},
		{Label: "Healthcare", Value: "healthcare"},
		{Label: "Marketplace", Value: "marketplace"},
		{Label: "Mobile-first version", Value: "mobile-first"},
		{Label: "AI project / AI-powered app", Value: "ai-project"},
	},
	CustomLabel: "Custom",
	Default:     Option{Label: "Skip (no preference)", Value: ""},
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
		{Label: "React", Value: "react"},
		{Label: "Vue", Value: "vue"},
		{Label: "Astro", Value: "astro"},
		{Label: "Laravel", Value: "laravel"},
		{Label: "Spring Boot", Value: "spring-boot"},
		{Label: "Node.js", Value: "nodejs"},
		{Label: "Python", Value: "python"},
		{Label: "Rust", Value: "rust"},
		{Label: "Java", Value: "java"},
		{Label: "TypeScript", Value: "typescript"},
	},
	CustomLabel: "Custom (comma separated)",
	Default:     Option{Label: "Go", Value: "go"},
}

var DatabasePrompt = SelectPrompt{
	Title:       "Database",
	Description: "Select a database preference (or choose no database).",
	Options: []Option{
		{Label: "No database", Value: "none"},
		{Label: "PostgreSQL", Value: "postgresql"},
		{Label: "MySQL", Value: "mysql"},
		{Label: "SQLite", Value: "sqlite"},
		{Label: "MongoDB", Value: "mongodb"},
		{Label: "Redis", Value: "redis"},
	},
	CustomLabel: "Custom",
	Default:     Option{Label: "No database", Value: "none"},
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
