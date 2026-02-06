package input

import "quibit/internal/techstack"

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

func ProgrammingLanguagePromptWithDefault(defaultLanguageID string) SelectPrompt {
	languages := techstack.Languages()
	options := make([]Option, 0, len(languages))
	defaultLabel := ""
	for _, l := range languages {
		options = append(options, Option{Label: l.Label, Value: l.ID})
		if l.ID == defaultLanguageID {
			defaultLabel = l.Label
		}
	}
	if defaultLabel == "" {
		defaultLanguageID = "go"
		defaultLabel = "Go"
	}

	return SelectPrompt{
		Title:       "Programming Language",
		Description: "Pick a language first. You can also choose Custom/Manual.",
		Options:     options,
		CustomLabel: "Custom / Manual Choice…",
		Default:     Option{Label: defaultLabel, Value: defaultLanguageID},
	}
}

func ProgrammingLanguagePrompt() SelectPrompt {
	return ProgrammingLanguagePromptWithDefault("go")
}

func FrameworkPrompt(languageID string) SelectPrompt {
	fws := techstack.FrameworksForLanguage(languageID)
	options := make([]Option, 0, len(fws))
	for _, fw := range fws {
		options = append(options, Option{Label: fw.Label, Value: fw.ID})
	}

	defaultOpt := Option{Label: "Custom / Manual Choice…", Value: ""}
	if len(options) > 0 {
		defaultOpt = options[0]
	}

	return SelectPrompt{
		Title:       "Framework / Library",
		Description: "Pick a framework/library/native based on your chosen language.",
		Options:     options,
		CustomLabel: "Custom / Manual Choice…",
		Default:     defaultOpt,
	}
}

var ApplicationTypePrompt = SelectPrompt{
	Title:       "Application Type",
	Description: "Choose what you’re building.",
	Options: []Option{
		{Label: "Web Application", Value: "web"},
		{Label: "Desktop Application", Value: "desktop"},
		{Label: "Machine Learning Project", Value: "ml"},
		{Label: "CLI Tool", Value: "cli"},
		{Label: "Mobile Application", Value: "mobile"},
		{Label: "Backend API / Service", Value: "backend-api"},
	},
	CustomLabel: "Custom…",
	Default:     Option{Label: "Web Application", Value: "web"},
}

var WebArchitecturePrompt = SelectPrompt{
	Title:       "Web Architecture",
	Description: "Select a structure for the web app.",
	Options: []Option{
		{Label: "MVC (monolith)", Value: "mvc"},
		{Label: "Frontend + Backend (separate)", Value: "split"},
	},
	CustomLabel: "Custom…",
	Default:     Option{Label: "Frontend + Backend (separate)", Value: "split"},
}

var WebMVCFrameworkPrompt = SelectPrompt{
	Title:       "MVC Framework",
	Description: "Pick a framework for an MVC-style web app.",
	Options: []Option{
		{Label: "Laravel", Value: "laravel"},
		{Label: "Next.js (fullstack)", Value: "nextjs"},
		{Label: "Django", Value: "django"},
		{Label: "Ruby on Rails", Value: "rails"},
		{Label: "Spring Boot (MVC)", Value: "spring-boot-mvc"},
		{Label: "ASP.NET Core MVC", Value: "aspnet-mvc"},
	},
	CustomLabel: "Custom…",
	Default:     Option{Label: "Laravel", Value: "laravel"},
}

var WebFrontendFrameworkPrompt = SelectPrompt{
	Title:       "Frontend Framework",
	Description: "Pick a frontend framework.",
	Options: []Option{
		{Label: "React", Value: "react"},
		{Label: "Vue", Value: "vue"},
		{Label: "Next.js (frontend)", Value: "nextjs-frontend"},
		{Label: "Nuxt", Value: "nuxt"},
		{Label: "SvelteKit", Value: "sveltekit"},
		{Label: "Angular", Value: "angular"},
		{Label: "Astro", Value: "astro"},
	},
	CustomLabel: "Custom…",
	Default:     Option{Label: "React", Value: "react"},
}

var WebBackendFrameworkPrompt = SelectPrompt{
	Title:       "Backend Framework",
	Description: "Pick a backend framework.",
	Options: []Option{
		{Label: "Go (Gin)", Value: "go-gin"},
		{Label: "Go (Fiber)", Value: "go-fiber"},
		{Label: "Node.js (Express)", Value: "nodejs-express"},
		{Label: "Node.js (NestJS)", Value: "nodejs-nestjs"},
		{Label: "Spring Boot", Value: "spring-boot"},
		{Label: "Laravel (API)", Value: "laravel-api"},
		{Label: "Django REST Framework", Value: "django-rest"},
		{Label: "FastAPI", Value: "fastapi"},
	},
	CustomLabel: "Custom…",
	Default:     Option{Label: "Go (Gin)", Value: "go-gin"},
}

func WebSplitStackSelectionModePrompt(part string) SelectPrompt {
	return SelectPrompt{
		Title:       part + " Selection",
		Description: "Choose whether to pick a language first or pick a framework directly.",
		Options: []Option{
			{Label: "Choose language first", Value: "language"},
			{Label: "Choose framework directly", Value: "framework"},
		},
		CustomLabel: "Custom…",
		Default:     Option{Label: "Choose framework directly", Value: "framework"},
	}
}

var ProjectKindPrompt = SelectPrompt{
	Title:       "Project Category (Optional)",
	Description: "Optionally bias the generator toward a specific domain.",
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
	CustomLabel: "Custom…",
	Default:     Option{Label: "Skip (no preference)", Value: ""},
}

var ComplexityPrompt = SelectPrompt{
	Title:       "Complexity Level",
	Description: "Select the target depth.",
	Options: []Option{
		{Label: "Beginner", Value: "beginner"},
		{Label: "Intermediate", Value: "intermediate"},
		{Label: "Advanced", Value: "advanced"},
	},
	CustomLabel: "Custom…",
	Default:     Option{Label: "Intermediate", Value: "intermediate"},
}

var DatabasePrompt = SelectPrompt{
	Title:       "Database(s)",
	Description: "Select a database preference (or none).",
	Options: []Option{
		{Label: "No database", Value: "none"},
		{Label: "PostgreSQL", Value: "postgresql"},
		{Label: "MySQL", Value: "mysql"},
		{Label: "SQLite", Value: "sqlite"},
		{Label: "MongoDB", Value: "mongodb"},
		{Label: "Redis", Value: "redis"},
	},
	CustomLabel: "Custom… (comma-separated)",
	Default:     Option{Label: "No database", Value: "none"},
}

var ProjectGoalPrompt = SelectPrompt{
	Title:       "Project Goal",
	Description: "Choose the primary intent.",
	Options: []Option{
		{Label: "Portfolio Project", Value: "portfolio project"},
		{Label: "Learning Experiment", Value: "learning experiment"},
		{Label: "Open Source Tool", Value: "open source tool"},
		{Label: "SaaS (Business product)", Value: "saas business product"},
		{Label: "Business / B2B tool (internal ops)", Value: "business b2b internal tool"},
		{Label: "Real-world solution for non-technical users", Value: "real-world solution for non-technical users"},
	},
	CustomLabel: "Custom…",
	Default:     Option{Label: "Portfolio Project", Value: "portfolio project"},
}

var EstimatedTimeframePrompt = SelectPrompt{
	Title:       "Estimated Timeframe",
	Description: "Select an expected delivery window.",
	Options: []Option{
		{Label: "1-2 weeks", Value: "1-2 weeks"},
		{Label: "2-4 weeks", Value: "2-4 weeks"},
		{Label: "1-3 months", Value: "1-3 months"},
	},
	CustomLabel: "Custom…",
	Default:     Option{Label: "2-4 weeks", Value: "2-4 weeks"},
}
