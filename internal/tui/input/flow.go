package input

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"quibit/internal/model"
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

	var techStack []string
	var database []string
	if strings.TrimSpace(appType) == "web" {
		arch, err := promptSelectWithCustom(in, out, reader, WebArchitecturePrompt)
		if err != nil {
			return model.ProjectInput{}, err
		}
		tui.Divider(out)

		switch strings.TrimSpace(strings.ToLower(arch)) {
		case "mvc":
			mvcFramework, err := promptSelectWithCustom(in, out, reader, WebMVCFrameworkPrompt)
			if err != nil {
				return model.ProjectInput{}, err
			}
			techStack = []string{strings.TrimSpace(mvcFramework)}
			tui.Divider(out)

			dbRaw, err := promptSelectWithCustom(in, out, reader, DatabasePrompt)
			if err != nil {
				return model.ProjectInput{}, err
			}
			database = parseList(dbRaw)
			tui.Divider(out)
		case "split":
			frontend, err := promptSelectWithCustom(in, out, reader, WebFrontendFrameworkPrompt)
			if err != nil {
				return model.ProjectInput{}, err
			}
			tui.Divider(out)

			backend, err := promptSelectWithCustom(in, out, reader, WebBackendFrameworkPrompt)
			if err != nil {
				return model.ProjectInput{}, err
			}
			tui.Divider(out)

			dbRaw, err := promptSelectWithCustom(in, out, reader, DatabasePrompt)
			if err != nil {
				return model.ProjectInput{}, err
			}
			database = parseList(dbRaw)
			tui.Divider(out)

			techStack = []string{strings.TrimSpace(frontend), strings.TrimSpace(backend)}
		default:
		}
	}

	if len(techStack) == 0 {
		techStackRaw, err := promptSelectWithCustom(in, out, reader, TechnologyStackPrompt)
		if err != nil {
			return model.ProjectInput{}, err
		}
		techStack = parseList(techStackRaw)
		tui.Divider(out)

		dbRaw, err := promptSelectWithCustom(in, out, reader, DatabasePrompt)
		if err != nil {
			return model.ProjectInput{}, err
		}
		database = parseList(dbRaw)
		tui.Divider(out)
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
		AppType:     appType,
		ProjectKind: projectKind,
		Complexity:  complexity,
		TechStack:   techStack,
		Database:    database,
		Goal:        goal,
		Timeframe:   timeframe,
	}, nil
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
