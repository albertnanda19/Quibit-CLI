package project

import (
	"hash/fnv"
	"strings"
)

type VariationAxis struct {
	Name    string
	Options []string
}

type IdeaVariationEngine struct {
	Axes []VariationAxis
}

func DefaultIdeaVariationEngine() IdeaVariationEngine {
	return IdeaVariationEngine{Axes: defaultVariationAxes()}
}

func (e IdeaVariationEngine) GenerateCandidates(base IdeaSpec, seed string, n int) []IdeaSpec {
	if n <= 0 {
		return []IdeaSpec{}
	}
	if len(e.Axes) == 0 {
		e = DefaultIdeaVariationEngine()
	}

	out := make([]IdeaSpec, 0, n)
	seen := map[string]struct{}{}

	base = base.Canonical()
	rootSeed := stableSeed(base.FingerprintString() + "|" + strings.TrimSpace(seed))

	for attempt := 0; len(out) < n && attempt < n*4; attempt++ {
		cand := e.Mutate(base, rootSeed, attempt)
		fp := cand.FingerprintString()
		if _, ok := seen[fp]; ok {
			continue
		}
		seen[fp] = struct{}{}
		out = append(out, cand)
	}

	return out
}

func (e IdeaVariationEngine) Mutate(base IdeaSpec, rootSeed uint32, attempt int) IdeaSpec {
	if len(e.Axes) == 0 {
		e = DefaultIdeaVariationEngine()
	}
	base = base.Canonical()

	idx := attempt
	axis := e.Axes[idx%len(e.Axes)]
	opt := pickOption(axis.Options, mix(rootSeed, uint32(idx)))

	out := base
	out.ArchitecturalAxis = composeAxis(base.ArchitecturalAxis, axis.Name, opt)
	out.ComplexityLevel = mutateComplexity(base.ComplexityLevel, mix(rootSeed, uint32(attempt+13)))
	return out.Canonical()
}

func defaultVariationAxes() []VariationAxis {
	return []VariationAxis{
		{
			Name: "architecture",
			Options: []string{
				"layered monolith with modular boundaries",
				"hexagonal architecture (ports/adapters)",
				"event-driven with async workers",
				"CQRS with read model projection",
				"plugin-based extensible core",
			},
		},
		{
			Name: "scale",
			Options: []string{
				"multi-tenant with tenant isolation",
				"high-throughput ingestion with backpressure",
				"offline-first sync and conflict resolution",
				"rate-limited public API with quotas",
				"low-latency real-time updates",
			},
		},
		{
			Name: "data_model",
			Options: []string{
				"append-only audit log with replay",
				"versioned entities with change history",
				"graph-like relationships and traversal",
				"time-series metrics/events storage",
				"document-style flexible schema with validation",
			},
		},
		{
			Name: "interaction_model",
			Options: []string{
				"webhook-driven integrations",
				"command queue with idempotency keys",
				"streaming updates (SSE/WebSocket)",
				"batch processing + scheduled jobs",
				"workflow engine with state machine",
			},
		},
	}
}

func composeAxis(existing string, axisName string, option string) string {
	existing = strings.TrimSpace(existing)
	axisName = normalizeScalar(axisName)
	option = strings.TrimSpace(option)
	if option == "" {
		return existing
	}
	if axisName == "" {
		axisName = "axis"
	}

	piece := axisName + ": " + option
	if existing == "" {
		return piece
	}

	parts := strings.Split(existing, " | ")
	out := make([]string, 0, len(parts)+1)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(normalizeScalar(p), axisName+":") {
			continue
		}
		out = append(out, p)
	}
	out = append(out, piece)
	return strings.Join(out, " | ")
}

func mutateComplexity(base string, s uint32) string {
	base = normalizeScalar(base)
	if base == "" {
		return base
	}

	levels := []string{"beginner", "intermediate", "advanced"}
	pos := -1
	for i := range levels {
		if levels[i] == base {
			pos = i
			break
		}
	}
	if pos < 0 {
		return base
	}

	switch s % 6 {
	case 0:
		if pos > 0 {
			return levels[pos-1]
		}
		return base
	case 1:
		if pos < len(levels)-1 {
			return levels[pos+1]
		}
		return base
	default:
		return base
	}
}

func pickOption(options []string, s uint32) string {
	if len(options) == 0 {
		return ""
	}
	idx := int(s % uint32(len(options)))
	return strings.TrimSpace(options[idx])
}

func stableSeed(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(s))))
	return h.Sum32()
}

func mix(a uint32, b uint32) uint32 {
	x := a ^ (b + 0x9e3779b9 + (a << 6) + (a >> 2))
	x ^= x << 13
	x ^= x >> 17
	x ^= x << 5
	return x
}
