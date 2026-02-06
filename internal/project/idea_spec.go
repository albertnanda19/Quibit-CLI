package project

import "strings"

type IdeaSpec struct {
	AppType           string
	DomainFocus       string
	CoreProblem       string
	ArchitecturalAxis string
	ComplexityLevel   string
}

func (s IdeaSpec) Canonical() IdeaSpec {
	out := IdeaSpec{}
	out.AppType = normalizeScalar(s.AppType)
	out.DomainFocus = normalizeScalar(s.DomainFocus)
	out.CoreProblem = normalizeScalar(s.CoreProblem)
	out.ArchitecturalAxis = canonicalAxis(s.ArchitecturalAxis)
	out.ComplexityLevel = normalizeScalar(s.ComplexityLevel)
	return out
}

func (s IdeaSpec) FingerprintString() string {
	c := s.Canonical()
	parts := []string{
		"app_type=" + c.AppType,
		"domain_focus=" + c.DomainFocus,
		"core_problem=" + c.CoreProblem,
		"architectural_axis=" + c.ArchitecturalAxis,
		"complexity_level=" + c.ComplexityLevel,
	}
	return strings.Join(parts, "|")
}

func canonicalAxis(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}

	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return normalizeScalar(s)
}
