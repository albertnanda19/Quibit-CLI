package input

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"quibit/internal/model"
)

func promptString(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func promptWithDefault(r *bufio.Reader, label string, defaultValue string) (string, error) {
	fmt.Printf("%s [%s]:\n> ", label, defaultValue)
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
	fmt.Println(p.Title)
	fmt.Println(p.Description)
	fmt.Println("")
	fmt.Println("Pilihan:")
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
			return "", fmt.Errorf("Error: invalid selection")
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
			fmt.Printf("%s\n\n", err.Error())
			continue
		}

		if strings.TrimSpace(line) == "0" {
			fmt.Print("\nCustom Input:\n> ")
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
				fmt.Printf("%s\n\n", err.Error())
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
			fmt.Printf("%s\n\n", err.Error())
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
	fmt.Println("---")
	fmt.Println("")

	projectKind, err := promptSelectWithDefault(r, ProjectKindPrompt, nil)
	if err != nil {
		return model.ProjectInput{}, err
	}
	fmt.Println("---")
	fmt.Println("")

	complexity, err := promptSelectWithDefault(r, ComplexityPrompt, func(v string) error {
		v = normalizeComplexity(v)
		return validateComplexity(v)
	})
	if err != nil {
		return model.ProjectInput{}, err
	}
	complexity = normalizeComplexity(complexity)
	fmt.Println("---")
	fmt.Println("")

	techStackRaw, err := promptSelectWithDefault(r, TechnologyStackPrompt, nil)
	if err != nil {
		return model.ProjectInput{}, err
	}
	techStack := parseTechStack(techStackRaw)
	fmt.Println("---")
	fmt.Println("")

	dbPref, err := promptSelectWithDefault(r, DatabasePrompt, nil)
	if err != nil {
		return model.ProjectInput{}, err
	}
	fmt.Println("---")
	fmt.Println("")

	goal, err := promptSelectWithDefault(r, ProjectGoalPrompt, nil)
	if err != nil {
		return model.ProjectInput{}, err
	}
	fmt.Println("---")
	fmt.Println("")

	timeframe, err := promptSelectWithDefault(r, EstimatedTimeframePrompt, nil)
	if err != nil {
		return model.ProjectInput{}, err
	}

	return model.ProjectInput{
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
