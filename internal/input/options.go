package input

type SelectOption struct {
	Label string
	Value string
}

type SelectPrompt struct {
	Title       string
	Description string
	Options     []SelectOption
	CustomLabel string
	Default     SelectOption
}

var ApplicationTypePrompt = SelectPrompt{
	Title:       "Application Type",
	Description: "Select the type of project you want to generate.",
	Options: []SelectOption{
		{Label: "Web Application", Value: "web"},
		{Label: "CLI Tool", Value: "cli"},
		{Label: "Mobile Application", Value: "mobile"},
	},
	CustomLabel: "Custom",
	Default:     SelectOption{Label: "Web Application", Value: "web"},
}

var WebArchitecturePrompt = SelectPrompt{
	Title:       "Web Architecture",
	Description: "Choose how you want to structure the web application.",
	Options: []SelectOption{
		{Label: "MVC (monolith)", Value: "mvc"},
		{Label: "Frontend + Backend (separate)", Value: "split"},
	},
	CustomLabel: "Custom",
	Default:     SelectOption{Label: "Frontend + Backend (separate)", Value: "split"},
}

var WebMVCFrameworkPrompt = SelectPrompt{
	Title:       "MVC Framework",
	Description: "Pick a framework for an MVC-style web application.",
	Options: []SelectOption{
		{Label: "Laravel", Value: "laravel"},
		{Label: "Next.js (fullstack)", Value: "nextjs"},
		{Label: "Django", Value: "django"},
		{Label: "Ruby on Rails", Value: "rails"},
		{Label: "Spring Boot (MVC)", Value: "spring-boot-mvc"},
		{Label: "ASP.NET Core MVC", Value: "aspnet-mvc"},
	},
	CustomLabel: "Custom",
	Default:     SelectOption{Label: "Laravel", Value: "laravel"},
}

var WebFrontendFrameworkPrompt = SelectPrompt{
	Title:       "Frontend Framework",
	Description: "Pick a frontend framework (Custom is available).",
	Options: []SelectOption{
		{Label: "React", Value: "react"},
		{Label: "Vue", Value: "vue"},
		{Label: "Next.js (frontend)", Value: "nextjs-frontend"},
		{Label: "Nuxt", Value: "nuxt"},
		{Label: "SvelteKit", Value: "sveltekit"},
		{Label: "Angular", Value: "angular"},
		{Label: "Astro", Value: "astro"},
	},
	CustomLabel: "Custom",
	Default:     SelectOption{Label: "React", Value: "react"},
}

var WebBackendFrameworkPrompt = SelectPrompt{
	Title:       "Backend Framework",
	Description: "Pick a backend framework (Custom is available).",
	Options: []SelectOption{
		{Label: "Go (Gin)", Value: "go-gin"},
		{Label: "Go (Fiber)", Value: "go-fiber"},
		{Label: "Node.js (Express)", Value: "nodejs-express"},
		{Label: "Node.js (NestJS)", Value: "nodejs-nestjs"},
		{Label: "Spring Boot", Value: "spring-boot"},
		{Label: "Laravel (API)", Value: "laravel-api"},
		{Label: "Django REST Framework", Value: "django-rest"},
		{Label: "FastAPI", Value: "fastapi"},
	},
	CustomLabel: "Custom",
	Default:     SelectOption{Label: "Go (Gin)", Value: "go-gin"},
}

var ProjectKindPrompt = SelectPrompt{
	Title:       "Project Category (Optional)",
	Description: "Optional: choose a kind of software (or skip to keep results like before).",
	Options: []SelectOption{
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
	Default:     SelectOption{Label: "Skip (AI will recommend)", Value: ""},
}

var ComplexityPrompt = SelectPrompt{
	Title:       "Complexity Level",
	Description: "Define the expected difficulty and depth of the project.",
	Options: []SelectOption{
		{Label: "Beginner", Value: "beginner"},
		{Label: "Intermediate", Value: "intermediate"},
		{Label: "Advanced", Value: "advanced"},
	},
	CustomLabel: "Custom",
	Default:     SelectOption{Label: "Intermediate", Value: "intermediate"},
}

var TechnologyStackPrompt = SelectPrompt{
	Title:       "Technology Stack",
	Description: "Select or customize the primary technologies.",
	Options: []SelectOption{
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
	Default:     SelectOption{Label: "Go", Value: "go"},
}

var DatabasePrompt = SelectPrompt{
	Title:       "Database",
	Description: "Select a database preference (or choose no database).",
	Options: []SelectOption{
		{Label: "No database", Value: "none"},
		{Label: "PostgreSQL", Value: "postgresql"},
		{Label: "MySQL", Value: "mysql"},
		{Label: "SQLite", Value: "sqlite"},
		{Label: "MongoDB", Value: "mongodb"},
		{Label: "Redis", Value: "redis"},
	},
	CustomLabel: "Custom",
	Default:     SelectOption{Label: "No database", Value: "none"},
}

var ProjectGoalPrompt = SelectPrompt{
	Title:       "Project Goal",
	Description: "What is the main purpose of this project?",
	Options: []SelectOption{
		{Label: "Portfolio Project", Value: "portfolio project"},
		{Label: "Learning Experiment", Value: "learning experiment"},
		{Label: "Open Source Tool", Value: "open source tool"},
	},
	CustomLabel: "Custom",
	Default:     SelectOption{Label: "Portfolio Project", Value: "portfolio project"},
}

var EstimatedTimeframePrompt = SelectPrompt{
	Title:       "Estimated Timeframe",
	Description: "Expected development duration.",
	Options: []SelectOption{
		{Label: "1–2 weeks", Value: "1–2 weeks"},
		{Label: "2–4 weeks", Value: "2–4 weeks"},
		{Label: "1–3 months", Value: "1–3 months"},
	},
	CustomLabel: "Custom",
	Default:     SelectOption{Label: "2–4 weeks", Value: "2–4 weeks"},
}
