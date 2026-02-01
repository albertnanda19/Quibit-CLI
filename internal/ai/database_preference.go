package ai

import (
	"encoding/json"
	"strings"
)

func databasePreferenceLine(dbs []string) string {
	dbs = normalizeDBList(dbs)
	if len(dbs) == 0 {
		return ""
	}
	if len(dbs) == 1 {
		if dbs[0] == "none" {
			return "- database_preference: none\n"
		}
		return "- database_preference: " + dbs[0] + "\n"
	}
	b, err := json.Marshal(dbs)
	if err != nil {
		// fallback: comma-separated
		return "- database_preference: " + strings.Join(dbs, ", ") + "\n"
	}
	return "- database_preference: " + string(b) + "\n"
}

func normalizeDBList(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		// default to none if caller didn't specify anything
		return []string{"none"}
	}
	// If mixed with "none", drop "none".
	hasNone := false
	for _, v := range out {
		if v == "none" {
			hasNone = true
			break
		}
	}
	if hasNone && len(out) > 1 {
		filtered := make([]string, 0, len(out)-1)
		for _, v := range out {
			if v == "none" {
				continue
			}
			filtered = append(filtered, v)
		}
		out = filtered
	}
	// De-dup while preserving order.
	seen := map[string]struct{}{}
	dedup := make([]string, 0, len(out))
	for _, v := range out {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		dedup = append(dedup, v)
	}
	return dedup
}
