package techstack

type Language struct {
	ID    string
	Label string
}

type Framework struct {
	ID    string
	Label string
}

var languages = []Language{
	{ID: "go", Label: "Go"},
	{ID: "javascript", Label: "JavaScript"},
	{ID: "typescript", Label: "TypeScript"},
	{ID: "python", Label: "Python"},
	{ID: "java", Label: "Java"},
	{ID: "kotlin", Label: "Kotlin"},
	{ID: "csharp", Label: "C#"},
	{ID: "rust", Label: "Rust"},
	{ID: "php", Label: "PHP"},
}

var frameworksByLanguage = map[string][]Framework{
	"go": {
		{ID: "go-native", Label: "Native / Standard Library (net/http)"},
		{ID: "go-net-http", Label: "net/http"},
		{ID: "go-fiber", Label: "Fiber"},
		{ID: "go-gin", Label: "Gin"},
		{ID: "go-echo", Label: "Echo"},
	},
	"javascript": {
		{ID: "nodejs", Label: "Node.js"},
		{ID: "express", Label: "Express"},
		{ID: "nest", Label: "NestJS"},
		{ID: "react", Label: "React"},
		{ID: "nextjs", Label: "Next.js"},
		{ID: "vue", Label: "Vue"},
	},
	"typescript": {
		{ID: "nodejs-ts", Label: "Node.js (TypeScript)"},
		{ID: "express-ts", Label: "Express (TypeScript)"},
		{ID: "nest-ts", Label: "NestJS"},
		{ID: "react-ts", Label: "React (TypeScript)"},
		{ID: "nextjs-ts", Label: "Next.js (TypeScript)"},
		{ID: "vue-ts", Label: "Vue (TypeScript)"},
	},
	"python": {
		{ID: "python-native", Label: "Native / Standard Library"},
		{ID: "flask", Label: "Flask"},
		{ID: "fastapi", Label: "FastAPI"},
		{ID: "django", Label: "Django"},
	},
	"java": {
		{ID: "java-native", Label: "Native / Standard Library"},
		{ID: "spring-boot", Label: "Spring Boot"},
	},
	"kotlin": {
		{ID: "kotlin-native", Label: "Native / Standard Library"},
		{ID: "ktor", Label: "Ktor"},
		{ID: "spring-kotlin", Label: "Spring (Kotlin)"},
	},
	"csharp": {
		{ID: "csharp-native", Label: "Native / Standard Library"},
		{ID: "aspnet-core", Label: "ASP.NET Core"},
	},
	"rust": {
		{ID: "rust-native", Label: "Native / Standard Library"},
		{ID: "actix", Label: "Actix Web"},
		{ID: "axum", Label: "Axum"},
		{ID: "rocket", Label: "Rocket"},
	},
	"php": {
		{ID: "php-native", Label: "Native PHP"},
		{ID: "laravel", Label: "Laravel"},
		{ID: "symfony", Label: "Symfony"},
	},
}

func Languages() []Language {
	out := make([]Language, len(languages))
	copy(out, languages)
	return out
}

func FrameworksForLanguage(languageID string) []Framework {
	fws, ok := frameworksByLanguage[languageID]
	if !ok {
		return []Framework{}
	}
	out := make([]Framework, len(fws))
	copy(out, fws)
	return out
}

func LanguageExists(languageID string) bool {
	for _, l := range languages {
		if l.ID == languageID {
			return true
		}
	}
	return false
}
