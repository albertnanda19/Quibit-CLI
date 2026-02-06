package ai

import (
	"encoding/json"
	"errors"
	"strings"
)

var (
	ErrGenerateProjectIdeaEmptyOutput        = errors.New("generate project idea: empty output")
	ErrGenerateProjectIdeaInvalidSchema      = errors.New("generate project idea: invalid schema")
	ErrGenerateProjectIdeaMissingOverview    = errors.New("generate project idea: missing overview")
	ErrGenerateProjectIdeaMissingTechStack   = errors.New("generate project idea: missing tech stack")
	ErrGenerateProjectIdeaMissingMVPScope    = errors.New("generate project idea: missing mvp scope")
	ErrGenerateProjectIdeaMissingLearning    = errors.New("generate project idea: missing learning outcomes")
)

func NormalizeGenerateProjectIdeaResponse(raw string) (GenerateProjectIdeaResponse, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return GenerateProjectIdeaResponse{}, ErrGenerateProjectIdeaEmptyOutput
	}

	if out, ok := normalizeGenerateProjectIdeaFromJSON(raw); ok {
		return out, validateGenerateProjectIdeaResponse(out)
	}

	out := normalizeGenerateProjectIdeaFromSections(raw)
	return out, validateGenerateProjectIdeaResponse(out)
}

func normalizeGenerateProjectIdeaFromJSON(raw string) (GenerateProjectIdeaResponse, bool) {
	candidate := extractJSONObject(raw)
	if candidate == "" {
		return GenerateProjectIdeaResponse{}, false
	}

	var out GenerateProjectIdeaResponse
	if err := json.Unmarshal([]byte(candidate), &out); err != nil {
		return GenerateProjectIdeaResponse{}, false
	}
	return sanitizeGenerateProjectIdeaResponse(out), true
}

func normalizeGenerateProjectIdeaFromSections(raw string) GenerateProjectIdeaResponse {
	text := stripCodeFences(raw)
	lines := splitLines(text)

	sections := splitByHeadings(lines)

	ov := ProjectOverview{}
	if s, ok := sections["overview"]; ok {
		ov.ProjectName = pickValueLine(s, []string{"project name", "name", "project"})
		ov.Tagline = pickValueLine(s, []string{"tagline"})
		ov.Problem = pickMultilineValue(s, []string{"problem"})
		ov.TargetUsers = pickList(s, []string{"target users", "users", "audience"})
		ov.SuccessMetrics = pickList(s, []string{"success metrics", "metrics", "success criteria", "kpis"})
	}

	ts := TechStack{}
	if s, ok := sections["tech_stack"]; ok {
		ts.Backend = pickValueLine(s, []string{"backend"})
		ts.Frontend = pickValueLine(s, []string{"frontend"})
		ts.Database = pickValueLine(s, []string{"database", "db"})
		ts.Infra = pickValueLine(s, []string{"infra", "infrastructure"})
		ts.Justification = pickMultilineValue(s, []string{"justification", "rationale", "reasoning"})
	}

	mvp := MVPScope{}
	if s, ok := sections["mvp_scope"]; ok {
		mvp.Goal = pickMultilineValue(s, []string{"goal"})
		mvp.MustHaveFeatures = pickList(s, []string{"must have", "must-have", "must have features", "core features", "features"})
		mvp.OutOfScope = pickList(s, []string{"out of scope", "out-of-scope", "excluded"})
	}

	learning := []string{}
	if s, ok := sections["learning_outcomes"]; ok {
		learning = pickList(s, []string{"learning outcomes", "outcomes", "learning"})
		if len(learning) == 0 {
			learning = pickBareList(s)
		}
	}

	out := GenerateProjectIdeaResponse{
		Overview:         ov,
		TechStack:        ts,
		MVPScope:         mvp,
		LearningOutcomes: learning,
	}
	return sanitizeGenerateProjectIdeaResponse(out)
}

func validateGenerateProjectIdeaResponse(in GenerateProjectIdeaResponse) error {
	in = sanitizeGenerateProjectIdeaResponse(in)

	if strings.TrimSpace(in.Overview.ProjectName) == "" && strings.TrimSpace(in.Overview.Problem) == "" {
		return ErrGenerateProjectIdeaMissingOverview
	}

	if strings.TrimSpace(in.TechStack.Backend) == "" && strings.TrimSpace(in.TechStack.Frontend) == "" && strings.TrimSpace(in.TechStack.Database) == "" {
		return ErrGenerateProjectIdeaMissingTechStack
	}

	if strings.TrimSpace(in.MVPScope.Goal) == "" && len(in.MVPScope.MustHaveFeatures) == 0 {
		return ErrGenerateProjectIdeaMissingMVPScope
	}

	if len(in.LearningOutcomes) == 0 {
		return ErrGenerateProjectIdeaMissingLearning
	}

	return nil
}

func sanitizeGenerateProjectIdeaResponse(in GenerateProjectIdeaResponse) GenerateProjectIdeaResponse {
	in.Overview.ProjectName = cleanText(in.Overview.ProjectName)
	in.Overview.Tagline = cleanText(in.Overview.Tagline)
	in.Overview.Problem = cleanText(in.Overview.Problem)
	in.Overview.TargetUsers = cleanList(in.Overview.TargetUsers)
	in.Overview.SuccessMetrics = cleanList(in.Overview.SuccessMetrics)

	in.TechStack.Backend = cleanText(in.TechStack.Backend)
	in.TechStack.Frontend = cleanText(in.TechStack.Frontend)
	in.TechStack.Database = cleanText(in.TechStack.Database)
	in.TechStack.Infra = cleanText(in.TechStack.Infra)
	in.TechStack.Justification = cleanText(in.TechStack.Justification)

	in.MVPScope.Goal = cleanText(in.MVPScope.Goal)
	in.MVPScope.MustHaveFeatures = cleanList(in.MVPScope.MustHaveFeatures)
	in.MVPScope.OutOfScope = cleanList(in.MVPScope.OutOfScope)

	in.LearningOutcomes = cleanList(in.LearningOutcomes)

	return in
}

func stripCodeFences(s string) string {
	s = strings.ReplaceAll(s, "```json", "```")
	parts := strings.Split(s, "```")
	if len(parts) < 3 {
		return s
	}

	best := ""
	for i := 1; i < len(parts); i += 2 {
		chunk := strings.TrimSpace(parts[i])
		if chunk == "" {
			continue
		}
		if len(chunk) > len(best) {
			best = chunk
		}
	}
	if best != "" {
		return best
	}
	return s
}

func extractJSONObject(raw string) string {
	raw = stripCodeFences(raw)
	r := strings.TrimSpace(raw)

	start := strings.IndexByte(r, '{')
	if start < 0 {
		return ""
	}

	depth := 0
	inString := false
	escape := false
	for i := start; i < len(r); i++ {
		c := r[i]
		if inString {
			if escape {
				escape = false
				continue
			}
			switch c {
			case '\\':
				escape = true
			case '"':
				inString = false
			}
			continue
		}

		switch c {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return strings.TrimSpace(r[start : i+1])
			}
		}
	}

	return ""
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.Split(s, "\n")
}

type sectionMap map[string][]string

func splitByHeadings(lines []string) sectionMap {
	out := sectionMap{}
	cur := ""

	for _, ln := range lines {
		line := strings.TrimSpace(ln)
		if line == "" {
			continue
		}

		key := canonicalHeading(line)
		if key != "" {
			cur = key
			if _, ok := out[cur]; !ok {
				out[cur] = []string{}
			}
			continue
		}

		if cur == "" {
			continue
		}
		out[cur] = append(out[cur], line)
	}

	return out
}

func canonicalHeading(line string) string {
	l := strings.TrimSpace(line)
	l = strings.Trim(l, "#*:- ")
	l = strings.ToLower(l)
	l = strings.ReplaceAll(l, "_", " ")
	l = strings.ReplaceAll(l, "-", " ")
	for strings.Contains(l, "  ") {
		l = strings.ReplaceAll(l, "  ", " ")
	}

	switch l {
	case "overview", "project overview", "project" :
		return "overview"
	case "tech stack", "technology", "stack", "recommended tech stack", "recommended technology", "recommended tech" :
		return "tech_stack"
	case "mvp", "mvp scope", "scope", "mvp plan" :
		return "mvp_scope"
	case "learning outcomes", "learning", "outcomes" :
		return "learning_outcomes"
	default:
		return ""
	}
}

func pickValueLine(lines []string, keys []string) string {
	for _, ln := range lines {
		k, v, ok := splitKeyValue(ln)
		if !ok {
			continue
		}
		if matchKey(k, keys) {
			return v
		}
	}
	return ""
}

func pickMultilineValue(lines []string, keys []string) string {
	var b strings.Builder
	capture := false

	for _, ln := range lines {
		k, v, ok := splitKeyValue(ln)
		if ok {
			if matchKey(k, keys) {
				capture = true
				if v != "" {
					if b.Len() > 0 {
						b.WriteByte(' ')
					}
					b.WriteString(v)
				}
				continue
			}
			if capture {
				break
			}
			continue
		}

		if capture {
			if isBulletLine(ln) {
				val := trimBullet(ln)
				if val == "" {
					continue
				}
				if b.Len() > 0 {
					b.WriteByte(' ')
				}
				b.WriteString(val)
				continue
			}

			if b.Len() > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(strings.TrimSpace(ln))
		}
	}

	return b.String()
}

func pickList(lines []string, keys []string) []string {
	capture := false
	out := []string{}

	for _, ln := range lines {
		k, v, ok := splitKeyValue(ln)
		if ok {
			if matchKey(k, keys) {
				capture = true
				if v != "" {
					out = append(out, splitInlineList(v)...)
				}
				continue
			}
			if capture {
				break
			}
			continue
		}

		if capture {
			if isBulletLine(ln) {
				val := trimBullet(ln)
				if val != "" {
					out = append(out, val)
				}
			}
		}
	}

	return out
}

func pickBareList(lines []string) []string {
	out := []string{}
	for _, ln := range lines {
		if isBulletLine(ln) {
			val := trimBullet(ln)
			if val != "" {
				out = append(out, val)
			}
		}
	}
	return out
}

func splitKeyValue(line string) (string, string, bool) {
	l := strings.TrimSpace(line)
	l = strings.TrimLeft(l, "#*-")
	l = strings.TrimSpace(l)

	idx := strings.Index(l, ":")
	if idx < 0 {
		idx = strings.Index(l, "-")
		if idx < 0 {
			return "", "", false
		}
	}

	k := strings.TrimSpace(l[:idx])
	v := strings.TrimSpace(l[idx+1:])
	if k == "" {
		return "", "", false
	}
	return k, v, true
}

func matchKey(k string, candidates []string) bool {
	k = normalizeKey(k)
	for _, c := range candidates {
		if k == normalizeKey(c) {
			return true
		}
	}
	return false
}

func normalizeKey(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}

func isBulletLine(line string) bool {
	l := strings.TrimSpace(line)
	return strings.HasPrefix(l, "-") || strings.HasPrefix(l, "*")
}

func trimBullet(line string) string {
	l := strings.TrimSpace(line)
	l = strings.TrimPrefix(l, "-")
	l = strings.TrimPrefix(l, "*")
	l = strings.TrimSpace(l)
	return strings.Trim(l, "\"'")
}

func splitInlineList(s string) []string {
	if s == "" {
		return nil
	}

	sep := ","
	if strings.Contains(s, ";") {
		sep = ";"
	}
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "[]")
		p = strings.TrimSpace(p)
		p = strings.Trim(p, "\"'")
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func cleanText(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"'")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return strings.TrimSpace(s)
}

func cleanList(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, v := range in {
		v = cleanText(v)
		if v == "" {
			continue
		}
		k := strings.ToLower(v)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, v)
	}
	return out
}
