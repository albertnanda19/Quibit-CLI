package similarity

import (
	"strings"
	"unicode"
)

func tokenizeText(s string) []string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return []string{}
	}

	seen := map[string]struct{}{}
	out := make([]string, 0)

	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		tok := b.String()
		b.Reset()
		if len(tok) < 3 {
			return
		}
		if _, ok := seen[tok]; ok {
			return
		}
		seen[tok] = struct{}{}
		out = append(out, tok)
	}

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			continue
		}
		flush()
	}
	flush()

	return out
}

func normalizeList(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(items))
	for _, v := range items {
		v = strings.TrimSpace(strings.ToLower(v))
		if v == "" {
			continue
		}
		if len(v) < 3 {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}

	sortStrings(out)
	return out
}

func normalizeTechStack(s string) []string {
	return normalizeList(tokenizeText(s))
}

func sortStrings(a []string) {
	if len(a) < 2 {
		return
	}
	for i := 0; i < len(a)-1; i++ {
		for j := i + 1; j < len(a); j++ {
			if a[j] < a[i] {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
}
