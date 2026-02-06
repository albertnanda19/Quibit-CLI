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
			tui.AppHeader(out)
			tui.Context(out, "Select a mode.")
			options := []tui.Option{
				{ID: "new", Label: "Generate project"},
				{ID: "idea", Label: "Generate from my own idea / problem"},
				{ID: "continue", Label: "Continue project"},
				{ID: "view", Label: "View saved projects"},
				{ID: "exit", Label: "Quit"},
			}

			selection, err := tui.SelectOption(os.Stdin, out, "", options)
			if err != nil {
				return fmt.Errorf("generate: %w", err)
			}

			switch selection.ID {
			case "new":
				tui.Transition(ctx, out)
				if err := runGenerateNew(ctx, os.Stdin, out); err != nil {
					return err
				}
			case "idea":
				tui.Transition(ctx, out)
				if err := runGenerateFromUserIdea(ctx, os.Stdin, out); err != nil {
					return err
				}
			case "continue":
				tui.Transition(ctx, out)
				if err := runContinueExisting(ctx, os.Stdin, out); err != nil {
					return err
				}
			case "view":
				tui.Transition(ctx, out)
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
	return runGenerateWithInput(ctx, in, out, input)
}

func runGenerateFromUserIdea(ctx context.Context, in *os.File, out io.Writer) error {
	input, err := tuiinput.CollectUserIdeaProjectInput(in, out)
	if err != nil {
		return err
	}
	return runGenerateWithInput(ctx, in, out, input)
}

func runGenerateWithInput(ctx context.Context, in *os.File, out io.Writer, input model.ProjectInput) error {
	var pendingReason *ai.RetryReason
	var pendingStrategy ai.PivotStrategy
	var lastReasonUsed *ai.RetryReason
	var lastMeta ai.AIResult
	var err error
generateLoop:
	for {
		if pendingReason == nil {
			preDecision, preScore, preErr := evaluateSimilarityPre(ctx, input)
			if preErr != nil {
				return preErr
			}
			if preDecision != project.SimilarityOK {
				tui.Status(out, fmt.Sprintf("Precheck similarity %.2f detected; applying deterministic pivot", preScore))
				pendingReason = ptrRetry(ai.RetrySimilarityTooHigh)
				pendingStrategy = selectPivotStrategy(ai.RetrySimilarityTooHigh)
			}
		}

		var idea ai.ProjectIdea
		var rawJSON string
		spin := tui.StartSpinner(ctx, out, "Generating project blueprint")
		if pendingReason == nil {
			lastReasonUsed = nil
			idea, rawJSON, lastMeta, err = ai.GenerateProjectIdeaOnceWithMeta(ctx, input)
		} else {
			lastReasonUsed = pendingReason
			idea, rawJSON, lastMeta, err = ai.GenerateProjectIdeaWithPivotOnceMeta(ctx, input, *pendingReason, pendingStrategy)
			pendingReason = nil
		}
		spin.Stop()
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		simSpin := tui.StartSpinner(ctx, out, "Syncing with saved projects")
		action, bestScore, err := evaluateSimilarity(ctx, idea, input)
		simSpin.Stop()
		if err != nil {
			return err
		}
		switch action {
		case project.SimilarityRegenerate:
			tui.Status(out, fmt.Sprintf("Similarity %.2f is high; you may choose to regenerate", bestScore))
		case project.SimilarityBlock:
			tui.PrintError(out, "Generation blocked", fmt.Errorf("similarity %.2f is too high", bestScore))
			return nil
		default:
		}

		printIdea(out, idea, input)

		selection, err := tui.SelectOption(in, out, "Choose next action.", []tui.Option{
			{ID: "accept", Label: "Accept and save"},
			{ID: "regenerate", Label: "Regenerate"},
			{ID: "regenerate_harder", Label: "Regenerate (increase complexity)"},
			{ID: "back", Label: "Back"},
		})
		if err != nil {
			return err
		}

		switch selection.ID {
		case "accept":
			saveSpin := tui.StartSpinner(ctx, out, "Saving project")
			err := saveGeneratedProject(ctx, input, idea, rawJSON, lastMeta, lastReasonUsed)
			saveSpin.Stop()
			if err != nil {
				if errors.Is(err, errDuplicateDNA) {
					tui.Status(out, "Duplicate result detected; regenerating")
					pendingReason = ptrRetry(ai.RetryDuplicateDNA)
					pendingStrategy = selectPivotStrategy(ai.RetryDuplicateDNA)
					continue
				}
				return err
			}
			tui.BlankLine(out)
			tui.Done(out, "Saved")
			for {
				after, _ := tui.SelectOption(in, out, "Choose next action.", []tui.Option{
					{ID: "copy", Label: "Copy output"},
					{ID: "same", Label: "Generate another (same inputs)"},
					{ID: "same_harder", Label: "Generate another (increase complexity)"},
					{ID: "same_easier", Label: "Generate another (decrease complexity)"},
					{ID: "back", Label: "Back"},
				})
				switch after.ID {
				case "copy":
					var buf bytes.Buffer
					printIdea(&buf, idea, input)
					fmt.Fprintln(&buf, "")
					fmt.Fprintln(&buf, "----")
					fmt.Fprintln(&buf, "Raw JSON")
					fmt.Fprintln(&buf, rawJSON)
					if err := tui.CopyToClipboard(out, buf.String()); err != nil {
						tui.PrintError(out, "Unable to copy to clipboard", err)
						continue
					}
					tui.Done(out, "Copied")
				case "same":
					pendingReason = nil
					lastReasonUsed = nil
					continue generateLoop
				case "same_harder":
					input.Complexity = bumpComplexity(input.Complexity)
					pendingReason = nil
					lastReasonUsed = nil
					continue generateLoop
				case "same_easier":
					input.Complexity = lowerComplexity(input.Complexity)
					pendingReason = nil
					lastReasonUsed = nil
					continue generateLoop
				default:
					return nil
				}
			}
		case "regenerate":
			pendingReason = ptrRetry(ai.RetryUserRejected)
			pendingStrategy = selectPivotStrategy(ai.RetryUserRejected)
			continue
		case "regenerate_harder":
			input.Complexity = bumpComplexity(input.Complexity)
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

func evaluateSimilarityPre(ctx context.Context, input model.ProjectInput) (project.SimilarityDecision, float64, error) {
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
		Overview:          "",
		MVPScope:          []string{},
		TechStack:         input.TechStack,
		Complexity:        input.Complexity,
		EstimatedDuration: input.Timeframe,
		AppType:           input.AppType,
		Goal:              input.Goal,
	}

	best := 0.0
	for _, row := range rows {
		stack, err := parseStringArray(row.TechStackJSON)
		if err != nil {
			return project.SimilarityOK, 0, fmt.Errorf("generate: parse tech stack: %w", err)
		}

		prev := project.Snapshot{
			Overview:          "",
			MVPScope:          []string{},
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
	loadSpin := tui.StartSpinner(ctx, out, "Loading saved projects")
	projects, err := loadRecentProjects(ctx)
	loadSpin.Stop()
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
		spin := tui.StartSpinner(ctx, out, "Generating next evolution")
		evo, rawJSON, meta, err := ai.GenerateProjectEvolutionWithMeta(ctx, input)
		spin.Stop()
		if err != nil {
			return fmt.Errorf("continue: %w", err)
		}

		printEvolution(out, evo)

		selection, err := tui.SelectOption(os.Stdin, out, "Choose next action.", []tui.Option{
			{ID: "accept", Label: "Accept and save"},
			{ID: "regenerate", Label: "Regenerate"},
			{ID: "back", Label: "Back"},
		})
		if err != nil {
			return err
		}

		switch selection.ID {
		case "accept":
			saveSpin := tui.StartSpinner(ctx, out, "Saving evolution")
			err := saveProjectEvolution(ctx, selected.ID, rawJSON, meta)
			saveSpin.Stop()
			if err != nil {
				return err
			}
			tui.Done(out, "Saved")
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
	loadSpin := tui.StartSpinner(ctx, out, "Loading saved projects")
	projects, err := loadRecentProjects(ctx)
	loadSpin.Stop()
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		tui.BlankLine(out)
		tui.Context(out, "No saved projects.")
		tui.Hint(out, "Generate a project to create your first saved entry.")
		return nil
	}

	entries := buildSavedProjectEntries(projects)
	selection, err := tui.SelectEntries(os.Stdin, out, "Select a project.", entries)
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

	printIdea(out, idea, model.ProjectInput{})

	evoSpin := tui.StartSpinner(ctx, out, "Loading evolutions")
	evolutions, err := loadProjectEvolutions(ctx, selected.ID)
	evoSpin.Stop()
	if err != nil {
		return err
	}

	tui.Heading(out, "Saved Evolutions")
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
		after, _ := tui.SelectOption(os.Stdin, out, "Choose next action.", []tui.Option{
			{ID: "copy", Label: "Copy output"},
			{ID: "back", Label: "Back"},
		})
		switch after.ID {
		case "copy":
			var buf bytes.Buffer
			printIdea(&buf, idea, model.ProjectInput{})
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
				tui.PrintError(out, "Unable to copy to clipboard", err)
				continue
			}
			tui.Done(out, "Copied")
		default:
			return nil
		}
	}
}

func buildSavedProjectEntries(projects []pmodels.Project) []tui.SelectEntry {
	type group struct {
		Title string
		Items []pmodels.Project
	}
	order := []string{
		"Web Application",
		"Mobile Application",
		"CLI Tool",
		"Backend / API Service",
		"Library / SDK",
		"Others",
	}
	groups := map[string]*group{}
	for i := range order {
		groups[order[i]] = &group{Title: order[i]}
	}

	for i := range projects {
		title := savedProjectGroupTitle(projects[i].AppType)
		groups[title].Items = append(groups[title].Items, projects[i])
	}

	entries := make([]tui.SelectEntry, 0, len(projects)+len(order))
	for _, title := range order {
		g := groups[title]
		if g == nil || len(g.Items) == 0 {
			continue
		}
		entries = append(entries, tui.SelectEntry{
			ID:         "header:" + strings.ToLower(strings.ReplaceAll(title, " ", "-")),
			Label:      title,
			Selectable: false,
		})
		for _, p := range g.Items {
			label := fmt.Sprintf("%s (%s, %s)", p.ProjectOverview, p.Complexity, p.Duration)
			entries = append(entries, tui.SelectEntry{
				ID:         p.ID.String(),
				Label:      "  ▸ " + label,
				Selectable: true,
			})
		}
	}
	return entries
}

func savedProjectGroupTitle(appType string) string {
	s := strings.ToLower(strings.TrimSpace(appType))
	if s == "" {
		return "Others"
	}
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")

	switch s {
	case "web":
		return "Web Application"
	case "mobile":
		return "Mobile Application"
	case "cli":
		return "CLI Tool"
	case "backend", "backend-api", "api", "service", "api-service":
		return "Backend / API Service"
	}

	if strings.Contains(s, "web") || strings.Contains(s, "frontend") || strings.Contains(s, "fullstack") {
		return "Web Application"
	}
	if strings.Contains(s, "mobile") || strings.Contains(s, "android") || strings.Contains(s, "ios") {
		return "Mobile Application"
	}
	if strings.Contains(s, "cli") || strings.Contains(s, "terminal") || strings.Contains(s, "command-line") {
		return "CLI Tool"
	}
	if strings.Contains(s, "backend") || strings.Contains(s, "api") || strings.Contains(s, "microservice") || strings.Contains(s, "service") {
		return "Backend / API Service"
	}
	if strings.Contains(s, "library") || strings.Contains(s, "sdk") || strings.Contains(s, "package") || strings.Contains(s, "framework") {
		return "Library / SDK"
	}
	return "Others"
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

func printIdea(out io.Writer, idea ai.ProjectIdea, input model.ProjectInput) {
	if strings.TrimSpace(input.UserIdea) != "" {
		printUserIdeaBlueprint(out, idea)
		return
	}
	tui.Heading(out, "Project")
	fmt.Fprintf(out, "%s — %s\n", idea.Project.Name, idea.Project.Tagline)

	tui.Heading(out, "Summary")
	fmt.Fprintln(out, idea.Project.Description.Summary)

	tui.Heading(out, "Detailed Explanation")
	fmt.Fprintln(out, idea.Project.Description.DetailedExplanation)

	tui.Heading(out, "Problem Statement")
	fmt.Fprintln(out, idea.Project.Problem.Problem)

	tui.Heading(out, "Why It Matters")
	fmt.Fprintln(out, idea.Project.Problem.WhyItMatters)

	tui.Heading(out, "Current Solutions and Gaps")
	fmt.Fprintln(out, idea.Project.Problem.CurrentSolutionsAndGaps)

	tui.Heading(out, "Target Users (Primary)")
	for _, item := range idea.Project.TargetUsers.Primary {
		fmt.Fprintf(out, "- %s\n", item)
	}

	tui.Heading(out, "Target Users (Secondary)")
	for _, item := range idea.Project.TargetUsers.Secondary {
		fmt.Fprintf(out, "- %s\n", item)
	}

	tui.Heading(out, "Use Cases")
	for _, item := range idea.Project.TargetUsers.UseCases {
		fmt.Fprintf(out, "- %s\n", item)
	}

	tui.Heading(out, "Value Proposition")
	for _, item := range idea.Project.ValueProp.KeyBenefits {
		fmt.Fprintf(out, "- %s\n", item)
	}

	tui.Heading(out, "Why This Project Is Interesting")
	fmt.Fprintln(out, idea.Project.ValueProp.WhyThisProjectIsInteresting)

	tui.Heading(out, "Portfolio Value")
	fmt.Fprintln(out, idea.Project.ValueProp.PortfolioValue)

	tui.Heading(out, "MVP Goal")
	fmt.Fprintln(out, idea.Project.MVP.Goal)

	tui.Heading(out, "MVP Must-Have Features")
	for _, item := range idea.Project.MVP.MustHave {
		fmt.Fprintf(out, "- %s\n", item)
	}

	tui.Heading(out, "MVP Nice-to-Have Features")
	for _, item := range idea.Project.MVP.NiceToHave {
		fmt.Fprintf(out, "- %s\n", item)
	}

	tui.Heading(out, "Out of Scope")
	for _, item := range idea.Project.MVP.OutOfScope {
		fmt.Fprintf(out, "- %s\n", item)
	}

	tui.Heading(out, "Recommended Tech Stack")
	fmt.Fprintf(out, "- Backend: %s\n", idea.Project.TechStack.Backend)
	fmt.Fprintf(out, "- Frontend: %s\n", idea.Project.TechStack.Frontend)
	fmt.Fprintf(out, "- Database: %s\n", idea.Project.TechStack.Database)
	fmt.Fprintf(out, "- Infra: %s\n", idea.Project.TechStack.Infra)

	tui.Heading(out, "Tech Stack Justification")
	fmt.Fprintln(out, idea.Project.TechStack.Justification)

	tui.Heading(out, "Complexity")
	fmt.Fprintln(out, idea.Project.Complexity)

	tui.Heading(out, "Estimated Duration")
	fmt.Fprintln(out, idea.Project.Duration.Range)

	tui.Heading(out, "Duration Assumptions")
	fmt.Fprintln(out, idea.Project.Duration.Assumptions)

	tui.Heading(out, "Future Extensions")
	for _, item := range idea.Project.Future {
		fmt.Fprintf(out, "- %s\n", item)
	}

	tui.Heading(out, "Learning Outcomes")
	for _, item := range idea.Project.Learning {
		fmt.Fprintf(out, "- %s\n", item)
	}
}

func printUserIdeaBlueprint(out io.Writer, idea ai.ProjectIdea) {
	tui.Heading(out, "Problem Statement")
	fmt.Fprintln(out, idea.Project.Problem.Problem)

	tui.Heading(out, "Project Overview")
	fmt.Fprintf(out, "%s — %s\n", idea.Project.Name, idea.Project.Tagline)
	if strings.TrimSpace(idea.Project.Description.Summary) != "" {
		fmt.Fprintln(out, idea.Project.Description.Summary)
	}

	tui.Heading(out, "Assumptions & Constraints")
	if strings.TrimSpace(idea.Project.Duration.Assumptions) != "" {
		fmt.Fprintln(out, idea.Project.Duration.Assumptions)
	} else {
		fmt.Fprintln(out, "-")
	}

	tui.Heading(out, "Tech Stack Breakdown")
	fmt.Fprintf(out, "- Backend: %s\n", idea.Project.TechStack.Backend)
	fmt.Fprintf(out, "- Frontend: %s\n", idea.Project.TechStack.Frontend)
	fmt.Fprintf(out, "- Database: %s\n", idea.Project.TechStack.Database)
	fmt.Fprintf(out, "- Infra: %s\n", idea.Project.TechStack.Infra)
	if strings.TrimSpace(idea.Project.TechStack.Justification) != "" {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, idea.Project.TechStack.Justification)
	}

	tui.Heading(out, "Core Features")
	for _, item := range idea.Project.MVP.MustHave {
		fmt.Fprintf(out, "- %s\n", item)
	}

	tui.Heading(out, "MVP Scope")
	fmt.Fprintln(out, "Included")
	for _, item := range idea.Project.MVP.MustHave {
		fmt.Fprintf(out, "- %s\n", item)
	}
	if len(idea.Project.MVP.OutOfScope) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Explicitly Excluded")
		for _, item := range idea.Project.MVP.OutOfScope {
			fmt.Fprintf(out, "- %s\n", item)
		}
	}

	tui.Heading(out, "Engineering Focus Areas")
	for _, item := range idea.Project.Learning {
		fmt.Fprintf(out, "- %s\n", item)
	}

	tui.Heading(out, "Future Extensions")
	for _, item := range idea.Project.Future {
		fmt.Fprintf(out, "- %s\n", item)
	}
}

func printEvolution(out io.Writer, evo ai.ProjectEvolution) {
	tui.Heading(out, "Next Project Evolution")
	fmt.Fprintln(out, evo.EvolutionOverview)

	tui.Heading(out, "Product Rationale")
	fmt.Fprintln(out, evo.ProductRationale)

	tui.Heading(out, "Technical Rationale")
	fmt.Fprintln(out, evo.TechnicalRationale)

	tui.Heading(out, "Proposed Enhancements")
	for _, item := range evo.ProposedEnhancements {
		fmt.Fprintf(out, "- %s\n", item)
	}

	tui.Heading(out, "Risk Considerations")
	for _, item := range evo.RiskConsiderations {
		fmt.Fprintf(out, "- %s\n", item)
	}
}
