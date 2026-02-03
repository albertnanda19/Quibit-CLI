package ai

import (
	"fmt"
	"regexp"
	"strings"
)

type qualityVerdict struct {
	passes   int
	hardFail bool
	reasons  []string
}

func (v qualityVerdict) ok() bool {
	// Quality bar: pass at least 3 criteria, and no hard-fail.
	return !v.hardFail && v.passes >= 3
}

func (v qualityVerdict) summary() string {
	if len(v.reasons) == 0 {
		if v.ok() {
			return "ok"
		}
		return fmt.Sprintf("passes=%d", v.passes)
	}
	return fmt.Sprintf("passes=%d; %s", v.passes, strings.Join(v.reasons, "; "))
}

func evaluateIdeaQuality(idea ProjectIdea) qualityVerdict {
	text := strings.ToLower(strings.TrimSpace(strings.Join([]string{
		idea.Project.Name,
		idea.Project.Tagline,
		idea.Project.Description.Summary,
		idea.Project.Description.DetailedExplanation,
		idea.Project.Problem.Problem,
		idea.Project.Problem.WhyItMatters,
		idea.Project.Problem.CurrentSolutionsAndGaps,
		strings.Join(idea.Project.ValueProp.KeyBenefits, " "),
		idea.Project.ValueProp.WhyThisProjectIsInteresting,
		idea.Project.ValueProp.PortfolioValue,
		idea.Project.MVP.Goal,
		strings.Join(idea.Project.MVP.MustHave, " "),
		strings.Join(idea.Project.MVP.NiceToHave, " "),
		strings.Join(idea.Project.MVP.OutOfScope, " "),
		idea.Project.TechStack.Justification,
		strings.Join(idea.Project.Future, " "),
		strings.Join(idea.Project.Learning, " "),
	}, " | ")))

	// Hard-fail clichés unless there is a clear extreme twist.
	if looksCliche(text) && !hasExtremeTwist(text) {
		return qualityVerdict{
			passes:   0,
			hardFail: true,
			reasons:  []string{"cliche/generic category (todo/chat/blog/ecommerce/weather/url shortener) without an extreme technical twist"},
		}
	}

	criteria := map[string]bool{
		"not_crud_generic": !looksLikeCRUD(text),
		"not_clone":        !looksLikeClone(text),
		"technical_depth":  hasTechnicalDepthSignals(text),
		"tradeoffs":        hasTradeoffSignals(text),
		"interviewable":    hasInterviewSignals(text),
		"scalable_or_pivot": hasScalabilitySignals(text),
		"non_trivial_constraints": hasConstraintSignals(text),
	}

	var passes int
	var missing []string
	for k, ok := range criteria {
		if ok {
			passes++
		} else {
			missing = append(missing, k)
		}
	}

	v := qualityVerdict{passes: passes}
	// Extra enforcement: passing "not CRUD" + "not clone" alone is not enough.
	// We want at least one strong technical signal (depth/constraint/trade-off).
	strongSignal := criteria["technical_depth"] || criteria["non_trivial_constraints"] || criteria["tradeoffs"]
	if passes >= 3 && !strongSignal {
		v.reasons = []string{"quality bar not met: missing strong signals (technical_depth or non_trivial_constraints or tradeoffs)"}
		return v
	}
	if passes < 3 {
		// Keep reasons short, but actionable.
		v.reasons = []string{"quality bar not met (need >=3 criteria): missing " + strings.Join(missing, ", ")}
	}
	return v
}

func looksCliche(text string) bool {
	needles := []string{
		"todo",
		"to-do",
		"habit tracker",
		"weather app",
		"url shortener",
		"shorten url",
		"blog platform",
		"e-commerce",
		"ecommerce",
		"shopping cart",
		"chat app",
	}
	for _, n := range needles {
		if strings.Contains(text, n) {
			return true
		}
	}
	return false
}

func hasExtremeTwist(text string) bool {
	// Allow clichés only when clearly constrained or technically special.
	return containsAny(text,
		"end-to-end encryption", "e2ee", "zero-knowledge",
		"differential privacy", "privacy budget",
		"crdt", "offline-first", "local-first", "conflict-free",
		"federated", "matrix protocol", "activitypub",
		"formal verification", "model checking",
		"deterministic replay", "tamper-evident", "append-only log",
		"real-time", "backpressure", "streaming",
	)
}

func looksLikeCRUD(text string) bool {
	if containsAny(text, "crud", "create read update delete", "create, read, update, delete") {
		return true
	}
	// Heuristic: lots of "manage/add/edit/delete" language and no depth signals.
	crudish := containsAny(text, "add/edit/delete", "add, edit, delete", "manage users", "manage items", "admin panel")
	return crudish && !hasTechnicalDepthSignals(text) && !hasConstraintSignals(text)
}

func looksLikeClone(text string) bool {
	// "X clone" or "like X" patterns.
	if containsAny(text, " clone", "like trello", "like notion", "like spotify", "like netflix", "like uber") {
		return true
	}
	cloneRe := regexp.MustCompile(`\b(clone of|a clone of|like\s+(notion|trello|spotify|netflix|uber|airbnb|twitter|instagram))\b`)
	return cloneRe.MatchString(text)
}

func hasTechnicalDepthSignals(text string) bool {
	return containsAny(text,
		"event-driven", "queue", "job queue", "streaming", "pub/sub",
		"idempotency", "dedup", "outbox", "saga",
		"rate limit", "backpressure",
		"observability", "tracing", "opentelemetry", "slo",
		"multi-tenant", "rbac", "abac", "audit log",
		"encryption", "key management", "kms",
		"indexing", "inverted index", "search ranking",
		"caching", "cache invalidation",
		"consistency", "distributed", "replication",
	)
}

func hasTradeoffSignals(text string) bool {
	return containsAny(text,
		"trade-off", "tradeoff", "vs.", " vs ",
		"latency vs", "cost vs", "consistency vs", "availability vs",
		"privacy vs", "accuracy vs", "throughput vs",
		"choose", "we choose", "we decided",
	)
}

func hasConstraintSignals(text string) bool {
	return containsAny(text,
		"performance", "latency", "throughput", "p99",
		"privacy", "pii", "gdpr", "hipaa",
		"reliability", "resilience", "fault", "retry", "circuit breaker",
		"offline", "low bandwidth",
		"security", "threat model", "abuse", "rate limiting",
		"dx", "developer experience", "schema enforcement",
	)
}

func hasScalabilitySignals(text string) bool {
	return containsAny(text,
		"multi-tenant", "shard", "sharding", "partition",
		"horizontal scale", "autoscaling",
		"plugin system", "extensions", "marketplace",
		"team plan", "enterprise", "organization",
	)
}

func hasInterviewSignals(text string) bool {
	return containsAny(text,
		"architecture", "system design", "data model",
		"consistency", "availability", "idempotency",
		"queue", "caching", "observability", "slo",
	)
}

func containsAny(text string, needles ...string) bool {
	for _, n := range needles {
		if n == "" {
			continue
		}
		if strings.Contains(text, strings.ToLower(n)) {
			return true
		}
	}
	return false
}

