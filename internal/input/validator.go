package input

import (
	"fmt"
	"strings"
)

func validateComplexity(v string) error {
	switch v {
	case "beginner", "intermediate", "advanced":
		return nil
	default:
		return fmt.Errorf("Error: complexity must be Beginner | Intermediate | Advanced")
	}
}

func normalizeComplexity(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}
