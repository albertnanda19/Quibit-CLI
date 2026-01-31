package dna

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"

	"quibit/internal/domain"
)

func HashProject(p domain.Project) string {
	norm := func(s string) string {
		s = strings.TrimSpace(s)
		s = strings.ToLower(s)
		return s
	}

	core := make([]string, 0, len(p.CoreFeatures))
	for _, v := range p.CoreFeatures {
		core = append(core, norm(v))
	}
	sort.Strings(core)

	users := make([]string, 0, len(p.TargetUsers))
	for _, v := range p.TargetUsers {
		users = append(users, norm(v))
	}
	sort.Strings(users)

	parts := []string{
		norm(p.Title),
		norm(p.ProblemStatement),
		strings.Join(core, ","),
		strings.Join(users, ","),
		norm(p.RecommendedStack),
		norm(p.EstimatedComplexity),
	}

	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}
