package project

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

type ProjectDNA struct {
	AppType            string
	PrimaryDomain      string
	CoreTechStack      []string
	ArchitecturalStyle string
	ComplexityLevel    string
}

func (d ProjectDNA) Canonical() ProjectDNA {
	out := ProjectDNA{}
	out.AppType = normalizeScalar(d.AppType)
	out.PrimaryDomain = normalizeScalar(d.PrimaryDomain)
	out.ArchitecturalStyle = normalizeScalar(d.ArchitecturalStyle)
	out.ComplexityLevel = normalizeScalar(d.ComplexityLevel)
	out.CoreTechStack = normalizeStringList(d.CoreTechStack)
	return out
}

func (d ProjectDNA) FingerprintString() string {
	c := d.Canonical()
	parts := []string{
		"app_type=" + c.AppType,
		"primary_domain=" + c.PrimaryDomain,
		"core_tech_stack=" + strings.Join(c.CoreTechStack, ","),
		"architectural_style=" + c.ArchitecturalStyle,
		"complexity_level=" + c.ComplexityLevel,
	}
	return strings.Join(parts, "|")
}

func (d ProjectDNA) FingerprintHash() string {
	sum := sha256.Sum256([]byte(d.FingerprintString()))
	return hex.EncodeToString(sum[:])
}

func normalizeScalar(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}

func normalizeStringList(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = normalizeScalar(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	if len(out) < 2 {
		return out
	}
	sort.Strings(out)
	return out
}
