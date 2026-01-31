package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"quibit/internal/domain"
)

type NextAction int

const (
	NextActionAcceptAndSave NextAction = iota + 1
	NextActionRegenerateSameInputs
	NextActionRegenerateModifiedInputs
	NextActionCancel
)

func DisplayProject(out io.Writer, p domain.Project) {
	fmt.Fprintln(out, "Title")
	fmt.Fprintln(out, p.Title)
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Summary")
	fmt.Fprintln(out, p.Summary)
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Problem Statement")
	fmt.Fprintln(out, p.ProblemStatement)
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Target Users")
	for _, item := range p.TargetUsers {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Core Features")
	for _, item := range p.CoreFeatures {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "MVP Scope")
	for _, item := range p.MVPScope {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Optional Extensions")
	for _, item := range p.OptionalExtensions {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		fmt.Fprintf(out, "- %s\n", item)
	}
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Recommended Stack")
	fmt.Fprintln(out, p.RecommendedStack)
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Estimated Complexity")
	fmt.Fprintln(out, p.EstimatedComplexity)
	fmt.Fprintln(out, "")

	fmt.Fprintln(out, "Estimated Duration")
	fmt.Fprintln(out, p.EstimatedDuration)
	fmt.Fprintln(out, "")
}

func PromptNextAction(in io.Reader, out io.Writer) (NextAction, error) {
	r := bufio.NewReader(in)
	for {
		fmt.Fprintln(out, "Review Result")
		fmt.Fprintln(out, "What would you like to do?")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "1) Accept and save project")
		fmt.Fprintln(out, "2) Regenerate with same inputs")
		fmt.Fprintln(out, "3) Regenerate with modified inputs")
		fmt.Fprintln(out, "4) Cancel")
		fmt.Fprintln(out, "")
		fmt.Fprint(out, "Select an option [1]: ")

		line, err := r.ReadString('\n')
		if err != nil {
			return 0, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return NextActionAcceptAndSave, nil
		}

		switch line {
		case "1":
			return NextActionAcceptAndSave, nil
		case "2":
			return NextActionRegenerateSameInputs, nil
		case "3":
			return NextActionRegenerateModifiedInputs, nil
		case "4":
			return NextActionCancel, nil
		default:
			fmt.Fprintln(out, "Invalid option. Please select 1-4.")
			fmt.Fprintln(out, "")
		}
	}
}
