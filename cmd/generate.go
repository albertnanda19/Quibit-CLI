package cmd

import (
	"errors"
	"fmt"

	"quibit/internal/ai"
	"quibit/internal/cli"
	"quibit/internal/db"
	"quibit/internal/dna"
	"quibit/internal/input"
	"quibit/internal/persistence/repository"

	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new portfolio project idea.",
	RunE: func(cmd *cobra.Command, args []string) error {
		currentInput, err := input.CollectProjectInput()
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

		ctx := cmd.Context()
		out := cmd.OutOrStdout()
		inReader := cmd.InOrStdin()

		fmt.Fprintln(out, "Connecting to Gemini...")
		client, err := ai.NewGeminiClient(ctx)
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}
		fmt.Fprintln(out, "Gemini connected.")

		pg, err := ai.NewProjectGenerator(client)
		if err != nil {
			return fmt.Errorf("generate: %w", err)
		}

	outer:
		for {
			generated, err := pg.Generate(ctx, currentInput)
			if err != nil {
				return fmt.Errorf("generate: %w", err)
			}

			cli.DisplayProject(out, generated.Project)

			for {
				action, err := cli.PromptNextAction(inReader, out)
				if err != nil {
					return fmt.Errorf("generate: %w", err)
				}

				switch action {
				case cli.NextActionAcceptAndSave:
					dnaHash := dna.HashProject(generated.Project)

					gdb, err := db.Connect(ctx)
					if err != nil {
						return fmt.Errorf("generate: %w", err)
					}
					sqlDB, err := gdb.DB()
					if err != nil {
						return fmt.Errorf("generate: get sql db: %w", err)
					}

					repo, err := repository.NewProjectRepository(gdb)
					if err != nil {
						_ = sqlDB.Close()
						return fmt.Errorf("generate: %w", err)
					}

					_, err = repo.Save(ctx, repository.SaveParams{
						Project:    generated.Project,
						DNAHash:    dnaHash,
						AIProvider: "gemini",
						RawAIJSON:  generated.RawJSON,
					})
					if err != nil {
						_ = sqlDB.Close()
						if errors.Is(err, repository.ErrDuplicateDNAHash) {
							fmt.Fprintln(out, "Duplicate project idea detected.")
							fmt.Fprintln(out, "")
							continue
						}
						return fmt.Errorf("generate: %w", err)
					}
					_ = sqlDB.Close()

					fmt.Fprintln(out, "Project saved.")
					return nil

				case cli.NextActionRegenerateSameInputs:
					fmt.Fprintln(out, "")
					continue outer

				case cli.NextActionRegenerateModifiedInputs:
					updated, err := cli.RunWizard()
					if err != nil {
						return fmt.Errorf("generate: %w", err)
					}
					currentInput = updated
					fmt.Fprintln(out, "")
					continue outer

				case cli.NextActionCancel:
					return nil

				default:
					return fmt.Errorf("generate: invalid action")
				}
			}
		}
	},
}
