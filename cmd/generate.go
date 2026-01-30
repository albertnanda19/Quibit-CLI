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

		out, err := g.GenerateText(ctx, "Generate one simple software project idea.")
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), out)
		return nil
	},
}
