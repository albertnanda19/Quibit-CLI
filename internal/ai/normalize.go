package ai

import (
	"encoding/json"
	"strings"
)

func normalizePromptContractJSON(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw
	}

	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return raw
	}

	normalizeKeysRecursive(v)

	b, err := json.Marshal(v)
	if err != nil {
		return raw
	}
	return string(b)
}

func normalizeKeysRecursive(v any) {
	switch t := v.(type) {
	case map[string]any:

		out := make(map[string]any, len(t))
		for k, val := range t {
			newK := canonicalizeKey(k)
			if newK == "" {
				newK = k
			}
			out[newK] = val
		}

		for k := range t {
			delete(t, k)
		}
		for k, val := range out {
			t[k] = val
			normalizeKeysRecursive(val)
		}
	case []any:
		for i := range t {
			normalizeKeysRecursive(t[i])
		}
	default:
		return
	}
}

func canonicalizeKey(k string) string {
	k = strings.TrimSpace(k)
	if k == "" {
		return ""
	}

	k = strings.ToLower(k)
	k = strings.ReplaceAll(k, "-", "_")
	k = strings.ReplaceAll(k, " ", "_")
	k = strings.ReplaceAll(k, ".", "_")
	for strings.Contains(k, "__") {
		k = strings.ReplaceAll(k, "__", "_")
	}
	k = strings.Trim(k, "_")

	if _, ok := knownPromptContractKeys[k]; ok {
		return k
	}
	return ""
}

var knownPromptContractKeys = map[string]struct{}{

	"project":                         {},
	"name":                            {},
	"tagline":                         {},
	"description":                     {},
	"summary":                         {},
	"detailed_explanation":            {},
	"problem_statement":               {},
	"problem":                         {},
	"why_it_matters":                  {},
	"current_solutions_and_gaps":      {},
	"target_users":                    {},
	"primary":                         {},
	"secondary":                       {},
	"use_cases":                       {},
	"value_proposition":               {},
	"key_benefits":                    {},
	"why_this_project_is_interesting": {},
	"portfolio_value":                 {},
	"mvp":                             {},
	"goal":                            {},
	"must_have_features":              {},
	"nice_to_have_features":           {},
	"out_of_scope":                    {},
	"recommended_tech_stack":          {},
	"backend":                         {},
	"frontend":                        {},
	"database":                        {},
	"infra":                           {},
	"justification":                   {},
	"complexity":                      {},
	"estimated_duration":              {},
	"range":                           {},
	"assumptions":                     {},
	"future_extensions":               {},
	"learning_outcomes":               {},

	"evolution_overview":    {},
	"product_rationale":     {},
	"technical_rationale":   {},
	"proposed_enhancements": {},
	"risk_considerations":   {},

	"title":                {},
	"core_features":        {},
	"mvp_scope":            {},
	"optional_extensions":  {},
	"recommended_stack":    {},
	"estimated_complexity": {},
}
