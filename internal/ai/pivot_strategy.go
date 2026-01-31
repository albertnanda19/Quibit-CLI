package ai

import (
	"fmt"
	"strings"

	"quibit/internal/domain"
)

type Pivot struct {
	Reason string
	Prompt string
}

func BuildPivot(dominant string, ref domain.Project) Pivot {
	dominant = strings.TrimSpace(strings.ToLower(dominant))

	reason := dominant
	pivot := ""

	switch dominant {
	case "features":
		pivot = "Replace 1-2 core features with different capabilities and change the main workflow."
	case "tech stack":
		pivot = "Shift the implementation approach and architecture while keeping within the allowed tech constraints."
	case "target users":
		pivot = "Change the target user segment to a different audience and adjust the value proposition accordingly."
	case "title":
		pivot = "Change the project framing and title to a distinct concept and domain context."
	default:
		reason = "features"
		pivot = "Replace 1-2 core features with different capabilities and change the main workflow."
	}

	refTitle := strings.TrimSpace(ref.Title)
	refProblem := strings.TrimSpace(ref.ProblemStatement)

	prompt := "Must significantly differ from the reference project below.\n" +
		"You must NOT reuse the same title, and you must NOT keep more than half of the same core features.\n" +
		"Pivot instruction: " + pivot + "\n" +
		"Reference project to avoid:\n" +
		"- title: " + safeLine(refTitle) + "\n" +
		"- problem_statement: " + safeLine(refProblem) + "\n" +
		"- core_features: [" + safeCSV(ref.CoreFeatures) + "]\n" +
		"- target_users: [" + safeCSV(ref.TargetUsers) + "]\n" +
		"- recommended_stack: " + safeLine(ref.RecommendedStack) + "\n" +
		"- estimated_complexity: " + safeLine(ref.EstimatedComplexity) + "\n"

	return Pivot{Reason: reason, Prompt: prompt}
}

func safeLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.TrimSpace(s)
	if s == "" {
		return "-"
	}
	return s
}

func safeCSV(items []string) string {
	out := make([]string, 0, len(items))
	for _, v := range items {
		v = safeLine(v)
		if v == "-" {
			continue
		}
		out = append(out, v)
	}
	return fmt.Sprintf("%s", strings.Join(out, ", "))
}
