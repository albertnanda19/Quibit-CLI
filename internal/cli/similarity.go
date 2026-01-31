package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type SimilarityAction int

const (
	SimilarityActionAutoPivot SimilarityAction = iota + 1
	SimilarityActionModifyInputs
	SimilarityActionCancel
)

func PromptSimilarityResolution(in io.Reader, out io.Writer, similarityScore float64) (SimilarityAction, error) {
	r := bufio.NewReader(in)
	for {
		fmt.Fprintf(out, "Your project idea is very similar to an existing one (similarity: %.2f).\n", similarityScore)
		fmt.Fprintln(out, "Quibit will regenerate a more distinctive variant automatically.")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "1) Auto-regenerate with smart pivot")
		fmt.Fprintln(out, "2) Modify inputs manually")
		fmt.Fprintln(out, "3) Cancel")
		fmt.Fprintln(out, "")
		fmt.Fprint(out, "Select an option [1]: ")

		line, err := r.ReadString('\n')
		if err != nil {
			return 0, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			return SimilarityActionAutoPivot, nil
		}

		switch line {
		case "1":
			return SimilarityActionAutoPivot, nil
		case "2":
			return SimilarityActionModifyInputs, nil
		case "3":
			return SimilarityActionCancel, nil
		default:
			fmt.Fprintln(out, "Invalid option. Please select 1-3.")
			fmt.Fprintln(out, "")
		}
	}
}
