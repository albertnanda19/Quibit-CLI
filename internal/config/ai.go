package config

type AIConfig struct {
	GeminiAPIKey string
	HFToken      string
}

func LoadAIConfig() AIConfig {
	_ = LoadDotEnv(".env")
	return AIConfig{
		GeminiAPIKey: GetenvOptional("GEMINI_API_KEY"),
		HFToken:      GetenvOptional("HF_TOKEN"),
	}
}
