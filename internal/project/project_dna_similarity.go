package project

import (
	"regexp"
	"strings"
)

type ProjectDNASimilarityWeights struct {
	AppType            float64
	PrimaryDomain      float64
	CoreTechStack      float64
	ArchitecturalStyle float64
	ComplexityLevel    float64
}

func DefaultProjectDNASimilarityWeights() ProjectDNASimilarityWeights {
	return ProjectDNASimilarityWeights{
		AppType:            0.20,
		PrimaryDomain:      0.30,
		CoreTechStack:      0.25,
		ArchitecturalStyle: 0.15,
		ComplexityLevel:    0.10,
	}
}

type ProjectDNASimilarityBreakdown struct {
	AppTypeSimilarity            float64
	PrimaryDomainSimilarity      float64
	CoreTechStackSimilarity      float64
	ArchitecturalStyleSimilarity float64
	ComplexityLevelMatch         float64
	Total                        float64
}

func ScoreProjectDNASimilarity(a, b ProjectDNA) ProjectDNASimilarityBreakdown {
	w := DefaultProjectDNASimilarityWeights()
	return ScoreProjectDNASimilarityWithWeights(a, b, w)
}

func ScoreProjectDNASimilarityWithWeights(a, b ProjectDNA, w ProjectDNASimilarityWeights) ProjectDNASimilarityBreakdown {
	ca := a.Canonical()
	cb := b.Canonical()

	appType := scalarSimilarity(ca.AppType, cb.AppType)
	domain := scalarSimilarity(ca.PrimaryDomain, cb.PrimaryDomain)
	arch := scalarSimilarity(ca.ArchitecturalStyle, cb.ArchitecturalStyle)
	tech := jaccardStrings(ca.CoreTechStack, cb.CoreTechStack)

	complexity := 0.0
	if ca.ComplexityLevel != "" && ca.ComplexityLevel == cb.ComplexityLevel {
		complexity = 1.0
	}

	total := 0.0
	total += appType * w.AppType
	total += domain * w.PrimaryDomain
	total += tech * w.CoreTechStack
	total += arch * w.ArchitecturalStyle
	total += complexity * w.ComplexityLevel

	return ProjectDNASimilarityBreakdown{
		AppTypeSimilarity:            clamp01(appType),
		PrimaryDomainSimilarity:      clamp01(domain),
		CoreTechStackSimilarity:      clamp01(tech),
		ArchitecturalStyleSimilarity: clamp01(arch),
		ComplexityLevelMatch:         clamp01(complexity),
		Total:                        clamp01(total),
	}
}

var dnaTokenRe = regexp.MustCompile(`[a-z0-9]+`)

func scalarSimilarity(a, b string) float64 {
	a = normalizeScalar(a)
	b = normalizeScalar(b)
	if a == "" || b == "" {
		return 0.0
	}
	if a == b {
		return 1.0
	}
	return jaccardTokenText(a, b)
}

func jaccardTokenText(a, b string) float64 {
	setA := dnaTokenSet(a)
	setB := dnaTokenSet(b)
	if len(setA) == 0 && len(setB) == 0 {
		return 0.0
	}
	inter := 0
	for k := range setA {
		if _, ok := setB[k]; ok {
			inter++
		}
	}
	union := len(setA) + len(setB) - inter
	if union == 0 {
		return 0.0
	}
	return float64(inter) / float64(union)
}

func dnaTokenSet(s string) map[string]struct{} {
	s = strings.TrimSpace(strings.ToLower(s))
	out := make(map[string]struct{})
	if s == "" {
		return out
	}
	for _, t := range dnaTokenRe.FindAllString(s, -1) {
		if t == "" {
			continue
		}
		out[t] = struct{}{}
	}
	return out
}

func jaccardStrings(a, b []string) float64 {
	na := normalizeStringList(a)
	nb := normalizeStringList(b)
	if len(na) == 0 && len(nb) == 0 {
		return 0.0
	}

	i, j := 0, 0
	intersection := 0
	union := 0
	for i < len(na) && j < len(nb) {
		if na[i] == nb[j] {
			intersection++
			union++
			i++
			j++
			continue
		}
		union++
		if na[i] < nb[j] {
			i++
		} else {
			j++
		}
	}
	union += (len(na) - i) + (len(nb) - j)
	if union == 0 {
		return 0.0
	}
	return float64(intersection) / float64(union)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
