package similarity

import (
	"os"
	"strconv"
)

type Thresholds struct {
	AcceptableMax float64
	TooSimilarMax float64
}

func DefaultThresholds() Thresholds {
	return Thresholds{AcceptableMax: 0.55, TooSimilarMax: 0.75}
}

func LoadThresholdsFromEnv() Thresholds {
	t := DefaultThresholds()

	if v := os.Getenv("SIMILARITY_ACCEPTABLE_MAX"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			t.AcceptableMax = f
		}
	}
	if v := os.Getenv("SIMILARITY_TOO_SIMILAR_MAX"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			t.TooSimilarMax = f
		}
	}

	return t
}

type Category int

const (
	CategoryAcceptable Category = iota + 1
	CategoryTooSimilar
	CategoryDuplicate
)

func Categorize(score float64, t Thresholds) Category {
	if score < t.AcceptableMax {
		return CategoryAcceptable
	}
	if score < t.TooSimilarMax {
		return CategoryTooSimilar
	}
	return CategoryDuplicate
}
