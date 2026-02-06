package input

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"quibit/internal/model"
	"quibit/internal/techstack"
)

func promptString(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptWithDefault(r *bufio.Reader, label string, defaultValue string) (string, error) {
	fmt.Println("")
	fmt.Println(label)
	if strings.TrimSpace(defaultValue) != "" {
		fmt.Printf("Default: %s\n", defaultValue)
	}
	fmt.Println("Input:")
	fmt.Print("> ")
	v, err := promptString(r)
	if err != nil {
		return "", err
	}
	if v == "" {
		v = defaultValue
	}
	fmt.Print("\n")
	return v, nil
}

type validatorFn func(string) error

func renderSelectPrompt(p SelectPrompt) {
	fmt.Println("")
	fmt.Println(p.Title)
	if strings.TrimSpace(p.Description) != "" {
		fmt.Println(p.Description)
	}
	fmt.Println("")
	fmt.Println("Options:")
	for i, opt := range p.Options {
		fmt.Printf("%d) %s\n", i+1, opt.Label)
	}
	fmt.Printf("0) %s\n", p.CustomLabel)
	fmt.Println("")
	fmt.Printf("Default: %s\n", p.Default.Label)
	fmt.Println("")
	fmt.Println("Input:")
	fmt.Print("> ")
}

func parseSelectInput(line string, p SelectPrompt) (string, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return p.Default.Value, nil
	}

	if n, err := strconv.Atoi(line); err == nil {
		if n == 0 {
			return "", nil
		}
		if n < 0 || n > len(p.Options) {
			return "", fmt.Errorf("Invalid selection. Choose 1-%d, 0 for custom, or press Enter for default.", len(p.Options))
		}
		return p.Options[n-1].Value, nil
	}

	return line, nil
}

func promptSelectWithDefault(r *bufio.Reader, p SelectPrompt, validate validatorFn) (string, error) {
	for {
		renderSelectPrompt(p)
		line, err := promptString(r)
		if err != nil {
			return "", err
		}

		v, err := parseSelectInput(line, p)
		if err != nil {
			fmt.Printf("Error: %s\n\n", strings.TrimSpace(err.Error()))
			continue
		}

		if strings.TrimSpace(line) == "0" {
			fmt.Print("\nCustom input\nInput:\n> ")
			custom, err := promptString(r)
			if err != nil {
				return "", err
			}
			custom = strings.TrimSpace(custom)
			if custom == "" {
				v = p.Default.Value
			} else {
				v = custom
			}
		}

		if validate != nil {
			if err := validate(v); err != nil {
				fmt.Printf("Error: %s\n\n", strings.TrimSpace(err.Error()))
				continue
			}
		}

		fmt.Print("\n")
		return v, nil
	}
}

func promptWithValidation(r *bufio.Reader, label string, defaultValue string, validate validatorFn) (string, error) {
	for {
		v, err := promptWithDefault(r, label, defaultValue)
		if err != nil {
			return "", err
		}
		if err := validate(v); err != nil {
			fmt.Printf("Error: %s\n\n", strings.TrimSpace(err.Error()))
			continue
		}
		return v, nil
	}
}

func CollectProjectInput() (model.ProjectInput, error) {
	r := bufio.NewReader(os.Stdin)

	appType, err := promptSelectWithDefault(r, ApplicationTypePrompt, nil)
	if err != nil {
		return model.ProjectInput{}, err
	}
	fmt.Println(strings.Repeat("-", 42))

	complexity, err := promptSelectWithDefault(r, ComplexityPrompt, func(v string) error {
		v = normalizeComplexity(v)
		return validateComplexity(v)
	})
	if err != nil {
		return model.ProjectInput{}, err
	}
	complexity = normalizeComplexity(complexity)
	fmt.Println(strings.Repeat("-", 42))

	var techStack []string
	var dbPref []string
	if strings.TrimSpace(appType) == "web" {
		arch, err := promptSelectWithDefault(r, WebArchitecturePrompt, nil)
		if err != nil {
			return model.ProjectInput{}, err
		}
		fmt.Println(strings.Repeat("-", 42))

		switch strings.TrimSpace(strings.ToLower(arch)) {
		case "mvc":
			mvcFramework, err := promptSelectWithDefault(r, WebMVCFrameworkPrompt, nil)
			if err != nil {
				return model.ProjectInput{}, err
			}
			techStack = []string{strings.TrimSpace(mvcFramework)}
			fmt.Println(strings.Repeat("-", 42))
		case "split":
			frontend, err := promptSelectWithDefault(r, WebFrontendFrameworkPrompt, nil)
			if err != nil {
				return model.ProjectInput{}, err
			}
			fmt.Println(strings.Repeat("-", 42))

			backend, err := promptSelectWithDefault(r, WebBackendFrameworkPrompt, nil)
			if err != nil {
				return model.ProjectInput{}, err
			}
			fmt.Println(strings.Repeat("-", 42))

			techStack = []string{strings.TrimSpace(frontend), strings.TrimSpace(backend)}
		default:
		}
	}

	if len(techStack) == 0 {
		languageID, err := promptSelectWithDefault(r, ProgrammingLanguagePrompt(), nil)
		if err != nil {
			return model.ProjectInput{}, err
		}
		fmt.Println(strings.Repeat("-", 42))

		var frameworkRaw string
		if techstack.LanguageExists(strings.TrimSpace(languageID)) {
			frameworkRaw, err = promptSelectWithDefault(r, FrameworkPrompt(strings.TrimSpace(languageID)), nil)
			if err != nil {
				return model.ProjectInput{}, err
			}
		} else {
			frameworkRaw, err = promptWithDefault(r, "Framework / Library (Custom / Manual)", "")
			if err != nil {
				return model.ProjectInput{}, err
			}
		}
		frameworks := parseTechStack(frameworkRaw)
		techStack = append([]string{strings.TrimSpace(languageID)}, frameworks...)
		fmt.Println(strings.Repeat("-", 42))
	}

	dbRaw, err := promptSelectWithDefault(r, DatabasePrompt, nil)
	if err != nil {
		return model.ProjectInput{}, err
	}
	dbPref = parseTechStack(dbRaw)
	fmt.Println(strings.Repeat("-", 42))

	projectKind, err := promptSelectWithDefault(r, ProjectKindPrompt, nil)
	if err != nil {
		return model.ProjectInput{}, err
	}
	fmt.Println(strings.Repeat("-", 42))

	goal, err := promptSelectWithDefault(r, ProjectGoalPrompt, nil)
	if err != nil {
		return model.ProjectInput{}, err
	}
	fmt.Println(strings.Repeat("-", 42))

	timeframe, err := promptSelectWithDefault(r, EstimatedTimeframePrompt, nil)
	if err != nil {
		return model.ProjectInput{}, err
	}

	return model.ProjectInput{
		UserIdea:    "",
		AppType:     appType,
		ProjectKind: projectKind,
		Complexity:  complexity,
		TechStack:   techStack,
		Database:    dbPref,
		Goal:        goal,
		Timeframe:   timeframe,
	}, nil
}

func parseTechStack(v string) []string {
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
