package project

import (
	"regexp"
	"strings"
)

type Snapshot struct {
	Overview          string
	MVPScope          []string
	TechStack         []string
	Complexity        string
	EstimatedDuration string
	AppType           string
	Goal              string
}

type SimilarityDecision int

const (
	SimilarityOK SimilarityDecision = iota
	SimilarityRegenerate
	SimilarityBlock
)

func DecideSimilarity(score float64) SimilarityDecision {
	switch {
	case score >= 0.75:
		return SimilarityBlock
	case score >= 0.55:
		return SimilarityRegenerate
	default:
		return SimilarityOK
	}
}

func JaccardSimilarity(a, b Snapshot) float64 {
	setA := tokenSet(a)
	setB := tokenSet(b)
	if len(setA) == 0 && len(setB) == 0 {
		return 0
	}
	intersect := 0
	for k := range setA {
		if _, ok := setB[k]; ok {
			intersect++
		}
	}
	union := len(setA) + len(setB) - intersect
	if union == 0 {
		return 0
	}
	return float64(intersect) / float64(union)
}

var tokenRe = regexp.MustCompile(`[a-z0-9]+`)

func tokenSet(s Snapshot) map[string]struct{} {
	out := make(map[string]struct{})
	add := func(v string) {
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "" {
			return
		}
		for _, t := range tokenRe.FindAllString(v, -1) {
			if t == "" {
				continue
			}
			out[t] = struct{}{}
		}
	}

	add(s.Overview)
	for _, v := range s.MVPScope {
		add(v)
	}
	for _, v := range s.TechStack {
		add(v)
	}
	add(s.Complexity)
	add(s.EstimatedDuration)
	add(s.AppType)
	add(s.Goal)

	return out
}
