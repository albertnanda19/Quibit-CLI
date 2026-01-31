package project

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

func HashContent(overview string, mvpScope []string, techStack []string, complexity string, duration string) string {
	norm := func(s string) string {
		s = strings.TrimSpace(strings.ToLower(s))
		return s
	}

	mvp := make([]string, 0, len(mvpScope))
	for _, v := range mvpScope {
		v = norm(v)
		if v == "" {
			continue
		}
		mvp = append(mvp, v)
	}
	sort.Strings(mvp)

	tech := make([]string, 0, len(techStack))
	for _, v := range techStack {
		v = norm(v)
		if v == "" {
			continue
		}
		tech = append(tech, v)
	}
	sort.Strings(tech)

	parts := []string{
		norm(overview),
		strings.Join(mvp, ","),
		strings.Join(tech, ","),
		norm(complexity),
		norm(duration),
	}

	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}
