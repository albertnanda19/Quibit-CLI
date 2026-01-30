package cmd

import (
	"fmt"

	"quibit/internal/ai"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new portfolio project idea.",
	RunE: func(cmd *cobra.Command, args []string) error {
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
		idea, err := g.GenerateProjectIdea(ctx)
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), "Title:")
		fmt.Fprintln(cmd.OutOrStdout(), idea.Title)
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "Description:")
		fmt.Fprintln(cmd.OutOrStdout(), idea.Description)
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "Complexity:")
		fmt.Fprintln(cmd.OutOrStdout(), idea.Complexity)
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "Tech Stack:")
		for _, item := range idea.TechStack {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", item)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "Core Features:")
		for _, item := range idea.CoreFeatures {
			fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", item)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "")
		fmt.Fprintln(cmd.OutOrStdout(), "Twist:")
		fmt.Fprintln(cmd.OutOrStdout(), idea.Twist)
		return nil
	},
}
