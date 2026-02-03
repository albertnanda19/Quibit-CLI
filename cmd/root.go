package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"quibit/internal/config"
	"quibit/internal/db"
	"quibit/internal/persistence"
	"quibit/internal/tui"

	"github.com/spf13/cobra"
)

var migrate bool
var noAnim bool
var noSplash bool
var splashOnce sync.Once

func splashModeFromCmd(cmd *cobra.Command) string {
	if cmd == nil {
		return "idle"
	}
	name := strings.ToLower(strings.TrimSpace(cmd.Name()))
	switch name {
	case "", "quibit":
		return "idle"
	case "generate":
		return "generate"
	case "continue":
		return "continue"
	case "browse":
		return "browse"
	default:
		return "idle"
	}
}

var rootCmd = &cobra.Command{
	Use:           "quibit",
	Short:         "Quibit is a personal CLI to generate portfolio project ideas.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		tui.SetMotionEnabled(!noAnim)
		if !migrate {
			if !noSplash && !config.SplashDisabledByEnv() {
				splashOnce.Do(func() {
					mode := splashModeFromCmd(cmd)
					shown, _ := tui.ShowSplashScreen(cmd.Context(), os.Stdin, cmd.OutOrStdout(), mode)
					if shown {
						_ = config.MarkSplashSeen()
					}
				})
			}
		}
		if migrate {
			{
				ctx := cmd.Context()
				out := cmd.OutOrStdout()
				spin := tui.StartSpinner(ctx, out, "Running database migrations")
				defer spin.Stop()
				gdb, err := db.Connect(ctx)
				if err != nil {
					return fmt.Errorf("migrate: %w", err)
				}

				sqlDB, err := gdb.DB()
				if err != nil {
					return fmt.Errorf("migrate: get sql db: %w", err)
				}

				if err := persistence.AutoMigrate(ctx, gdb); err != nil {
					return fmt.Errorf("migrate: %w", err)
				}
				if err := sqlDB.Close(); err != nil {
					return fmt.Errorf("migrate: close sql db: %w", err)
				}
				spin.Stop()
				tui.Done(out, "Database migrations completed")
			}

			os.Exit(0)
			return nil
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&migrate, "migrate", false, "Run database migrations")
	rootCmd.PersistentFlags().BoolVar(&noAnim, "no-anim", false, "Disable subtle CLI animations")
	rootCmd.PersistentFlags().BoolVar(&noSplash, "no-splash", false, "Disable startup splash")
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(continueCmd)
	rootCmd.AddCommand(browseCmd)
}
