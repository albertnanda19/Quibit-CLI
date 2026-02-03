package ai

import (
	"fmt"
	"regexp"
	"strings"
)

type qualityVerdict struct {
	decision qualityDecision
	hardFail bool
	reasons  []string
}

type qualityDecision string

const (
	qualityAccept     qualityDecision = "ACCEPT"
	qualityRefine     qualityDecision = "REFINE"
	qualityPivot      qualityDecision = "PIVOT"
	qualityRegenerate qualityDecision = "REGENERATE"
)

func (v qualityVerdict) ok() bool {
	return !v.hardFail && v.decision == qualityAccept
}

func (v qualityVerdict) summary() string {
	if len(v.reasons) == 0 {
		if v.ok() {
			return "ok"
		}
		if v.decision == "" {
			return "decision=UNKNOWN"
		}
		return fmt.Sprintf("decision=%s", v.decision)
	}
	if v.decision == "" {
		return strings.Join(v.reasons, "; ")
	}
	return fmt.Sprintf("decision=%s; %s", v.decision, strings.Join(v.reasons, "; "))
}

func evaluateIdeaQuality(idea ProjectIdea) qualityVerdict {

	allText := strings.ToLower(strings.TrimSpace(strings.Join([]string{
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

	interestingText := strings.ToLower(strings.TrimSpace(strings.Join([]string{
		idea.Project.ValueProp.WhyThisProjectIsInteresting,
		idea.Project.ValueProp.PortfolioValue,
		idea.Project.Tagline,
		idea.Project.Description.Summary,
		idea.Project.TechStack.Justification,
	}, " | ")))

	if looksCliche(allText) && !hasExtremeTwist(allText) {
		return qualityVerdict{
			hardFail: true,
			decision: qualityRegenerate,
			reasons:  []string{"anti-generic FAIL: cliché category without an extreme technical twist"},
		}
	}
	if looksLikeClone(allText) {
		return qualityVerdict{
			hardFail: true,
			decision: qualityRegenerate,
			reasons:  []string{"anti-generic FAIL: clone framing (\"X clone\" / \"like X\")"},
		}
	}
	if looksLikeCRUD(allText) && !hasTechnicalDepthSignals(allText) && !hasConstraintSignals(allText) && !hasExtremeTwist(allText) {
		return qualityVerdict{
			hardFail: true,
			decision: qualityRegenerate,
			reasons:  []string{"anti-generic FAIL: CRUD-y scope with no depth/constraints/twist"},
		}
	}

	technicalDepth := hasTechnicalDepthSignals(allText) || hasExtremeTwist(allText)
	nonTrivialConstraint := hasConstraintSignals(allText) || hasExtremeTwist(allText)
	tradeoffs := hasTradeoffSignals(allText)
	if !technicalDepth {
		return qualityVerdict{
			hardFail: true,
			decision: qualityRegenerate,
			reasons:  []string{"technical depth FAIL: no concrete engineering depth signals (reads like a thin app idea)"},
		}
	}

	diffOK := hasDifferentiationSignals(interestingText) || (hasExtremeTwist(allText) && hasTechnicalDepthSignals(allText))
	if !diffOK {
		return qualityVerdict{
			decision: qualityPivot,
			reasons:  []string{"differentiation FAIL: no clear unique core differentiator (would not stop a reviewer from scrolling)"},
		}
	}

	scopeOK, scopeReason := scopeRealismCheck(idea)
	if !scopeOK {
		return qualityVerdict{
			decision: qualityRefine,
			reasons:  []string{"scope/realism FAIL: " + scopeReason},
		}
	}

	interviewable := hasInterviewSignals(allText)
	if !interviewable {
		return qualityVerdict{
			decision: qualityRefine,
			reasons:  []string{"portfolio worthiness FAIL: not clearly interviewable (missing architecture/system-design cues)"},
		}
	}
	if !nonTrivialConstraint || !tradeoffs {
		missing := []string{}
		if !nonTrivialConstraint {
			missing = append(missing, "non-trivial constraint")
		}
		if !tradeoffs {
			missing = append(missing, "explicit trade-off")
		}
		return qualityVerdict{
			decision: qualityRefine,
			reasons:  []string{"technical depth incomplete: missing " + strings.Join(missing, " + ")},
		}
	}

	return qualityVerdict{decision: qualityAccept}
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
		"expense tracker",
		"personal finance",
		"pomodoro",
		"notes app",
		"note-taking",
		"recipe app",
		"movie tracker",
	}
	for _, n := range needles {
		if strings.Contains(text, n) {
			return true
		}
	}
	return false
}

func hasExtremeTwist(text string) bool {
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
	crudish := containsAny(text,
		"add/edit/delete", "add, edit, delete",
		"manage users", "manage items", "admin panel", "admin dashboard",
		"login", "sign in", "sign-up", "register", "authentication",
		"dashboard", "profile page", "settings page",
	)
	return crudish && !hasTechnicalDepthSignals(text) && !hasConstraintSignals(text)
}

func looksLikeClone(text string) bool {
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
		"crdt", "offline-first", "local-first",
		"vector", "embedding", "retrieval", "rag",
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

func hasDifferentiationSignals(text string) bool {
	if hasExtremeTwist(text) {
		return true
	}
	return containsAny(text,
		"tamper-evident", "append-only log", "merkle",
		"deterministic replay",
		"threat model",
		"policy engine", "rego", "opa",
		"crdt", "local-first", "offline-first",
		"zero-knowledge", "zk", "end-to-end encryption", "e2ee",
		"differential privacy", "privacy budget",
		"backpressure", "outbox", "saga", "idempotency",
		"vector index", "inverted index",
	)
}

func scopeRealismCheck(idea ProjectIdea) (bool, string) {
	must := idea.Project.MVP.MustHave
	nice := idea.Project.MVP.NiceToHave
	out := idea.Project.MVP.OutOfScope

	if len(must) >= 8 {
		return false, "MVP must-have list is too large (>=8) for a solo MVP"
	}

	bigRockCount := 0
	all := strings.ToLower(strings.Join(must, " | "))
	bigRocks := []string{
		"payments", "subscription", "billing",
		"marketplace",
		"recommendation", "ranking",
		"real-time chat", "messaging",
		"social feed",
		"multi-tenant",
		"admin dashboard", "admin panel",
		"ml training", "train model",
	}
	for _, k := range bigRocks {
		if strings.Contains(all, k) {
			bigRockCount++
		}
	}
	if bigRockCount >= 3 {
		return false, "too many big-scope features packed into MVP (payments/chat/recommendations/multi-tenant/etc.)"
	}

	fluff := func(items []string) bool {
		if len(items) == 0 {
			return true
		}
		joined := strings.ToLower(strings.Join(items, " | "))
		return containsAny(joined, "etc", "more features", "improvements", "enhancements", "tbd") && len(items) <= 2
	}
	if fluff(nice) || fluff(out) {
		return false, "scope lists are too vague (nice-to-have/out-of-scope read like placeholders)"
	}
	return true, ""
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
