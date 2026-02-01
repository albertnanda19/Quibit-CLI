package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"quibit/internal/ai"
	"quibit/internal/db"
	"quibit/internal/model"
	pmodels "quibit/internal/persistence/models"
	"quibit/internal/project"
	"quibit/internal/tui"
	tuiinput "quibit/internal/tui/input"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new portfolio project idea.",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		ctx := cmd.Context()
		for {
			options := []tui.Option{
				{ID: "new", Label: "Generate New Project"},
				{ID: "continue", Label: "Continue Existing Project"},
				{ID: "view", Label: "View Saved Projects"},
				{ID: "exit", Label: "Exit"},
			}

			selection, err := tui.SelectOption(os.Stdin, out, "Select mode:", options)
			if err != nil {
				return fmt.Errorf("generate: %w", err)
			}

			switch selection.ID {
			case "new":
				if err := runGenerateNew(ctx, os.Stdin, out); err != nil {
					return err
				}
			case "continue":
				if err := runContinueExisting(ctx, os.Stdin, out); err != nil {
					return err
				}
			case "view":
				if err := runViewSavedProjects(ctx, out); err != nil {
					return err
				}
			case "exit":
				return nil
			default:
				return fmt.Errorf("generate: invalid selection")
			}
		}
	},
}

func runGenerateNew(ctx context.Context, in *os.File, out io.Writer) error {
	input, err := tuiinput.CollectNewProjectInput(in, out)
	if err != nil {
		return err
	}
	var pendingReason *ai.RetryReason
	var pendingStrategy ai.PivotStrategy
	var lastReasonUsed *ai.RetryReason
	var lastMeta ai.AIResult
	var diversityRef *project.Snapshot
	diversityAttempts := 0
generateLoop:
	for {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Generating idea with AI")

		var idea ai.ProjectIdea
		var rawJSON string
		if pendingReason == nil {
			lastReasonUsed = nil
			idea, rawJSON, lastMeta, err = ai.GenerateProjectIdeaWithMeta(ctx, input)
		} else {
			lastReasonUsed = pendingReason
			idea, rawJSON, lastMeta, err = ai.GenerateProjectIdeaWithPivotMeta(ctx, input, *pendingReason, pendingStrategy)
			pendingReason = nil
		}
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		// In-session diversity guard: when user regenerates (or generates again with same inputs),
		// avoid returning an idea that's basically identical except the name.
		if diversityRef != nil {
			current := makeSnapshotFromIdea(idea, input)
			score := project.JaccardSimilarity(current, *diversityRef)
			decision := project.DecideSimilarity(score)
			if decision != project.SimilarityOK && diversityAttempts < 3 {
				diversityAttempts++
				pendingReason = ptrRetry(ai.RetrySimilarityTooHigh)
				pendingStrategy = rotatePivotStrategy(diversityAttempts)
				continue
			}
			// stop enforcing once we got something distinct enough (or we hit attempt cap)
			diversityRef = nil
			diversityAttempts = 0
		}

		action, bestScore, err := evaluateSimilarity(ctx, idea, input)
		if err != nil {
			return err
		}
		switch action {
		case project.SimilarityRegenerate:
			fmt.Fprintf(out, "\nSimilarity score %.2f detected. Regenerating...\n", bestScore)
			pendingReason = ptrRetry(ai.RetrySimilarityTooHigh)
			pendingStrategy = selectPivotStrategy(ai.RetrySimilarityTooHigh)
			continue
		case project.SimilarityBlock:
			fmt.Fprintf(out, "\nSimilarity score %.2f is too high. Generation blocked.\n", bestScore)
			return nil
		default:
		}

		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Project")
		fmt.Fprintf(out, "%s — %s\n", idea.Project.Name, idea.Project.Tagline)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Summary")
		fmt.Fprintln(out, idea.Project.Description.Summary)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Detailed Explanation")
		fmt.Fprintln(out, idea.Project.Description.DetailedExplanation)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Problem Statement")
		fmt.Fprintln(out, idea.Project.Problem.Problem)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Why It Matters")
		fmt.Fprintln(out, idea.Project.Problem.WhyItMatters)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Current Solutions and Gaps")
		fmt.Fprintln(out, idea.Project.Problem.CurrentSolutionsAndGaps)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Target Users (Primary)")
		for _, item := range idea.Project.TargetUsers.Primary {
			fmt.Fprintf(out, "- %s\n", item)
		}
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Target Users (Secondary)")
		for _, item := range idea.Project.TargetUsers.Secondary {
			fmt.Fprintf(out, "- %s\n", item)
		}
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Use Cases")
		for _, item := range idea.Project.TargetUsers.UseCases {
			fmt.Fprintf(out, "- %s\n", item)
		}
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Value Proposition")
		for _, item := range idea.Project.ValueProp.KeyBenefits {
			fmt.Fprintf(out, "- %s\n", item)
		}
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Why This Project Is Interesting")
		fmt.Fprintln(out, idea.Project.ValueProp.WhyThisProjectIsInteresting)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Portfolio Value")
		fmt.Fprintln(out, idea.Project.ValueProp.PortfolioValue)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "MVP Goal")
		fmt.Fprintln(out, idea.Project.MVP.Goal)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "MVP Must-Have Features")
		for _, item := range idea.Project.MVP.MustHave {
			fmt.Fprintf(out, "- %s\n", item)
		}
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "MVP Nice-to-Have Features")
		for _, item := range idea.Project.MVP.NiceToHave {
			fmt.Fprintf(out, "- %s\n", item)
		}
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Out of Scope")
		for _, item := range idea.Project.MVP.OutOfScope {
			fmt.Fprintf(out, "- %s\n", item)
		}
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Recommended Tech Stack")
		fmt.Fprintf(out, "- Backend: %s\n", idea.Project.TechStack.Backend)
		fmt.Fprintf(out, "- Frontend: %s\n", idea.Project.TechStack.Frontend)
		fmt.Fprintf(out, "- Database: %s\n", idea.Project.TechStack.Database)
		fmt.Fprintf(out, "- Infra: %s\n", idea.Project.TechStack.Infra)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Tech Stack Justification")
		fmt.Fprintln(out, idea.Project.TechStack.Justification)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Complexity")
		fmt.Fprintln(out, idea.Project.Complexity)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Estimated Duration")
		fmt.Fprintln(out, idea.Project.Duration.Range)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Duration Assumptions")
		fmt.Fprintln(out, idea.Project.Duration.Assumptions)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Future Extensions")
		for _, item := range idea.Project.Future {
			fmt.Fprintf(out, "- %s\n", item)
		}
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Learning Outcomes")
		for _, item := range idea.Project.Learning {
			fmt.Fprintf(out, "- %s\n", item)
		}

		selection, err := tui.SelectOption(in, out, "Next action:", []tui.Option{
			{ID: "accept", Label: "Accept"},
			{ID: "regenerate", Label: "Regenerate"},
			{ID: "regenerate_harder", Label: "Regenerate (higher complexity)"},
			{ID: "back", Label: "Back"},
		})
		if err != nil {
			return err
		}

		switch selection.ID {
		case "accept":
			if err := saveGeneratedProject(ctx, input, idea, rawJSON, lastMeta, lastReasonUsed); err != nil {
				if errors.Is(err, errDuplicateDNA) {
					fmt.Fprintln(out, "\nDuplicate DNA detected. Regenerating...")
					pendingReason = ptrRetry(ai.RetryDuplicateDNA)
					pendingStrategy = selectPivotStrategy(ai.RetryDuplicateDNA)
					continue
				}
				return err
			}
			fmt.Fprintln(out, "")
			fmt.Fprintln(out, "Project saved.")
			for {
				after, _ := tui.SelectOption(in, out, "Next action:", []tui.Option{
					{ID: "copy", Label: "Copy output to clipboard"},
					{ID: "same", Label: "Generate another project (same inputs)"},
					{ID: "same_harder", Label: "Generate another project (higher complexity)"},
					{ID: "same_easier", Label: "Generate another project (lower complexity)"},
					{ID: "back", Label: "Back"},
				})
				switch after.ID {
				case "copy":
					var buf bytes.Buffer
					printIdea(&buf, idea)
					fmt.Fprintln(&buf, "")
					fmt.Fprintln(&buf, "----")
					fmt.Fprintln(&buf, "Raw JSON")
					fmt.Fprintln(&buf, rawJSON)
					if err := tui.CopyToClipboard(out, buf.String()); err != nil {
						fmt.Fprintf(out, "Copy failed: %v\n", err)
						continue
					}
					fmt.Fprintln(out, "Copied to clipboard.")
				case "same":
					// Keep the same input; just run generation again.
					diversityRef = ptrSnapshot(makeSnapshotFromIdea(idea, input))
					diversityAttempts = 0
					pendingReason = nil
					lastReasonUsed = nil
					continue generateLoop
				case "same_harder":
					// Keep the same input; increase complexity then generate again.
					input.Complexity = bumpComplexity(input.Complexity)
					diversityRef = ptrSnapshot(makeSnapshotFromIdea(idea, input))
					diversityAttempts = 0
					pendingReason = nil
					lastReasonUsed = nil
					continue generateLoop
				case "same_easier":
					// Keep the same input; decrease complexity then generate again.
					input.Complexity = lowerComplexity(input.Complexity)
					diversityRef = ptrSnapshot(makeSnapshotFromIdea(idea, input))
					diversityAttempts = 0
					pendingReason = nil
					lastReasonUsed = nil
					continue generateLoop
				default:
					return nil
				}
			}
			return nil
		case "regenerate":
			diversityRef = ptrSnapshot(makeSnapshotFromIdea(idea, input))
			diversityAttempts = 0
			pendingReason = ptrRetry(ai.RetryUserRejected)
			pendingStrategy = selectPivotStrategy(ai.RetryUserRejected)
			continue
		case "regenerate_harder":
			input.Complexity = bumpComplexity(input.Complexity)
			diversityRef = ptrSnapshot(makeSnapshotFromIdea(idea, input))
			diversityAttempts = 0
			pendingReason = ptrRetry(ai.RetryUserRejected)
			pendingStrategy = selectPivotStrategy(ai.RetryUserRejected)
			continue
		case "back":
			return nil
		default:
			return fmt.Errorf("generate: invalid selection")
		}
	}
}

var errDuplicateDNA = errors.New("duplicate dna")

func saveGeneratedProject(ctx context.Context, input model.ProjectInput, idea ai.ProjectIdea, rawJSON string, meta ai.AIResult, retryReason *ai.RetryReason) error {
	gdb, err := db.Connect(ctx)
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return fmt.Errorf("generate: get sql db: %w", err)
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	mvp := idea.Project.MVP.MustHave
	stack := flattenTechStack(idea.Project.TechStack)
	overview := buildProjectOverview(idea)

	mvpJSON, err := json.Marshal(mvp)
	if err != nil {
		return fmt.Errorf("generate: marshal mvp scope: %w", err)
	}
	techJSON, err := json.Marshal(stack)
	if err != nil {
		return fmt.Errorf("generate: marshal tech stack: %w", err)
	}

	providerUsed := strings.TrimSpace(meta.ProviderUsed)
	if providerUsed == "" {
		providerUsed = "gemini"
	}
	var providerErrPtr *string
	if strings.TrimSpace(meta.ProviderError) != "" {
		v := meta.ProviderError
		providerErrPtr = &v
	}
	var retryPtr *string
	if retryReason != nil && strings.TrimSpace(string(*retryReason)) != "" {
		v := string(*retryReason)
		retryPtr = &v
	}

	row := pmodels.Project{
		ID:              uuid.New(),
		Title:           strings.TrimSpace(idea.Project.Name),
		Summary:         strings.TrimSpace(idea.Project.Description.Summary),
		ProjectOverview: overview,
		ProjectKind:     strings.TrimSpace(input.ProjectKind),
		MVPScopeJSON:    string(mvpJSON),
		TechStackJSON:   string(techJSON),
		RawAIOutput:     rawJSON,
		AppType:         input.AppType,
		Goal:            input.Goal,

		Complexity: idea.Project.Complexity,
		Duration:   idea.Project.Duration.Range,

		DNAHash:         project.HashContent(overview, mvp, stack, idea.Project.Complexity, idea.Project.Duration.Range),
		SimilarityScore: 0,
		PivotReason:     retryPtr,

		AIProvider:    providerUsed,
		ProviderUsed:  providerUsed,
		FallbackUsed:  meta.FallbackUsed,
		ProviderError: providerErrPtr,
		LatencyMS:     meta.LatencyMS,
		RetryReason:   retryPtr,

		CreatedAt: time.Now(),
	}

	tx := gdb.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("generate: save project: begin transaction: %w", tx.Error)
	}
	defer func() { _ = tx.Rollback() }()

	if err := tx.Create(&row).Error; err != nil {
		if isUniqueViolation(err) {
			return errDuplicateDNA
		}
		return fmt.Errorf("generate: save project: %w", err)
	}

	// Persist structured feature lists to project_features so nothing is lost in downstream usage.
	var features []pmodels.ProjectFeature
	appendFeatures := func(typ string, items []string) {
		for _, v := range items {
			v = strings.TrimSpace(v)
			if v == "" {
				continue
			}
			features = append(features, pmodels.ProjectFeature{
				ID:          uuid.New(),
				ProjectID:   row.ID,
				Type:        typ,
				Description: v,
			})
		}
	}
	appendFeatures("mvp_must_have", idea.Project.MVP.MustHave)
	appendFeatures("mvp_nice_to_have", idea.Project.MVP.NiceToHave)
	appendFeatures("out_of_scope", idea.Project.MVP.OutOfScope)
	appendFeatures("future_extension", idea.Project.Future)
	appendFeatures("learning_outcome", idea.Project.Learning)
	appendFeatures("key_benefit", idea.Project.ValueProp.KeyBenefits)

	if len(features) > 0 {
		if err := tx.Create(&features).Error; err != nil {
			return fmt.Errorf("generate: save project features: %w", err)
		}
	}

	// Also store a compact meta record (in addition to raw_ai_output on projects) for easy querying.
	targetUsersJSON, err := json.Marshal(map[string]any{
		"primary":   idea.Project.TargetUsers.Primary,
		"secondary": idea.Project.TargetUsers.Secondary,
		"use_cases": idea.Project.TargetUsers.UseCases,
	})
	if err != nil {
		return fmt.Errorf("generate: marshal target users: %w", err)
	}
	metaRow := pmodels.ProjectMeta{
		ProjectID:   row.ID,
		TargetUsers: string(targetUsersJSON),
		TechStack:   strings.Join(stack, ", "),
		RawAIOutput: rawJSON,
	}
	if err := tx.Create(&metaRow).Error; err != nil {
		return fmt.Errorf("generate: save project meta: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("generate: save project: commit: %w", err)
	}

	return nil
}

func evaluateSimilarity(ctx context.Context, idea ai.ProjectIdea, input model.ProjectInput) (project.SimilarityDecision, float64, error) {
	gdb, err := db.Connect(ctx)
	if err != nil {
		return project.SimilarityOK, 0, fmt.Errorf("generate: %w", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return project.SimilarityOK, 0, fmt.Errorf("generate: get sql db: %w", err)
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	limit := loadSimilarityLimit()
	var rows []pmodels.Project
	if err := gdb.Order("created_at desc").Limit(limit).Find(&rows).Error; err != nil {
		return project.SimilarityOK, 0, fmt.Errorf("generate: load recent projects: %w", err)
	}
	if len(rows) == 0 {
		return project.SimilarityOK, 0, nil
	}

	current := project.Snapshot{
		Overview:          buildProjectOverview(idea),
		MVPScope:          idea.Project.MVP.MustHave,
		TechStack:         flattenTechStack(idea.Project.TechStack),
		Complexity:        idea.Project.Complexity,
		EstimatedDuration: idea.Project.Duration.Range,
		AppType:           input.AppType,
		Goal:              input.Goal,
	}

	best := 0.0
	for _, row := range rows {
		mvp, err := parseStringArray(row.MVPScopeJSON)
		if err != nil {
			return project.SimilarityOK, 0, fmt.Errorf("generate: parse mvp scope: %w", err)
		}
		stack, err := parseStringArray(row.TechStackJSON)
		if err != nil {
			return project.SimilarityOK, 0, fmt.Errorf("generate: parse tech stack: %w", err)
		}

		prev := project.Snapshot{
			Overview:          row.ProjectOverview,
			MVPScope:          mvp,
			TechStack:         stack,
			Complexity:        row.Complexity,
			EstimatedDuration: row.Duration,
			AppType:           row.AppType,
			Goal:              row.Goal,
		}
		score := project.JaccardSimilarity(current, prev)
		if score > best {
			best = score
		}
	}

	return project.DecideSimilarity(best), best, nil
}

func parseStringArray(raw string) ([]string, error) {
	if raw == "" {
		return []string{}, nil
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, err
	}
	return out, nil
}

func loadSimilarityLimit() int {
	const defaultLimit = 50
	if v := os.Getenv("SIMILARITY_LOOKBACK_N"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return defaultLimit
}

func buildProjectOverview(idea ai.ProjectIdea) string {
	name := strings.TrimSpace(idea.Project.Name)
	tagline := strings.TrimSpace(idea.Project.Tagline)
	summary := strings.TrimSpace(idea.Project.Description.Summary)
	parts := []string{name}
	if tagline != "" {
		parts = append(parts, tagline)
	}
	if summary != "" {
		parts = append(parts, summary)
	}
	return strings.Join(parts, " — ")
}

func flattenTechStack(stack ai.ProjectTechStack) []string {
	out := []string{}
	fields := []string{stack.Backend, stack.Frontend, stack.Database, stack.Infra}
	for _, v := range fields {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		out = append(out, v)
	}
	return out
}

func selectPivotStrategy(reason ai.RetryReason) ai.PivotStrategy {
	switch reason {
	case ai.RetrySimilarityTooHigh:
		return ai.PivotChangeTargetUser
	case ai.RetryDuplicateDNA:
		return ai.PivotContextShift
	case ai.RetryUserRejected:
		return ai.PivotFeatureReplacement
	default:
		return ai.PivotFeatureReplacement
	}
}

func ptrRetry(r ai.RetryReason) *ai.RetryReason {
	return &r
}

func bumpComplexity(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "beginner":
		return "intermediate"
	case "intermediate":
		return "advanced"
	case "advanced":
		return "advanced"
	default:
		// Preserve current behavior if user entered a custom complexity string.
		return strings.TrimSpace(v)
	}
}

func lowerComplexity(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "advanced":
		return "intermediate"
	case "intermediate":
		return "beginner"
	case "beginner":
		return "beginner"
	default:
		// Preserve current behavior if user entered a custom complexity string.
		return strings.TrimSpace(v)
	}
}

func rotatePivotStrategy(attempt int) ai.PivotStrategy {
	switch attempt % 3 {
	case 1:
		return ai.PivotChangeTargetUser
	case 2:
		return ai.PivotContextShift
	default:
		return ai.PivotFeatureReplacement
	}
}

func makeSnapshotFromIdea(idea ai.ProjectIdea, input model.ProjectInput) project.Snapshot {
	return project.Snapshot{
		Overview:          buildProjectOverview(idea),
		MVPScope:          idea.Project.MVP.MustHave,
		TechStack:         flattenTechStack(idea.Project.TechStack),
		Complexity:        idea.Project.Complexity,
		EstimatedDuration: idea.Project.Duration.Range,
		AppType:           input.AppType,
		Goal:              input.Goal,
	}
}

func ptrSnapshot(s project.Snapshot) *project.Snapshot { return &s }

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "sqlstate 23505") {
		return true
	}
	if strings.Contains(s, "duplicate key") {
		return true
	}
	if strings.Contains(s, "unique constraint") {
		return true
	}
	return false
}

func runContinueExisting(ctx context.Context, _ *os.File, out io.Writer) error {
	projects, err := loadRecentProjects(ctx)
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		fmt.Fprintln(out, "No existing projects found.")
		return nil
	}

	options := make([]tui.Option, 0, len(projects))
	for _, p := range projects {
		options = append(options, tui.Option{
			ID:    p.ID.String(),
			Label: fmt.Sprintf("%s (%s, %s)", p.ProjectOverview, p.Complexity, p.Duration),
		})
	}

	selection, err := tui.SelectOption(os.Stdin, out, "Select a project:", options)
	if err != nil {
		return err
	}

	var selected *pmodels.Project
	for i := range projects {
		if projects[i].ID.String() == selection.ID {
			selected = &projects[i]
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("continue: invalid selection")
	}

	mvp, err := parseStringArray(selected.MVPScopeJSON)
	if err != nil {
		return fmt.Errorf("continue: parse mvp scope: %w", err)
	}
	stack, err := parseStringArray(selected.TechStackJSON)
	if err != nil {
		return fmt.Errorf("continue: parse tech stack: %w", err)
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Project Context")
	fmt.Fprintln(out, selected.ProjectOverview)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "MVP Scope")
	for _, item := range mvp {
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Tech Stack")
	for _, item := range stack {
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Complexity")
	fmt.Fprintln(out, selected.Complexity)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Estimated Duration")
	fmt.Fprintln(out, selected.Duration)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "App Type")
	fmt.Fprintln(out, selected.AppType)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Goal")
	fmt.Fprintln(out, selected.Goal)
	return runProjectEvolution(ctx, out, selected, mvp, stack)
}

func loadRecentProjects(ctx context.Context) ([]pmodels.Project, error) {
	gdb, err := db.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("continue: %w", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("continue: get sql db: %w", err)
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	limit := loadSimilarityLimit()
	var rows []pmodels.Project
	if err := gdb.Order("created_at desc").Limit(limit).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("continue: load projects: %w", err)
	}
	return rows, nil
}

func runProjectEvolution(ctx context.Context, out io.Writer, selected *pmodels.Project, mvp []string, stack []string) error {
	input := ai.EvolutionInput{
		ProjectOverview:   selected.ProjectOverview,
		MVPScope:          mvp,
		TechStack:         stack,
		Complexity:        selected.Complexity,
		EstimatedDuration: selected.Duration,
		AppType:           selected.AppType,
		Goal:              selected.Goal,
	}

	for {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Generating next evolution with AI")

		evo, rawJSON, meta, err := ai.GenerateProjectEvolutionWithMeta(ctx, input)
		if err != nil {
			return fmt.Errorf("continue: %w", err)
		}

		printEvolution(out, evo)

		selection, err := tui.SelectOption(os.Stdin, out, "Next action:", []tui.Option{
			{ID: "accept", Label: "Accept"},
			{ID: "regenerate", Label: "Regenerate"},
			{ID: "back", Label: "Back"},
		})
		if err != nil {
			return err
		}

		switch selection.ID {
		case "accept":
			if err := saveProjectEvolution(ctx, selected.ID, rawJSON, meta); err != nil {
				return err
			}
			return nil
		case "regenerate":
			continue
		case "back":
			return nil
		default:
			return fmt.Errorf("continue: invalid selection")
		}
	}
}

func saveProjectEvolution(ctx context.Context, projectID uuid.UUID, rawJSON string, meta ai.AIResult) error {
	gdb, err := db.Connect(ctx)
	if err != nil {
		return fmt.Errorf("continue: %w", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return fmt.Errorf("continue: get sql db: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	providerUsed := strings.TrimSpace(meta.ProviderUsed)
	if providerUsed == "" {
		providerUsed = "gemini"
	}
	var providerErrPtr *string
	if strings.TrimSpace(meta.ProviderError) != "" {
		v := meta.ProviderError
		providerErrPtr = &v
	}

	row := pmodels.ProjectEvolution{
		ID:            uuid.New(),
		ProjectID:     projectID,
		RawAIOutput:   rawJSON,
		ProviderUsed:  providerUsed,
		FallbackUsed:  meta.FallbackUsed,
		ProviderError: providerErrPtr,
		LatencyMS:     meta.LatencyMS,
		CreatedAt:     time.Now(),
	}
	if err := gdb.Create(&row).Error; err != nil {
		return fmt.Errorf("continue: save evolution: %w", err)
	}
	return nil
}

func runViewSavedProjects(ctx context.Context, out io.Writer) error {
	projects, err := loadRecentProjects(ctx)
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		fmt.Fprintln(out, "No existing projects found.")
		return nil
	}

	options := make([]tui.Option, 0, len(projects))
	for _, p := range projects {
		options = append(options, tui.Option{
			ID:    p.ID.String(),
			Label: fmt.Sprintf("%s (%s, %s)", p.ProjectOverview, p.Complexity, p.Duration),
		})
	}

	selection, err := tui.SelectOption(os.Stdin, out, "Select a project:", options)
	if err != nil {
		return err
	}

	var selected *pmodels.Project
	for i := range projects {
		if projects[i].ID.String() == selection.ID {
			selected = &projects[i]
			break
		}
	}
	if selected == nil {
		return fmt.Errorf("view: invalid selection")
	}

	var idea ai.ProjectIdea
	if err := json.Unmarshal([]byte(selected.RawAIOutput), &idea); err != nil {
		return fmt.Errorf("view: parse saved raw_ai_output: %w", err)
	}

	printIdea(out, idea)

	evolutions, err := loadProjectEvolutions(ctx, selected.ID)
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Saved Evolutions")
	for i := range evolutions {
		var evo ai.ProjectEvolution
		if err := json.Unmarshal([]byte(evolutions[i].RawAIOutput), &evo); err != nil {
			return fmt.Errorf("view: parse saved evolution: %w", err)
		}
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Evolution #%d\n", i+1)
		printEvolution(out, evo)
	}

	for {
		after, _ := tui.SelectOption(os.Stdin, out, "Next action:", []tui.Option{
			{ID: "copy", Label: "Copy output to clipboard"},
			{ID: "back", Label: "Back"},
		})
		switch after.ID {
		case "copy":
			var buf bytes.Buffer
			printIdea(&buf, idea)
			if len(evolutions) > 0 {
				fmt.Fprintln(&buf, "")
				fmt.Fprintln(&buf, "Saved Evolutions")
				for i := range evolutions {
					var evo ai.ProjectEvolution
					if err := json.Unmarshal([]byte(evolutions[i].RawAIOutput), &evo); err != nil {
						return fmt.Errorf("view: parse saved evolution: %w", err)
					}
					fmt.Fprintln(&buf, "")
					fmt.Fprintf(&buf, "Evolution #%d\n", i+1)
					printEvolution(&buf, evo)
					fmt.Fprintln(&buf, "")
					fmt.Fprintln(&buf, "----")
					fmt.Fprintln(&buf, "Evolution Raw JSON")
					fmt.Fprintln(&buf, evolutions[i].RawAIOutput)
				}
			}
			fmt.Fprintln(&buf, "")
			fmt.Fprintln(&buf, "----")
			fmt.Fprintln(&buf, "Project Raw JSON")
			fmt.Fprintln(&buf, selected.RawAIOutput)
			if err := tui.CopyToClipboard(out, buf.String()); err != nil {
				fmt.Fprintf(out, "Copy failed: %v\n", err)
				continue
			}
			fmt.Fprintln(out, "Copied to clipboard.")
		default:
			return nil
		}
	}
	return nil
}

func loadProjectEvolutions(ctx context.Context, projectID uuid.UUID) ([]pmodels.ProjectEvolution, error) {
	gdb, err := db.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("view: %w", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("view: get sql db: %w", err)
	}
	defer func() { _ = sqlDB.Close() }()

	var rows []pmodels.ProjectEvolution
	if err := gdb.Where("project_id = ?", projectID).Order("created_at asc").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("view: load evolutions: %w", err)
	}
	return rows, nil
}

func printIdea(out io.Writer, idea ai.ProjectIdea) {
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Project")
	fmt.Fprintf(out, "%s — %s\n", idea.Project.Name, idea.Project.Tagline)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Summary")
	fmt.Fprintln(out, idea.Project.Description.Summary)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Detailed Explanation")
	fmt.Fprintln(out, idea.Project.Description.DetailedExplanation)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Problem Statement")
	fmt.Fprintln(out, idea.Project.Problem.Problem)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Why It Matters")
	fmt.Fprintln(out, idea.Project.Problem.WhyItMatters)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Current Solutions and Gaps")
	fmt.Fprintln(out, idea.Project.Problem.CurrentSolutionsAndGaps)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Target Users (Primary)")
	for _, item := range idea.Project.TargetUsers.Primary {
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Target Users (Secondary)")
	for _, item := range idea.Project.TargetUsers.Secondary {
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Use Cases")
	for _, item := range idea.Project.TargetUsers.UseCases {
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Value Proposition")
	for _, item := range idea.Project.ValueProp.KeyBenefits {
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Why This Project Is Interesting")
	fmt.Fprintln(out, idea.Project.ValueProp.WhyThisProjectIsInteresting)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Portfolio Value")
	fmt.Fprintln(out, idea.Project.ValueProp.PortfolioValue)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "MVP Goal")
	fmt.Fprintln(out, idea.Project.MVP.Goal)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "MVP Must-Have Features")
	for _, item := range idea.Project.MVP.MustHave {
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "MVP Nice-to-Have Features")
	for _, item := range idea.Project.MVP.NiceToHave {
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Out of Scope")
	for _, item := range idea.Project.MVP.OutOfScope {
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Recommended Tech Stack")
	fmt.Fprintf(out, "- Backend: %s\n", idea.Project.TechStack.Backend)
	fmt.Fprintf(out, "- Frontend: %s\n", idea.Project.TechStack.Frontend)
	fmt.Fprintf(out, "- Database: %s\n", idea.Project.TechStack.Database)
	fmt.Fprintf(out, "- Infra: %s\n", idea.Project.TechStack.Infra)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Tech Stack Justification")
	fmt.Fprintln(out, idea.Project.TechStack.Justification)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Complexity")
	fmt.Fprintln(out, idea.Project.Complexity)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Estimated Duration")
	fmt.Fprintln(out, idea.Project.Duration.Range)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Duration Assumptions")
	fmt.Fprintln(out, idea.Project.Duration.Assumptions)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Future Extensions")
	for _, item := range idea.Project.Future {
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Learning Outcomes")
	for _, item := range idea.Project.Learning {
		fmt.Fprintf(out, "- %s\n", item)
	}
}

func printEvolution(out io.Writer, evo ai.ProjectEvolution) {
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Next Project Evolution")
	fmt.Fprintln(out, evo.EvolutionOverview)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Product Rationale")
	fmt.Fprintln(out, evo.ProductRationale)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Technical Rationale")
	fmt.Fprintln(out, evo.TechnicalRationale)
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Proposed Enhancements")
	for _, item := range evo.ProposedEnhancements {
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Risk Considerations")
	for _, item := range evo.RiskConsiderations {
		fmt.Fprintf(out, "- %s\n", item)
	}
}
