package cmd

import (
	"fmt"

	"quibit/internal/engine"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new portfolio project idea.",
	RunE: func(cmd *cobra.Command, args []string) error {
		g := engine.NewGenerator()
		out := g.Generate()
		fmt.Fprintln(cmd.OutOrStdout(), out)
		return nil
	},
}
