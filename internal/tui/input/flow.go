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

	appType, err := promptSelectWithCustom(in, out, reader, ApplicationTypePrompt)
	if err != nil {
		return model.ProjectInput{}, err
	}
	printDivider(out)

	// Web-specific sub-flow (must happen before Project Category, per UX requirement).
	var techStack []string
	var database []string
	if strings.TrimSpace(appType) == "web" {
		arch, err := promptSelectWithCustom(in, out, reader, WebArchitecturePrompt)
		if err != nil {
			return model.ProjectInput{}, err
		}
		printDivider(out)

		switch strings.TrimSpace(strings.ToLower(arch)) {
		case "mvc":
			mvcFramework, err := promptSelectWithCustom(in, out, reader, WebMVCFrameworkPrompt)
			if err != nil {
				return model.ProjectInput{}, err
			}
			techStack = []string{strings.TrimSpace(mvcFramework)}
			printDivider(out)

			dbRaw, err := promptSelectWithCustom(in, out, reader, DatabasePrompt)
			if err != nil {
				return model.ProjectInput{}, err
			}
			database = parseList(dbRaw)
			printDivider(out)
		case "split":
			frontend, err := promptSelectWithCustom(in, out, reader, WebFrontendFrameworkPrompt)
			if err != nil {
				return model.ProjectInput{}, err
			}
			printDivider(out)

			backend, err := promptSelectWithCustom(in, out, reader, WebBackendFrameworkPrompt)
			if err != nil {
				return model.ProjectInput{}, err
			}
			printDivider(out)

			dbRaw, err := promptSelectWithCustom(in, out, reader, DatabasePrompt)
			if err != nil {
				return model.ProjectInput{}, err
			}
			database = parseList(dbRaw)
			printDivider(out)

			techStack = []string{strings.TrimSpace(frontend), strings.TrimSpace(backend)}
		default:
			// Unknown custom architecture: fall back to existing generic tech stack prompts.
		}
	}

	// Non-web (or unknown custom architecture) path uses generic tech/db prompts.
	if len(techStack) == 0 {
		techStackRaw, err := promptSelectWithCustom(in, out, reader, TechnologyStackPrompt)
		if err != nil {
			return model.ProjectInput{}, err
		}
		techStack = parseList(techStackRaw)
		printDivider(out)

		dbRaw, err := promptSelectWithCustom(in, out, reader, DatabasePrompt)
		if err != nil {
			return model.ProjectInput{}, err
		}
		database = parseList(dbRaw)
		printDivider(out)
	}

	// Project Category comes after tech selection so AI can recommend based on chosen tech.
	projectKind, err := promptSelectOptionalWithCustom(in, out, reader, ProjectKindPrompt)
	if err != nil {
		return model.ProjectInput{}, err
	}
	printDivider(out)

	complexity, err := promptSelectWithCustom(in, out, reader, ComplexityPrompt)
	if err != nil {
		return model.ProjectInput{}, err
	}
	printDivider(out)

	goal, err := promptSelectWithCustom(in, out, reader, ProjectGoalPrompt)
	if err != nil {
		return model.ProjectInput{}, err
	}
	printDivider(out)

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
	printHeader(out, p.Title, p.Description)
	options := buildOptions(p)
	selection, err := tui.SelectOptionWithDefault(in, out, "Use arrow keys, Enter to confirm.", options, p.Default.Value)
	if err != nil {
		return "", err
	}
	if selection.ID != "custom" {
		return selection.ID, nil
	}
	return promptWithDefault(reader, out, "Custom input", p.Default.Value)
}

func promptSelectOptionalWithCustom(in *os.File, out io.Writer, reader *bufio.Reader, p SelectPrompt) (string, error) {
	printHeader(out, p.Title, p.Description)
	options := buildOptionsOptional(p)
	selection, err := tui.SelectOptionWithDefault(in, out, "Use arrow keys, Enter to confirm.", options, "skip")
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
	// Default selection should be "skip" so pressing Enter preserves current behavior.
	options := []tui.Option{
		{ID: "skip", Label: p.Default.Label},
		{ID: "custom", Label: p.CustomLabel},
	}
	for _, opt := range p.Options {
		// If prompt already included a skip entry, avoid duplicating it.
		if strings.TrimSpace(opt.Value) == "" {
			continue
		}
		options = append(options, tui.Option{ID: opt.Value, Label: opt.Label})
	}
	return options
}

func promptWithDefault(reader *bufio.Reader, out io.Writer, label string, defaultValue string) (string, error) {
	fmt.Fprintf(out, "%s [%s]:\n> ", label, defaultValue)
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

func printHeader(out io.Writer, title string, desc string) {
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, title)
	if desc != "" {
		fmt.Fprintln(out, desc)
	}
	fmt.Fprintln(out, "")
}

func printDivider(out io.Writer) {
	fmt.Fprintln(out, "----")
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
