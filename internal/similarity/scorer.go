package similarity

import (
	"math"

	"quibit/internal/domain"
)

type Breakdown struct {
	TitleSimilarity            float64
	ProblemStatementSimilarity float64
	CoreFeaturesOverlap        float64
	TargetUsersOverlap         float64
	TechStackOverlap           float64
	ComplexityMatch            float64
	Total                      float64
}

func Score(a domain.Project, b domain.Project) Breakdown {
	title := jaccard(tokenizeText(a.Title), tokenizeText(b.Title))
	problem := jaccard(tokenizeText(a.ProblemStatement), tokenizeText(b.ProblemStatement))
	core := jaccard(normalizeList(a.CoreFeatures), normalizeList(b.CoreFeatures))
	users := jaccard(normalizeList(a.TargetUsers), normalizeList(b.TargetUsers))
	tech := jaccard(normalizeTechStack(a.RecommendedStack), normalizeTechStack(b.RecommendedStack))

	complexity := 0.0
	if a.EstimatedComplexity != "" && a.EstimatedComplexity == b.EstimatedComplexity {
		complexity = 1.0
	}

	total := 0.0
	total += title * 0.15
	total += problem * 0.25
	total += core * 0.25
	total += users * 0.15
	total += tech * 0.10
	total += complexity * 0.10

	return Breakdown{
		TitleSimilarity:            title,
		ProblemStatementSimilarity: problem,
		CoreFeaturesOverlap:        core,
		TargetUsersOverlap:         users,
		TechStackOverlap:           tech,
		ComplexityMatch:            complexity,
		Total:                      clamp01(total),
	}
}

func DominantDimension(b Breakdown) string {
	bestKey := "title"
	best := b.TitleSimilarity * 0.15

	if v := b.CoreFeaturesOverlap * 0.25; v > best {
		best = v
		bestKey = "features"
	}
	if v := b.TechStackOverlap * 0.10; v > best {
		best = v
		bestKey = "tech stack"
	}
	if v := b.TargetUsersOverlap * 0.15; v > best {
		best = v
		bestKey = "target users"
	}

	return bestKey
}

func jaccard(a []string, b []string) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}

	na := normalizeList(a)
	nb := normalizeList(b)

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
		return 0
	}
	return float64(intersection) / float64(union)
}

func clamp01(v float64) float64 {
	if math.IsNaN(v) {
		return 0
	}
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
