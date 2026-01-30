package cmd

import (
	"fmt"

	"quibit/internal/ai"
	"quibit/internal/input"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new portfolio project idea.",
	RunE: func(cmd *cobra.Command, args []string) error {
		in, err := input.CollectProjectInput()
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		ctx := cmd.Context()
		fmt.Fprintln(cmd.OutOrStdout(), "Connecting to Gemini...")
		client, err := ai.NewGeminiClient(ctx)
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Gemini connected.")

		g, err := ai.NewGenerator(client)
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		plan, err := g.GenerateProjectPlan(ctx, in)
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		out := cmd.OutOrStdout()
		fmt.Fprintln(out, "Title")
		fmt.Fprintln(out, plan.Title)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Description")
		fmt.Fprintln(out, plan.Description)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "App Type")
		fmt.Fprintln(out, plan.AppType)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Complexity")
		fmt.Fprintln(out, plan.Complexity)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Tech Stack")
		for _, item := range plan.TechStack {
			fmt.Fprintf(out, "- %s\n", item)
		}
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Goal")
		fmt.Fprintln(out, plan.Goal)
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Estimated Time")
		fmt.Fprintln(out, plan.EstimatedTime)
		fmt.Fprintln(out, "")

		fmt.Fprintln(out, "MVP:")
		fmt.Fprintln(out, "- Features")
		for _, item := range plan.MVP.Features {
			fmt.Fprintf(out, "  - %s\n", item)
		}
		fmt.Fprintln(out, "- User Flow")
		fmt.Fprintln(out, plan.MVP.UserFlow)
		fmt.Fprintln(out, "- Success Criteria")
		fmt.Fprintln(out, plan.MVP.SuccessCriteria)
		fmt.Fprintln(out, "")

		fmt.Fprintln(out, "Extended Ideas")
		for _, item := range plan.ExtendedIdeas {
			fmt.Fprintf(out, "- %s\n", item)
		}
		fmt.Fprintln(out, "")

		fmt.Fprintln(out, "Possible Challenges")
		for _, item := range plan.PossibleChallenges {
			fmt.Fprintf(out, "- %s\n", item)
		}
		fmt.Fprintln(out, "")

		fmt.Fprintln(out, "Next Steps")
		for _, item := range plan.NextSteps {
			fmt.Fprintf(out, "- %s\n", item)
		}
		return nil
	},
}
