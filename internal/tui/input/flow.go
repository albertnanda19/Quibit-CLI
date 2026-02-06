package input

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"quibit/internal/model"
	"quibit/internal/techstack"
	"quibit/internal/tui"
)

func CollectNewProjectInput(in *os.File, out io.Writer) (model.ProjectInput, error) {
	reader := bufio.NewReader(in)

	tui.AppHeader(out)
	tui.Heading(out, "Project setup")
	tui.Context(out, "Define constraints for generation. Defaults are preselected.")
	tui.Divider(out)

	appType, err := promptSelectWithCustom(in, out, reader, ApplicationTypePrompt)
	if err != nil {
		return model.ProjectInput{}, err
	}
	tui.Divider(out)

	techStack, database, err := collectTechStackAndDatabase(in, out, reader, appType)
	if err != nil {
		return model.ProjectInput{}, err
	}

	projectKind, err := promptSelectOptionalWithCustom(in, out, reader, ProjectKindPrompt)
	if err != nil {
		return model.ProjectInput{}, err
	}
	tui.Divider(out)

	complexity, err := promptSelectWithCustom(in, out, reader, ComplexityPrompt)
	if err != nil {
		return model.ProjectInput{}, err
	}
	tui.Divider(out)

	goal, err := promptSelectWithCustom(in, out, reader, ProjectGoalPrompt)
	if err != nil {
		return model.ProjectInput{}, err
	}
	tui.Divider(out)

	timeframe, err := promptSelectWithCustom(in, out, reader, EstimatedTimeframePrompt)
	if err != nil {
		return model.ProjectInput{}, err
	}

	return model.ProjectInput{
		UserIdea:    "",
		AppType:     appType,
		ProjectKind: projectKind,
		Complexity:  complexity,
		TechStack:   techStack,
		Database:    database,
		Goal:        goal,
		Timeframe:   timeframe,
	}, nil
}

func CollectUserIdeaProjectInput(in *os.File, out io.Writer) (model.ProjectInput, error) {
	reader := bufio.NewReader(in)

	tui.AppHeader(out)
	tui.Heading(out, "Project setup")
	tui.Context(out, "Define constraints for generation. Defaults are preselected.")
	tui.Divider(out)

	printStepHeader(out, "Idea / Problem", "Describe your project idea or problem in your own words.", "")
	userIdea, err := promptWithDefault(reader, out, "Input", "")
	if err != nil {
		return model.ProjectInput{}, err
	}
	tui.Divider(out)

	appType, err := promptSelectWithCustom(in, out, reader, ApplicationTypePrompt)
	if err != nil {
		return model.ProjectInput{}, err
	}
	tui.Divider(out)

	printStepHeader(out, "Tech Stack Mode", "Choose whether to use AI recommendation or select your own stack.", "")
	mode, err := tui.SelectOption(in, out, "", []tui.Option{
		{ID: "ai", Label: "Use AI recommended tech stack"},
		{ID: "manual", Label: "Pick tech stack myself"},
	})
	if err != nil {
		return model.ProjectInput{}, err
	}
	tui.Divider(out)

	var techStack []string
	var database []string
	if mode.ID == "manual" {
		techStack, database, err = collectTechStackAndDatabase(in, out, reader, appType)
		if err != nil {
			return model.ProjectInput{}, err
		}
	}

	return model.ProjectInput{
		UserIdea:    strings.TrimSpace(userIdea),
		AppType:     appType,
		ProjectKind: "",
		Complexity:  "intermediate",
		TechStack:   techStack,
		Database:    database,
		Goal:        "portfolio project",
		Timeframe:   "2-4 weeks",
	}, nil
}

func collectTechStackAndDatabase(in *os.File, out io.Writer, reader *bufio.Reader, appType string) ([]string, []string, error) {
	appType = strings.TrimSpace(appType)

	var techStack []string
	var database []string
	selectLanguageThenMaybeFramework := func(part string, frameworkPrompt SelectPrompt) (string, bool, error) {
		mode, err := promptSelectWithCustom(in, out, reader, WebSplitStackSelectionModePrompt(part))
		if err != nil {
			return "", false, err
		}
		tui.Divider(out)

		switch strings.TrimSpace(strings.ToLower(mode)) {
		case "language":
			lang, langAI, err := promptSelectWithAIRecommendation(in, out, reader, ProgrammingLanguagePrompt(), "Use AI recommendation")
			if err != nil {
				return "", false, err
			}
			if langAI {
				return "", true, nil
			}
			tui.Divider(out)

			fw, fwAI, err := promptSelectWithAIRecommendation(in, out, reader, FrameworkPrompt(strings.TrimSpace(lang)), "Use AI recommendation")
			if err != nil {
				return "", false, err
			}
			if fwAI {
				return strings.TrimSpace(lang), false, nil
			}
			return strings.TrimSpace(lang) + " " + strings.TrimSpace(fw), false, nil
		default:
			fw, fwAI, err := promptSelectWithAIRecommendation(in, out, reader, frameworkPrompt, "Use AI recommendation")
			if err != nil {
				return "", false, err
			}
			return strings.TrimSpace(fw), fwAI, nil
		}
	}

	if appType == "web" {
		arch, err := promptSelectWithCustom(in, out, reader, WebArchitecturePrompt)
		if err != nil {
			return nil, nil, err
		}
		tui.Divider(out)

		switch strings.TrimSpace(strings.ToLower(arch)) {
		case "mvc":
			mvcFramework, mvcAI, err := promptSelectWithAIRecommendation(in, out, reader, WebMVCFrameworkPrompt, "Use AI recommendation")
			if err != nil {
				return nil, nil, err
			}
			if !mvcAI {
				techStack = []string{strings.TrimSpace(mvcFramework)}
			}
			tui.Divider(out)

			dbRaw, dbAI, err := promptSelectWithAIRecommendation(in, out, reader, DatabasePrompt, "Use AI recommendation")
			if err != nil {
				return nil, nil, err
			}
			if !dbAI {
				database = parseList(dbRaw)
			}
			tui.Divider(out)
			return techStack, database, nil
		case "split":
			frontend, frontendAI, err := selectLanguageThenMaybeFramework("Frontend", WebFrontendFrameworkPrompt)
			if err != nil {
				return nil, nil, err
			}
			tui.Divider(out)

			backend, backendAI, err := selectLanguageThenMaybeFramework("Backend", WebBackendFrameworkPrompt)
			if err != nil {
				return nil, nil, err
			}
			tui.Divider(out)

			dbRaw, dbAI, err := promptSelectWithAIRecommendation(in, out, reader, DatabasePrompt, "Use AI recommendation")
			if err != nil {
				return nil, nil, err
			}
			if !dbAI {
				database = parseList(dbRaw)
			}
			tui.Divider(out)

			if !frontendAI {
				techStack = append(techStack, strings.TrimSpace(frontend))
			}
			if !backendAI {
				techStack = append(techStack, strings.TrimSpace(backend))
			}
			return techStack, database, nil
		default:
		}
	}

	languageID, err := promptSelectWithCustom(in, out, reader, ProgrammingLanguagePrompt())
	if err != nil {
		return nil, nil, err
	}
	tui.Divider(out)

	var frameworkRaw string
	if techstack.LanguageExists(strings.TrimSpace(languageID)) {
		frameworkRaw, err = promptSelectWithCustom(in, out, reader, FrameworkPrompt(strings.TrimSpace(languageID)))
		if err != nil {
			return nil, nil, err
		}
	} else {
		frameworkRaw, err = promptWithDefault(reader, out, "Framework / Library (Custom / Manual)", "")
		if err != nil {
			return nil, nil, err
		}
	}
	frameworks := parseList(frameworkRaw)
	techStack = append([]string{strings.TrimSpace(languageID)}, frameworks...)
	tui.Divider(out)

	dbRaw, err := promptSelectWithCustom(in, out, reader, DatabasePrompt)
	if err != nil {
		return nil, nil, err
	}
	database = parseList(dbRaw)
	tui.Divider(out)

	return techStack, database, nil
}

func promptSelectWithCustom(in *os.File, out io.Writer, reader *bufio.Reader, p SelectPrompt) (string, error) {
	printStepHeader(out, p.Title, p.Description, p.Default.Label)
	options := buildOptions(p)
	selection, err := tui.SelectOptionWithDefault(in, out, "", options, p.Default.Value)
	if err != nil {
		return "", err
	}
	if selection.ID != "custom" {
		return selection.ID, nil
	}
	return promptWithDefault(reader, out, "Custom input", p.Default.Value)
}

func promptSelectWithAIRecommendation(in *os.File, out io.Writer, reader *bufio.Reader, p SelectPrompt, aiLabel string) (string, bool, error) {
	printStepHeader(out, p.Title, p.Description, p.Default.Label)
	options := buildOptionsWithAI(p, aiLabel)
	selection, err := tui.SelectOptionWithDefault(in, out, "", options, p.Default.Value)
	if err != nil {
		return "", false, err
	}
	if selection.ID == "ai" {
		return "", true, nil
	}
	if selection.ID != "custom" {
		return selection.ID, false, nil
	}
	v, err := promptWithDefault(reader, out, "Custom input", p.Default.Value)
	if err != nil {
		return "", false, err
	}
	return v, false, nil
}

func promptSelectOptionalWithCustom(in *os.File, out io.Writer, reader *bufio.Reader, p SelectPrompt) (string, error) {
	printStepHeader(out, p.Title, p.Description, p.Default.Label)
	options := buildOptionsOptional(p)
	selection, err := tui.SelectOptionWithDefault(in, out, "", options, "skip")
	if err != nil {
		return "", err
	}
	switch selection.ID {
	case "skip":
		return "", nil
	case "custom":
		v, err := promptWithDefault(reader, out, "Custom input (leave empty to skip)", "")
		if err != nil {
			return "", err
		}
		v = strings.TrimSpace(v)
		return v, nil
	default:
		return selection.ID, nil
	}
}

func buildOptions(p SelectPrompt) []tui.Option {
	options := make([]tui.Option, 0, len(p.Options)+1)
	options = append(options, tui.Option{
		ID:    "custom",
		Label: p.CustomLabel,
	})
	for _, opt := range p.Options {
		options = append(options, tui.Option{
			ID:    opt.Value,
			Label: opt.Label,
		})
	}
	return options
}

func buildOptionsWithAI(p SelectPrompt, aiLabel string) []tui.Option {
	options := make([]tui.Option, 0, len(p.Options)+2)
	options = append(options, tui.Option{ID: "ai", Label: aiLabel})
	options = append(options, tui.Option{ID: "custom", Label: p.CustomLabel})
	for _, opt := range p.Options {
		options = append(options, tui.Option{ID: opt.Value, Label: opt.Label})
	}
	return options
}

func buildOptionsOptional(p SelectPrompt) []tui.Option {

	options := []tui.Option{
		{ID: "skip", Label: p.Default.Label},
		{ID: "custom", Label: p.CustomLabel},
	}
	for _, opt := range p.Options {

		if strings.TrimSpace(opt.Value) == "" {
			continue
		}
		options = append(options, tui.Option{ID: opt.Value, Label: opt.Label})
	}
	return options
}

func promptWithDefault(reader *bufio.Reader, out io.Writer, label string, defaultValue string) (string, error) {
	tui.BlankLine(out)
	tui.Context(out, label)
	if strings.TrimSpace(defaultValue) != "" {
		tui.DefaultValue(out, defaultValue)
	}
	tui.Divider(out)
	fmt.Fprint(out, tui.PromptPrefix(out))
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		line = defaultValue
	}
	fmt.Fprintln(out, "")
	return line, nil
}

func printStepHeader(out io.Writer, title string, desc string, defaultLabel string) {
	tui.Heading(out, title)
	tui.Context(out, desc)
	if strings.TrimSpace(defaultLabel) != "" {
		tui.DefaultValue(out, defaultLabel)
	}
	tui.Divider(out)
	tui.BlankLine(out)
}

func parseList(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return []string{}
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}
